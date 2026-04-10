package core_test

import (
	"strings"
	"testing"

	"github.com/Notailab/go-agent/agent/core"
)

func TestGetStaticSystemPrompt(t *testing.T) {
	t.Parallel()

	prompt := core.GetStaticSystemPrompt()
	for _, want := range []string{
		"You are an interactive agent assisting users with software engineering tasks.",
		"#System",
		"# Doing tasks",
		"# Using your tools",
		"# Tone and Output",
		"===== System Prompt Dynamic Boundary =====",
	} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("prompt missing %q: %s", want, prompt)
		}
	}
}
