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

	dir := t.TempDir()
	skillDir := filepath.Join(dir, "weather")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}

	content := "---\nname: weather\ndescription: Get current weather\n---\n# Weather\n"
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0o644); err != nil {
		t.Fatalf("write skill failed: %v", err)
	}

	skill := core.NewSkill(dir)
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
}
