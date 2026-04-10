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

func TestWriteFileTool(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	filePath := filepath.Join(dir, "test.txt")

	registry := core.NewToolRegistry(&tools.WriteFileTool{})
	tool, ok := registry.Resolve("Write_file")
	if !ok {
		t.Fatal("Write_file tool not found")
	}

	output, err := tool.Execute(fmt.Sprintf(`{"file_path":%q,"content":%q}`, filePath, "Hello, Go Agent!"))
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
	if got := string(data); got != "Hello, Go Agent!" {
		t.Fatalf("unexpected file content: %q", got)
	}
}
