package core_test

import (
	"strings"
	"testing"

	"github.com/Notailab/go-agent/agent/core"
)

func TestGetStaticSystemPrompt(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		prompt  string
		want    []string
	}{
		{
			name:   "default prompt contains all sections",
			prompt: core.GetStaticSystemPrompt(),
			want: []string{
				"You are an interactive agent assisting users with software engineering tasks.",
				"# System",
				"# Doing tasks",
				"# Executing actions with care",
				"# Using your tools",
				"# Tone and Output",
				"===== System Prompt Dynamic Boundary =====",
			},
		},
		{
			name:   "override prompt sections",
			prompt: core.BuildStaticSystemPrompt(core.StaticSystemPromptOverrides{IntroSection: ptrString("intro"), BoundarySection: ptrString("boundary")}),
			want: []string{
				"intro",
				"boundary",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, want := range tt.want {
				if !strings.Contains(tt.prompt, want) {
					t.Fatalf("prompt missing %q: %s", want, tt.prompt)
				}
			}
		})
	}
}

func TestBuildStaticSystemPromptOrder(t *testing.T) {
	t.Parallel()

	prompt := core.BuildStaticSystemPrompt(core.StaticSystemPromptOverrides{})
	sections := []string{
		"You are an interactive agent assisting users with software engineering tasks.",
		"# System",
		"# Doing tasks",
		"# Executing actions with care",
		"# Using your tools",
		"# Tone and Output",
		"===== System Prompt Dynamic Boundary =====",
	}

	lastIndex := -1
	for _, section := range sections {
		index := strings.Index(prompt, section)
		if index == -1 {
			t.Fatalf("prompt missing section %q: %s", section, prompt)
		}
		if index < lastIndex {
			t.Fatalf("section %q is out of order: %s", section, prompt)
		}
		lastIndex = index
	}
}

func TestBuildStaticSystemPromptOverrides(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		overrides core.StaticSystemPromptOverrides
		want      []string
		notWant   []string
	}{
		{
			name:      "custom intro and boundary",
			overrides: core.StaticSystemPromptOverrides{IntroSection: ptrString("custom intro"), BoundarySection: ptrString("custom boundary")},
			want:      []string{"custom intro", "custom boundary"},
			notWant:   []string{"You are an interactive agent assisting users with software engineering tasks.", "===== System Prompt Dynamic Boundary ====="},
		},
		{
			name:      "custom task section",
			overrides: core.StaticSystemPromptOverrides{DoingTasksSection: ptrString("do custom things")},
			want:      []string{"do custom things"},
			notWant:   []string{"# Doing tasks"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prompt := core.BuildStaticSystemPrompt(tt.overrides)
			for _, want := range tt.want {
				if !strings.Contains(prompt, want) {
					t.Fatalf("prompt missing %q: %s", want, prompt)
				}
			}
			for _, notWant := range tt.notWant {
				if strings.Contains(prompt, notWant) {
					t.Fatalf("prompt unexpectedly contains %q: %s", notWant, prompt)
				}
			}
		})
	}
}

func ptrString(value string) *string {
	return &value
}
