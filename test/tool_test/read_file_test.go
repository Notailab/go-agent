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
	content := make([]byte, 3000)
	for i := range content {
		content[i] = 'a'
	}
	if err := os.WriteFile(filePath, content, 0o644); err != nil {
		t.Fatalf("write file failed: %v", err)
	}

	registry := core.NewToolRegistry(&tools.ReadFileTool{})
	tool, ok := registry.Resolve("Read_file")
	if !ok {
		t.Fatal("Read_file tool not found")
	}

	output, err := tool.Execute(fmt.Sprintf(`{"file_path":%q,"offset":7,"limit":2}`, filePath))
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}
	if output != "aa" {
		t.Fatalf("unexpected output: %q", output)
	}

	output, err = tool.Execute(fmt.Sprintf(`{"file_path":%q,"offset":7}`, filePath))
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}
	if len(output) != 2000 {
		t.Fatalf("unexpected output length: %d", len(output))
	}
	if output != string(content[7:2007]) {
		t.Fatalf("unexpected output: %q", output)
	}
}
