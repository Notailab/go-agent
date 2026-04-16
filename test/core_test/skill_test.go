package core_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Notailab/go-agent/agent/core"
)

func TestNewSkillParsesMetadata(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func(t *testing.T) []string
		wantNil bool
		check   func(t *testing.T, skill *core.Skill)
	}{
		{
			name: "parse single skill metadata",
			setup: func(t *testing.T) []string {
				dir := t.TempDir()
				skillDir := filepath.Join(dir, "weather")
				if err := os.MkdirAll(skillDir, 0o755); err != nil {
					t.Fatalf("mkdir failed: %v", err)
				}
				content := "---\nname: weather\ndescription: Get current weather\n---\n# Weather\n"
				if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0o644); err != nil {
					t.Fatalf("write skill failed: %v", err)
				}
				return []string{dir}
			},
			check: func(t *testing.T, skill *core.Skill) {
				if skill == nil {
					t.Fatal("expected skill to be created")
				}
				if len(skill.SkillDir) != 1 {
					t.Fatalf("unexpected skill dir count: %d", len(skill.SkillDir))
				}
				if len(skill.SkillDesc) != 1 {
					t.Fatalf("unexpected skill desc count: %d", len(skill.SkillDesc))
				}
				if !strings.Contains(skill.SkillDesc[0], "name: weather") || !strings.Contains(skill.SkillDesc[0], "path:") {
					t.Fatalf("unexpected skill description: %q", skill.SkillDesc[0])
				}
			},
		},
		{
			name: "ignore markdown without metadata",
			setup: func(t *testing.T) []string {
				dir := t.TempDir()
				skillDir := filepath.Join(dir, "plain")
				if err := os.MkdirAll(skillDir, 0o755); err != nil {
					t.Fatalf("mkdir failed: %v", err)
				}
				if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("# Plain Skill\n"), 0o644); err != nil {
					t.Fatalf("write skill failed: %v", err)
				}
				return []string{dir}
			},
			check: func(t *testing.T, skill *core.Skill) {
				if skill == nil {
					t.Fatal("expected skill to be created")
				}
				if len(skill.SkillDir) != 1 {
					t.Fatalf("unexpected skill dir count: %d", len(skill.SkillDir))
				}
				if len(skill.SkillDesc) != 1 {
					t.Fatalf("unexpected skill desc count: %d", len(skill.SkillDesc))
				}
				if skill.SkillDesc[0] != "" {
					t.Fatalf("expected empty description, got %q", skill.SkillDesc[0])
				}
			},
		},
		{
			name: "parse multiple skills",
			setup: func(t *testing.T) []string {
				dir := t.TempDir()
				for _, item := range []struct {
					name string
					meta string
				}{
					{name: "weather", meta: "name: weather\ndescription: Get current weather"},
					{name: "news", meta: "name: news\ndescription: Get latest news"},
				} {
					skillDir := filepath.Join(dir, item.name)
					if err := os.MkdirAll(skillDir, 0o755); err != nil {
						t.Fatalf("mkdir failed: %v", err)
					}
					content := "---\n" + item.meta + "\n---\n# Skill\n"
					if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0o644); err != nil {
						t.Fatalf("write skill failed: %v", err)
					}
				}
				return []string{dir}
			},
			check: func(t *testing.T, skill *core.Skill) {
				if skill == nil {
					t.Fatal("expected skill to be created")
				}
				if len(skill.SkillDir) != 2 {
					t.Fatalf("unexpected skill dir count: %d", len(skill.SkillDir))
				}
				if len(skill.SkillDesc) != 2 {
					t.Fatalf("unexpected skill desc count: %d", len(skill.SkillDesc))
				}
				joined := strings.Join(skill.SkillDesc, "\n\n")
				if !strings.Contains(joined, "name: weather") || !strings.Contains(joined, "name: news") {
					t.Fatalf("unexpected skill descriptions: %q", joined)
				}
			},
		},
		{
			name: "missing path returns nil",
			setup: func(t *testing.T) []string {
				return []string{filepath.Join(t.TempDir(), "missing")}
			},
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			paths := tt.setup(t)
			skill := core.NewSkill(paths...)
			if tt.wantNil {
				if skill != nil {
					t.Fatalf("expected nil skill, got %#v", skill)
				}
				return
			}
			if tt.check != nil {
				tt.check(t, skill)
			}
		})
	}
}
