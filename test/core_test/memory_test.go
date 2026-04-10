package core_test

import (
	"testing"

	"github.com/Notailab/go-agent/agent/core"
)

func TestMemoryDialogLimit(t *testing.T) {
	t.Parallel()

	memory := core.NewMemory()
	memory.SetDialogLimit(2)

	memory.AddChat(core.RoleUser, "u1")
	memory.AddChat(core.RoleAssistant, "a1")
	memory.AddChat(core.RoleUser, "u2")
	memory.AddChat(core.RoleAssistant, "a2")
	memory.AddChat(core.RoleUser, "u3")
	memory.AddChat(core.RoleAssistant, "a3")

	memory.CheckDialogLimit()

	got := memory.ChatMemory()
	if len(got) != 4 {
		t.Fatalf("unexpected chat length: %d", len(got))
	}
	if got[0].Content != "u2" || got[1].Content != "a2" || got[2].Content != "u3" || got[3].Content != "a3" {
		t.Fatalf("unexpected chat memory: %#v", got)
	}
	if len(memory.DialogIndex) != 2 || memory.DialogIndex[0] != 0 || memory.DialogIndex[1] != 2 {
		t.Fatalf("unexpected dialog index: %#v", memory.DialogIndex)
	}
}

func TestMemoryReplaceRebuildsDialogIndex(t *testing.T) {
	t.Parallel()

	memory := core.NewMemory()
	memory.AddChat(core.RoleUser, "user 1")
	memory.AddChat(core.RoleAssistant, "assistant 1")
	memory.AddChat(core.RoleUser, "user 2")

	memory.Replace(0, 2, []core.ChatMessage{{Role: core.RoleUser, Content: "summary"}})

	got := memory.ChatMemory()
	if len(got) != 2 {
		t.Fatalf("unexpected chat length: %d", len(got))
	}
	if got[0].Content != "summary" || got[1].Content != "user 2" {
		t.Fatalf("unexpected chat memory: %#v", got)
	}
	if len(memory.DialogIndex) != 2 || memory.DialogIndex[0] != 0 || memory.DialogIndex[1] != 1 {
		t.Fatalf("unexpected dialog index: %#v", memory.DialogIndex)
	}
}
