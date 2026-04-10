package tool_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Notailab/go-agent/agent/core"
	"github.com/Notailab/go-agent/agent/tools"
)

func TestEditFileTool(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	filePath := filepath.Join(dir, "edit.txt")
	if err := os.WriteFile(filePath, []byte("Hello, world!"), 0o644); err != nil {
		t.Fatalf("write file failed: %v", err)
	}

	registry := core.NewToolRegistry(&tools.EditFileTool{})
	tool, ok := registry.Resolve("Edit_file")
	if !ok {
		t.Fatal("Edit_file tool not found")
	}

	output, err := tool.Execute(fmt.Sprintf(`{"file_path":%q,"old_text":%q,"new_text":%q}`, filePath, "Hello", "Hi"))
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}
	if !strings.Contains(output, filePath) {
		t.Fatalf("unexpected output: %q", output)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("read file failed: %v", err)
	}
	if got := string(data); got != "Hi, world!" {
		t.Fatalf("unexpected file content: %q", got)
	}
}
