package core_test

import (
	"strings"
	"testing"

	"github.com/Notailab/go-agent/agent/core"
	"github.com/Notailab/go-agent/agent/storage"
)

func TestMemoryReplaceRebuildsDialogIndex(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "replace rebuilds dialog index",
			run: func(t *testing.T) {
				memory := core.NewMemory(
					storage.NewInMemoryChatStore(),
					storage.NewInMemoryLongStore(),
				)
				if err := memory.AddChat(core.RoleUser, "user 1"); err != nil {
					t.Fatalf("add chat failed: %v", err)
				}
				if err := memory.AddChat(core.RoleAssistant, "assistant 1"); err != nil {
					t.Fatalf("add chat failed: %v", err)
				}
				if err := memory.AddChat(core.RoleUser, "user 2"); err != nil {
					t.Fatalf("add chat failed: %v", err)
				}

				if err := memory.ReplaceChat(0, 2, []core.ChatMessage{{Role: core.RoleUser, Content: "summary"}}); err != nil {
					t.Fatalf("replace chat failed: %v", err)
				}

				got := memory.ChatMemory()
				if len(got) != 2 {
					t.Fatalf("unexpected chat length: %d", len(got))
				}
				if got[0].Content != "summary" || got[1].Content != "user 2" {
					t.Fatalf("unexpected chat memory: %#v", got)
				}
			},
		},
		{
			name: "add chat records role and content",
			run: func(t *testing.T) {
				memory := core.NewMemory(
					storage.NewInMemoryChatStore(),
					storage.NewInMemoryLongStore(),
				)
				if err := memory.AddChat(core.RoleUser, "hello"); err != nil {
					t.Fatalf("add chat failed: %v", err)
				}

				got := memory.ChatMemory()
				if len(got) != 1 {
					t.Fatalf("unexpected chat length: %d", len(got))
				}
				if got[0].Role != core.RoleUser || got[0].Content != "hello" {
					t.Fatalf("unexpected chat message: %#v", got[0])
				}
			},
		},
		{
			name: "add tool call copies input",
			run: func(t *testing.T) {
				memory := core.NewMemory(
					storage.NewInMemoryChatStore(),
					storage.NewInMemoryLongStore(),
				)
				toolCalls := []core.ToolCall{{Id: "1", Type: "function"}}
				if err := memory.AddToolCall(toolCalls); err != nil {
					t.Fatalf("add tool call failed: %v", err)
				}
				toolCalls[0].Id = "changed"

				got := memory.ChatMemory()
				if len(got) != 1 {
					t.Fatalf("unexpected chat length: %d", len(got))
				}
				if got[0].Role != core.RoleAssistant || len(got[0].ToolCalls) != 1 || got[0].ToolCalls[0].Id != "1" {
					t.Fatalf("unexpected tool call message: %#v", got[0])
				}
			},
		},
		{
			name: "add tool result appends tool message",
			run: func(t *testing.T) {
				memory := core.NewMemory(
					storage.NewInMemoryChatStore(),
					storage.NewInMemoryLongStore(),
				)
				if err := memory.AddToolResult("call-1", "result"); err != nil {
					t.Fatalf("add tool result failed: %v", err)
				}

				got := memory.ChatMemory()
				if len(got) != 1 {
					t.Fatalf("unexpected chat length: %d", len(got))
				}
				if got[0].Role != core.RoleTool || got[0].ToolCallID != "call-1" || got[0].Content != "result" {
					t.Fatalf("unexpected tool result message: %#v", got[0])
				}
			},
		},
		{
			name: "add tool call rejects empty slice",
			run: func(t *testing.T) {
				memory := core.NewMemory(
					storage.NewInMemoryChatStore(),
					storage.NewInMemoryLongStore(),
				)
				if err := memory.AddToolCall(nil); err == nil || !strings.Contains(err.Error(), "tool calls is empty") {
					t.Fatalf("expected empty tool calls error, got: %v", err)
				}
			},
		},
		{
			name: "long memory formatting keeps indexes",
			run: func(t *testing.T) {
				memory := core.NewMemory(
					storage.NewInMemoryChatStore(),
					storage.NewInMemoryLongStore(),
				)
				if err := memory.OperateLongMemory(core.LongMemoryCreate, 0, "entry1"); err != nil {
					t.Fatalf("create long memory failed: %v", err)
				}
				if err := memory.OperateLongMemory(core.LongMemoryCreate, 0, ""); err != nil {
					t.Fatalf("create empty long memory failed: %v", err)
				}
				if err := memory.OperateLongMemory(core.LongMemoryCreate, 0, "entry3"); err != nil {
					t.Fatalf("create long memory failed: %v", err)
				}

				got := memory.LongMemory()
				if !strings.Contains(got, "Long memory:") || !strings.Contains(got, "0 entry1") || !strings.Contains(got, "2 entry3") {
					t.Fatalf("unexpected long memory: %q", got)
				}
				if strings.Contains(got, "1 ") {
					t.Fatalf("empty entry should be omitted: %q", got)
				}
			},
		},
		{
			name: "nil memory errors",
			run: func(t *testing.T) {
				var memory *core.Memory
				if err := memory.ReplaceChat(0, 1, nil); err == nil || !strings.Contains(err.Error(), "memory is nil") {
					t.Fatalf("expected nil memory error, got: %v", err)
				}
				if err := memory.AddChat(core.RoleUser, "hello"); err == nil || !strings.Contains(err.Error(), "memory is nil") {
					t.Fatalf("expected nil memory error, got: %v", err)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}
