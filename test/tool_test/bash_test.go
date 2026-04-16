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

func TestBashTool(t *testing.T) {
	t.Parallel()

	cwdDir := t.TempDir()

	registry := core.NewToolRegistry(&tools.BashTool{})
	tool, ok := registry.Resolve("Bash")
	if !ok {
		t.Fatal("Bash tool not found")
	}

	tests := []struct {
		name    string
		params  func() string
		want    string
		wantErr string
		check   func(t *testing.T, output string)
	}{
		{
			name: "simple command",
			params: func() string {
				return `{"command":"echo hello"}`
			},
			want: "hello",
		},
		{
			name: "command with cwd",
			params: func() string {
				return fmt.Sprintf(`{"command":"pwd","cwd":%q}`, cwdDir)
			},
			check: func(t *testing.T, output string) {
				if got := strings.TrimSpace(output); got != cwdDir {
					t.Fatalf("unexpected cwd output: %q", got)
				}
			},
		},
		{
			name: "command failure with stderr",
			params: func() string {
				return `{"command":"echo out; echo err 1>&2; exit 3"}`
			},
			wantErr: "command failed with exit code 3",
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "out") || !strings.Contains(output, "err") {
					t.Fatalf("unexpected combined output: %q", output)
				}
			},
		},
		{
			name: "timeout",
			params: func() string {
				return `{"command":"echo start; sleep 2","timeout_seconds":1}`
			},
			wantErr: "command timed out after 1s",
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "start") {
					t.Fatalf("expected partial output before timeout, got %q", output)
				}
			},
		},
		{
			name: "invalid cwd path",
			params: func() string {
				return `{"command":"pwd","cwd":"/path/does/not/exist"}`
			},
			wantErr: "invalid cwd:",
		},
		{
			name: "cwd is file",
			params: func() string {
				dir := t.TempDir()
				filePath := filepath.Join(dir, "cwd.txt")
				if err := os.WriteFile(filePath, []byte("x"), 0o644); err != nil {
					t.Fatalf("write file failed: %v", err)
				}
				return fmt.Sprintf(`{"command":"pwd","cwd":%q}`, filePath)
			},
			wantErr: "invalid cwd: not a directory:",
		},
		{
			name:    "missing command",
			params:  func() string { return `{"cwd":"."}` },
			wantErr: "missing required parameters: [command]",
		},
		{
			name:    "empty command",
			params:  func() string { return `{"command":"   "}` },
			wantErr: "missing required parameter: command",
		},
		{
			name:    "invalid command type",
			params:  func() string { return `{"command":123}` },
			wantErr: `invalid type for parameter "command"`,
		},
		{
			name:    "invalid cwd type",
			params:  func() string { return `{"command":"pwd","cwd":123}` },
			wantErr: `invalid type for parameter "cwd"`,
		},
		{
			name:    "invalid timeout type",
			params:  func() string { return `{"command":"pwd","timeout_seconds":"fast"}` },
			wantErr: `invalid type for parameter "timeout_seconds"`,
		},
		{
			name:    "empty params",
			params:  func() string { return "" },
			wantErr: "parameters JSON is empty",
		},
		{
			name:    "invalid json",
			params:  func() string { return `{"command":` },
			wantErr: "failed to parse parameters JSON:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := tool.Execute(tt.params())
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("unexpected error: %v", err)
				}
				if tt.check != nil {
					tt.check(t, output)
				}
				return
			}

			if err != nil {
				t.Fatalf("execute failed: %v", err)
			}
			if tt.want != "" && strings.TrimSpace(output) != tt.want {
				t.Fatalf("unexpected output: %q", strings.TrimSpace(output))
			}
			if tt.check != nil {
				tt.check(t, output)
			}
		})
	}
}
