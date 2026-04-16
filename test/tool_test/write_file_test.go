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
	missingDirPath := filepath.Join(dir, "missing", "test.txt")

	registry := core.NewToolRegistry(&tools.WriteFileTool{})
	tool, ok := registry.Resolve("Write_file")
	if !ok {
		t.Fatal("Write_file tool not found")
	}

	tests := []struct {
		name    string
		params  string
		want    string
		wantErr string
		check   func(t *testing.T)
	}{
		{
			name:   "write new file",
			params: fmt.Sprintf(`{"file_path":%q,"content":%q}`, filePath, "Hello, Go Agent!"),
			want:   fmt.Sprintf("Content written to %s", filePath),
			check: func(t *testing.T) {
				data, err := os.ReadFile(filePath)
				if err != nil {
					t.Fatalf("read file failed: %v", err)
				}
				if got := string(data); got != "Hello, Go Agent!" {
					t.Fatalf("unexpected file content: %q", got)
				}
			},
		},
		{
			name:   "overwrite existing file",
			params: fmt.Sprintf(`{"file_path":%q,"content":%q}`, filePath, "updated content"),
			want:   fmt.Sprintf("Content written to %s", filePath),
			check: func(t *testing.T) {
				data, err := os.ReadFile(filePath)
				if err != nil {
					t.Fatalf("read file failed: %v", err)
				}
				if got := string(data); got != "updated content" {
					t.Fatalf("unexpected file content: %q", got)
				}
			},
		},
		{
			name:   "write empty content",
			params: fmt.Sprintf(`{"file_path":%q,"content":%q}`, filePath, ""),
			want:   fmt.Sprintf("Content written to %s", filePath),
			check: func(t *testing.T) {
				data, err := os.ReadFile(filePath)
				if err != nil {
					t.Fatalf("read file failed: %v", err)
				}
				if got := string(data); got != "" {
					t.Fatalf("unexpected file content: %q", got)
				}
			},
		},
		{
			name:    "missing parent directory",
			params:  fmt.Sprintf(`{"file_path":%q,"content":%q}`, missingDirPath, "content"),
			wantErr: "no such file or directory",
		},
		{
			name:    "invalid file path type",
			params:  `{"file_path":123,"content":"content"}`,
			wantErr: `invalid type for parameter "file_path"`,
		},
		{
			name:    "invalid content type",
			params:  fmt.Sprintf(`{"file_path":%q,"content":123}`, filePath),
			wantErr: `invalid type for parameter "content"`,
		},
		{
			name:    "missing file path",
			params:  `{"content":"content"}`,
			wantErr: "missing required parameters: [file_path]",
		},
		{
			name:    "missing content",
			params:  fmt.Sprintf(`{"file_path":%q}`, filePath),
			wantErr: "missing required parameters: [content]",
		},
		{
			name:    "empty params",
			params:  "",
			wantErr: "parameters JSON is empty",
		},
		{
			name:    "invalid json",
			params:  `{"file_path":`,
			wantErr: "failed to parse parameters JSON:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := tool.Execute(tt.params)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}

			if err != nil {
				t.Fatalf("execute failed: %v", err)
			}
			if output != tt.want {
				t.Fatalf("unexpected output: %q", output)
			}
			if tt.check != nil {
				tt.check(t)
			}
		})
	}
}
