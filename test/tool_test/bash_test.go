package tool_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/Notailab/go-agent/agent/core"
	"github.com/Notailab/go-agent/agent/tools"
)

func TestBashTool(t *testing.T) {
	t.Parallel()

	registry := core.NewToolRegistry(&tools.BashTool{})
	tool, ok := registry.Resolve("Bash")
	if !ok {
		t.Fatal("Bash tool not found")
	}

	output, err := tool.Execute(`{"command":"echo hello"}`)
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}
	if got := strings.TrimSpace(output); got != "hello" {
		t.Fatalf("unexpected output: %q", got)
	}

	_, err = tool.Execute(fmt.Sprintf(`{"command":"pwd","cwd":%q}`, t.TempDir()))
	if err != nil {
		t.Fatalf("execute with cwd failed: %v", err)
	}
}
