package tool_test

import (
	"strings"
	"testing"

	"github.com/Notailab/go-agent/agent/core"
	"github.com/Notailab/go-agent/agent/storage"
	"github.com/Notailab/go-agent/agent/tools"
)

func TestLongMemoryTool_CreateUpdateDelete(t *testing.T) {
	t.Parallel()

	memory := core.NewMemory(
		storage.NewInMemoryChatStore(),
		storage.NewInMemoryLongStore(),
	)
	tool := tools.NewLongMemoryTool(memory)

	// create should require non-empty context
	_, err := tool.Execute(`{"operation":"create","index":0,"context":""}`)
	if err == nil || !strings.Contains(err.Error(), "context is required for create") {
		t.Fatalf("expected create context error, got: %v", err)
	}

	// successful create
	out, err := tool.Execute(`{"operation":"create","index":0,"context":"entry1"}`)
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if !strings.Contains(out, "create") {
		t.Fatalf("unexpected create output: %q", out)
	}
	if !strings.Contains(memory.LongMemory(), "entry1") {
		t.Fatalf("memory missing entry1: %q", memory.LongMemory())
	}

	// update requires index and non-empty context
	_, err = tool.Execute(`{"operation":"update","index":0,"context":""}`)
	if err == nil || !strings.Contains(err.Error(), "context is required for update") {
		t.Fatalf("expected update context error, got: %v", err)
	}

	out, err = tool.Execute(`{"operation":"update","index":0,"context":"entry1-updated"}`)
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}
	if !strings.Contains(out, "update") {
		t.Fatalf("unexpected update output: %q", out)
	}
	if !strings.Contains(memory.LongMemory(), "entry1-updated") {
		t.Fatalf("memory not updated: %q", memory.LongMemory())
	}

	// delete requires index; context is ignored
	out, err = tool.Execute(`{"operation":"delete","index":0,"context":"should-be-ignored"}`)
	if err != nil {
		t.Fatalf("delete failed: %v", err)
	}
	if !strings.Contains(out, "delete") {
		t.Fatalf("unexpected delete output: %q", out)
	}
	if strings.Contains(memory.LongMemory(), "entry1-updated") {
		t.Fatalf("memory not deleted: %q", memory.LongMemory())
	}
}
