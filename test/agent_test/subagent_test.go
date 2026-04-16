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

	tests := []struct {
		name    string
		run     func(t *testing.T)
	}{
		{
			name: "uses custom prompt and no tools",
			run: func(t *testing.T) {
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
				if len(captured.Messages) != 2 {
					t.Fatalf("unexpected message count: %d", len(captured.Messages))
				}
				if captured.Messages[0].Role != core.RoleSystem {
					t.Fatalf("unexpected first message role: %s", captured.Messages[0].Role)
				}
				if !strings.Contains(captured.Messages[0].Content, "summarize messages subagent") {
					t.Fatalf("custom subagent prompt was not used: %q", captured.Messages[0].Content)
				}
				if captured.Messages[1].Role != core.RoleUser {
					t.Fatalf("unexpected second message role: %s", captured.Messages[1].Role)
				}
				if !strings.Contains(captured.Messages[1].Content, "Please summarize this conversation.") {
					t.Fatalf("unexpected user message: %q", captured.Messages[1].Content)
				}
				if len(captured.Tools) != 0 {
					t.Fatalf("subagent should not send tools: %#v", captured.Tools)
				}
			},
		},
		{
			name: "nil subagent returns error",
			run: func(t *testing.T) {
				var sub *agent.SubAgent
				_, err := sub.Run(context.Background(), nil)
				if err == nil || !strings.Contains(err.Error(), "subagent is nil") {
					t.Fatalf("expected nil subagent error, got: %v", err)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}
