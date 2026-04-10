package tool_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/Notailab/go-agent/agent/core"
	"github.com/Notailab/go-agent/agent/tools"
)

func TestReadFileTool(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	filePath := filepath.Join(dir, "read.txt")
	if err := os.WriteFile(filePath, []byte("Hello, Go Agent!"), 0o644); err != nil {
		t.Fatalf("write file failed: %v", err)
	}

	registry := core.NewToolRegistry(&tools.ReadFileTool{})
	tool, ok := registry.Resolve("Read_file")
	if !ok {
		t.Fatal("Read_file tool not found")
	}

	output, err := tool.Execute(fmt.Sprintf(`{"file_path":%q,"limit":5}`, filePath))
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}
	if output != "Hello" {
		t.Fatalf("unexpected output: %q", output)
	}
}
