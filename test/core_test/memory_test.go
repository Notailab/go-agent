package core_test

import (
	"testing"

	"github.com/Notailab/go-agent/agent/core"
	"github.com/Notailab/go-agent/agent/storage"
)

func TestMemoryReplaceRebuildsDialogIndex(t *testing.T) {
	t.Parallel()

	memory := core.NewMemory(
		storage.NewInMemoryChatStore(),
		storage.NewInMemoryLongStore(),
	)
	memory.AddChat(core.RoleUser, "user 1")
	memory.AddChat(core.RoleAssistant, "assistant 1")
	memory.AddChat(core.RoleUser, "user 2")

	memory.ReplaceChat(0, 2, []core.ChatMessage{{Role: core.RoleUser, Content: "summary"}})

	got := memory.ChatMemory()
	if len(got) != 2 {
		t.Fatalf("unexpected chat length: %d", len(got))
	}
	if got[0].Content != "summary" || got[1].Content != "user 2" {
		t.Fatalf("unexpected chat memory: %#v", got)
	}
}
