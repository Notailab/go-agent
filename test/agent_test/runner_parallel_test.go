package agent_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Notailab/go-agent/agent/agent"
	"github.com/Notailab/go-agent/agent/core"
	"github.com/Notailab/go-agent/agent/storage"
)

type blockingTool struct {
	name    string
	started chan<- string
	release <-chan struct{}
	result  string
	delay   time.Duration
}

func (t *blockingTool) Name() string {
	return t.name
}

func (t *blockingTool) Description() string {
	return "blocking test tool"
}

func (t *blockingTool) Parameters() core.Parameters {
	return core.Parameters{Type: "object"}
}

func (t *blockingTool) Execute(_ string) (string, error) {
	t.started <- t.name
	<-t.release
	time.Sleep(t.delay)
	return t.result, nil
}

func TestReactAgentRunsToolsInParallelAndKeepsOrder(t *testing.T) {
	t.Parallel()

	var (
		requestMu sync.Mutex
		requests  []core.ChatRequest
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}

		var req core.ChatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request failed: %v", err)
		}

		requestMu.Lock()
		requests = append(requests, req)
		requestIndex := len(requests) - 1
		requestMu.Unlock()

		callA := core.ToolCall{Id: "call-a", Type: "function"}
		callA.Function.Name = "slow_a"
		callA.Function.Arguments = `{"anything":true}`

		callB := core.ToolCall{Id: "call-b", Type: "function"}
		callB.Function.Name = "fast_b"
		callB.Function.Arguments = `{"anything":true}`

		responses := []core.LLMResult{
			{
				Choices: []core.LLMChoice{{Message: core.LLMMessage{ToolCalls: []core.ToolCall{callA, callB}}}},
			},
			{
				Choices: []core.LLMChoice{{Message: core.LLMMessage{Content: "done"}}},
			},
		}

		if requestIndex >= len(responses) {
			t.Fatalf("unexpected request count: %d", requestIndex+1)
		}

		_ = json.NewEncoder(w).Encode(responses[requestIndex])
	}))
	defer server.Close()

	started := make(chan string, 2)
	release := make(chan struct{})

	memory := core.NewMemory(
		storage.NewInMemoryChatStore(),
		storage.NewInMemoryLongStore(),
	)

	react := agent.NewReactAgent(
		agent.WithLLM(server.URL, "test-model", "test-key"),
		agent.WithMemory(memory),
		agent.WithTools(
			&blockingTool{name: "slow_a", started: started, release: release, result: "result-a", delay: 120 * time.Millisecond},
			&blockingTool{name: "fast_b", started: started, release: release, result: "result-b", delay: 10 * time.Millisecond},
		),
	)

	resultCh := make(chan struct {
		output string
		err    error
	}, 1)

	go func() {
		output, err := react.Run(context.Background(), "hello")
		resultCh <- struct {
			output string
			err    error
		}{output: output, err: err}
	}()

	startedNames := make(map[string]bool, 2)
	waitForStart := func() {
		t.Helper()
		select {
		case got := <-started:
			startedNames[got] = true
			if len(startedNames) == 2 {
				return
			}
		case <-time.After(1 * time.Second):
			t.Fatal("timed out waiting for tool starts")
		}
	}

	waitForStart()
	waitForStart()

	if !startedNames["slow_a"] || !startedNames["fast_b"] {
		t.Fatalf("unexpected started tools: %#v", startedNames)
	}

	close(release)

	select {
	case result := <-resultCh:
		if result.err != nil {
			t.Fatalf("run failed: %v", result.err)
		}
		if result.output != "done" {
			t.Fatalf("unexpected output: %q", result.output)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for run to finish")
	}

	got := memory.ChatMemory()
	if len(got) != 5 {
		t.Fatalf("unexpected chat length: %d", len(got))
	}
	if got[1].Role != core.RoleAssistant || len(got[1].ToolCalls) != 2 {
		t.Fatalf("unexpected tool call message: %#v", got[1])
	}
	if got[2].Role != core.RoleTool || got[2].ToolCallID != "call-a" || got[2].Content != "result-a" {
		t.Fatalf("unexpected first tool result: %#v", got[2])
	}
	if got[3].Role != core.RoleTool || got[3].ToolCallID != "call-b" || got[3].Content != "result-b" {
		t.Fatalf("unexpected second tool result: %#v", got[3])
	}
	if got[4].Role != core.RoleAssistant || got[4].Content != "done" {
		t.Fatalf("unexpected final assistant message: %#v", got[4])
	}

	requestMu.Lock()
	defer requestMu.Unlock()
	if len(requests) != 2 {
		t.Fatalf("unexpected request count: %d", len(requests))
	}
	if !strings.Contains(requests[1].Messages[len(requests[1].Messages)-2].Content, "result-a") {
		t.Fatalf("second request did not include first tool result: %#v", requests[1].Messages)
	}
	if !strings.Contains(requests[1].Messages[len(requests[1].Messages)-1].Content, "result-b") {
		t.Fatalf("second request did not include second tool result: %#v", requests[1].Messages)
	}
}
