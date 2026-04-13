package agent_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/Notailab/go-agent/agent/agent"
	"github.com/Notailab/go-agent/agent/core"
	"github.com/Notailab/go-agent/agent/storage"
	"github.com/Notailab/go-agent/agent/tools"
)

func TestReactAgentRunUsesLLMAndMemory(t *testing.T) {
	t.Parallel()

	var (
		mu       sync.Mutex
		requests []core.ChatRequest
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/chat/completions" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}

		var req core.ChatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request failed: %v", err)
		}

		mu.Lock()
		requests = append(requests, req)
		idx := len(requests) - 1
		mu.Unlock()

		responses := []string{"pong-1", "pong-2"}
		content := responses[idx]
		_ = json.NewEncoder(w).Encode(core.LLMResult{
			Choices: []core.LLMChoice{{Message: core.LLMMessage{Content: content}}},
		})
	}))
	defer server.Close()

	react := agent.NewReactAgent(
		agent.WithLLM(server.URL, "test-model", "test-key"),
		agent.WithMemory(core.NewMemory(
			storage.NewInMemoryChatStore(),
			storage.NewInMemoryLongStore(),
		)),
		agent.WithTools(&tools.BashTool{}),
	)

	first, err := react.Run(context.Background(), "hello")
	if err != nil {
		t.Fatalf("first run failed: %v", err)
	}
	if first != "pong-1" {
		t.Fatalf("unexpected first output: %q", first)
	}

	second, err := react.Run(context.Background(), "again")
	if err != nil {
		t.Fatalf("second run failed: %v", err)
	}
	if second != "pong-2" {
		t.Fatalf("unexpected second output: %q", second)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(requests) != 2 {
		t.Fatalf("unexpected request count: %d", len(requests))
	}
	if len(requests[0].Messages) != 3 {
		t.Fatalf("unexpected first message count: %d", len(requests[0].Messages))
	}
	if requests[0].Messages[0].Role != core.RoleSystem {
		t.Fatalf("unexpected first message role: %s", requests[0].Messages[0].Role)
	}
	if requests[0].Messages[2].Content != "hello" {
		t.Fatalf("unexpected first user message: %q", requests[0].Messages[2].Content)
	}

	if len(requests[1].Messages) != 5 {
		t.Fatalf("unexpected second message count: %d", len(requests[1].Messages))
	}
	foundAssistant := false
	for _, msg := range requests[1].Messages {
		if msg.Role == core.RoleAssistant && msg.Content == "pong-1" {
			foundAssistant = true
			break
		}
	}
	if !foundAssistant {
		t.Fatalf("second request did not include prior assistant reply: %#v", requests[1].Messages)
	}
}
