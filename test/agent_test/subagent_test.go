package agent_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Notailab/go-agent/agent/agent"
	"github.com/Notailab/go-agent/agent/core"
	"github.com/Notailab/go-agent/agent/storage"
)

func TestSummarizeMessageSubAgentUsesCustomPrompt(t *testing.T) {
	t.Parallel()

	var captured core.ChatRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if err := json.NewDecoder(r.Body).Decode(&captured); err != nil {
			t.Fatalf("decode request failed: %v", err)
		}

		_ = json.NewEncoder(w).Encode(core.LLMResult{
			Choices: []core.LLMChoice{{Message: core.LLMMessage{Content: "summary"}}},
		})
	}))
	defer server.Close()

	react := agent.NewReactAgent(
		agent.WithLLM(server.URL, "test-model", "test-key"),
		agent.WithMemory(core.NewMemory(
			storage.NewInMemoryChatStore(),
			storage.NewInMemoryLongStore(),
		)),
	)

	sub := agent.SummarizeMessageSubAgent(react)
	output, err := sub.Run(context.Background(), []core.ChatMessage{{Role: core.RoleUser, Content: "Please summarize this conversation."}})
	if err != nil {
		t.Fatalf("subagent run failed: %v", err)
	}
	if output != "summary" {
		t.Fatalf("unexpected subagent output: %q", output)
	}
	if len(captured.Messages) == 0 {
		t.Fatal("expected messages to be sent")
	}
	if captured.Messages[0].Role != core.RoleSystem {
		t.Fatalf("unexpected first message role: %s", captured.Messages[0].Role)
	}
	if !strings.Contains(captured.Messages[0].Content, "summarize messages subagent") {
		t.Fatalf("custom subagent prompt was not used: %q", captured.Messages[0].Content)
	}
}
