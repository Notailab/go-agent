package tool_test

import (
	"strings"
	"testing"

	"github.com/Notailab/go-agent/agent/core"
	"github.com/Notailab/go-agent/agent/storage"
	"github.com/Notailab/go-agent/agent/tools"
)

func TestLongMemoryTool_CreateUpdateDelete(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() (*core.Memory, string)
		params  func() string
		want    string
		wantErr string
		check   func(t *testing.T, memory *core.Memory, output string)
	}{
		{
			name: "create entry",
			setup: func() (*core.Memory, string) {
				memory := core.NewMemory(storage.NewInMemoryChatStore(), storage.NewInMemoryLongStore())
				return memory, ""
			},
			params: func() string {
				return `{"operation":"create","index":0,"context":"entry1"}`
			},
			want: "Long-term memory create operation successful",
			check: func(t *testing.T, memory *core.Memory, output string) {
				if !strings.Contains(memory.LongMemory(), "entry1") {
					t.Fatalf("memory missing entry1: %q", memory.LongMemory())
				}
			},
		},
		{
			name: "create requires context",
			setup: func() (*core.Memory, string) {
				memory := core.NewMemory(storage.NewInMemoryChatStore(), storage.NewInMemoryLongStore())
				return memory, ""
			},
			params: func() string {
				return `{"operation":"create","index":0,"context":""}`
			},
			wantErr: "context is required for create",
		},
		{
			name: "update entry",
			setup: func() (*core.Memory, string) {
				memory := core.NewMemory(storage.NewInMemoryChatStore(), storage.NewInMemoryLongStore())
				tool := tools.NewLongMemoryTool(memory)
				if _, err := tool.Execute(`{"operation":"create","index":0,"context":"entry1"}`); err != nil {
					return nil, err.Error()
				}
				return memory, ""
			},
			params: func() string {
				return `{"operation":"update","index":0,"context":"entry1-updated"}`
			},
			want: "Long-term memory update operation successful",
			check: func(t *testing.T, memory *core.Memory, output string) {
				if !strings.Contains(memory.LongMemory(), "entry1-updated") {
					t.Fatalf("memory not updated: %q", memory.LongMemory())
				}
			},
		},
		{
			name: "update requires context",
			setup: func() (*core.Memory, string) {
				memory := core.NewMemory(storage.NewInMemoryChatStore(), storage.NewInMemoryLongStore())
				return memory, ""
			},
			params: func() string {
				return `{"operation":"update","index":0,"context":""}`
			},
			wantErr: "context is required for update",
		},
		{
			name: "update requires non-negative index",
			setup: func() (*core.Memory, string) {
				memory := core.NewMemory(storage.NewInMemoryChatStore(), storage.NewInMemoryLongStore())
				return memory, ""
			},
			params: func() string {
				return `{"operation":"update","index":-1,"context":"entry1-updated"}`
			},
			wantErr: "index must be non-negative for update",
		},
		{
			name: "delete entry",
			setup: func() (*core.Memory, string) {
				memory := core.NewMemory(storage.NewInMemoryChatStore(), storage.NewInMemoryLongStore())
				tool := tools.NewLongMemoryTool(memory)
				if _, err := tool.Execute(`{"operation":"create","index":0,"context":"entry1"}`); err != nil {
					return nil, err.Error()
				}
				return memory, ""
			},
			params: func() string {
				return `{"operation":"delete","index":0,"context":"should-be-ignored"}`
			},
			want: "Long-term memory delete operation successful",
			check: func(t *testing.T, memory *core.Memory, output string) {
				if strings.Contains(memory.LongMemory(), "entry1") {
					t.Fatalf("memory not deleted: %q", memory.LongMemory())
				}
			},
		},
		{
			name: "delete requires non-negative index",
			setup: func() (*core.Memory, string) {
				memory := core.NewMemory(storage.NewInMemoryChatStore(), storage.NewInMemoryLongStore())
				return memory, ""
			},
			params: func() string {
				return `{"operation":"delete","index":-1,"context":""}`
			},
			wantErr: "index must be non-negative for delete",
		},
		{
			name: "unsupported operation",
			setup: func() (*core.Memory, string) {
				memory := core.NewMemory(storage.NewInMemoryChatStore(), storage.NewInMemoryLongStore())
				return memory, ""
			},
			params: func() string {
				return `{"operation":"move","index":0,"context":"entry"}`
			},
			wantErr: "unsupported operation: move",
		},
		{
			name: "missing operation",
			setup: func() (*core.Memory, string) {
				memory := core.NewMemory(storage.NewInMemoryChatStore(), storage.NewInMemoryLongStore())
				return memory, ""
			},
			params: func() string {
				return `{"index":0,"context":"entry"}`
			},
			wantErr: "missing required parameters: [operation]",
		},
		{
			name: "invalid operation type",
			setup: func() (*core.Memory, string) {
				memory := core.NewMemory(storage.NewInMemoryChatStore(), storage.NewInMemoryLongStore())
				return memory, ""
			},
			params: func() string {
				return `{"operation":123,"index":0,"context":"entry"}`
			},
			wantErr: `invalid type for parameter "operation"`,
		},
		{
			name: "invalid index type",
			setup: func() (*core.Memory, string) {
				memory := core.NewMemory(storage.NewInMemoryChatStore(), storage.NewInMemoryLongStore())
				return memory, ""
			},
			params: func() string {
				return `{"operation":"create","index":"zero","context":"entry"}`
			},
			wantErr: `invalid type for parameter "index"`,
		},
		{
			name: "invalid context type",
			setup: func() (*core.Memory, string) {
				memory := core.NewMemory(storage.NewInMemoryChatStore(), storage.NewInMemoryLongStore())
				return memory, ""
			},
			params: func() string {
				return `{"operation":"create","index":0,"context":123}`
			},
			wantErr: `invalid type for parameter "context"`,
		},
		{
			name: "empty params",
			setup: func() (*core.Memory, string) {
				memory := core.NewMemory(storage.NewInMemoryChatStore(), storage.NewInMemoryLongStore())
				return memory, ""
			},
			params: func() string {
				return ""
			},
			wantErr: "parameters JSON is empty",
		},
		{
			name: "invalid json",
			setup: func() (*core.Memory, string) {
				memory := core.NewMemory(storage.NewInMemoryChatStore(), storage.NewInMemoryLongStore())
				return memory, ""
			},
			params: func() string {
				return `{"operation":`
			},
			wantErr: "failed to parse parameters JSON:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			memory, setupErr := tt.setup()
			if setupErr != "" {
				t.Fatalf("setup failed: %s", setupErr)
			}
			tool := tools.NewLongMemoryTool(memory)

			output, err := tool.Execute(tt.params())
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}

			if err != nil {
				t.Fatalf("execute failed: %v", err)
			}
			if output != tt.want {
				t.Fatalf("unexpected output: %q", output)
			}
			if tt.check != nil {
				tt.check(t, memory, output)
			}
		})
	}
}
