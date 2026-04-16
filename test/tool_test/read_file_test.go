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

func TestReadFileTool(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	filePath := filepath.Join(dir, "read.txt")
	content := strings.Repeat("a", 3000)
	if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
		t.Fatalf("write file failed: %v", err)
	}

	registry := core.NewToolRegistry(&tools.ReadFileTool{})
	tool, ok := registry.Resolve("Read_file")
	if !ok {
		t.Fatal("Read_file tool not found")
	}

	tests := []struct {
		name    string
		params  string
		want    string
		wantErr string
	}{
		{
			name:   "with limit",
			params: fmt.Sprintf(`{"file_path":%q,"offset":7,"limit":2}`, filePath),
			want:   "aa",
		},
		{
			name:   "default limit",
			params: fmt.Sprintf(`{"file_path":%q,"offset":7}`, filePath),
			want:   content[7:2007],
		},
		{
			name:   "offset beyond file",
			params: fmt.Sprintf(`{"file_path":%q,"offset":3000}`, filePath),
			want:   "",
		},
		{
			name:    "negative offset",
			params:  fmt.Sprintf(`{"file_path":%q,"offset":-1}`, filePath),
			wantErr: "offset must be non-negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := tool.Execute(tt.params)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if err.Error() != tt.wantErr {
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
		})
	}
}

func TestReadFileToolErrors(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	existingFile := filepath.Join(dir, "existing.txt")
	content := "hello"
	if err := os.WriteFile(existingFile, []byte(content), 0o644); err != nil {
		t.Fatalf("write file failed: %v", err)
	}
	emptyFile := filepath.Join(dir, "empty.txt")
	if err := os.WriteFile(emptyFile, []byte(""), 0o644); err != nil {
		t.Fatalf("write empty file failed: %v", err)
	}

	registry := core.NewToolRegistry(&tools.ReadFileTool{})
	tool, ok := registry.Resolve("Read_file")
	if !ok {
		t.Fatal("Read_file tool not found")
	}

	tests := []struct {
		name    string
		params  string
		want    string
		wantErr string
	}{
		{
			name:   "offset at file end",
			params: fmt.Sprintf(`{"file_path":%q,"offset":5}`, existingFile),
			want:   "",
		},
		{
			name:   "limit zero reads to end",
			params: fmt.Sprintf(`{"file_path":%q,"offset":1,"limit":0}`, existingFile),
			want:   content[1:],
		},
		{
			name:   "empty file",
			params: fmt.Sprintf(`{"file_path":%q,"offset":0}`, emptyFile),
			want:   "",
		},
		{
			name:    "missing file",
			params:  fmt.Sprintf(`{"file_path":%q,"offset":0}`, filepath.Join(dir, "missing.txt")),
			wantErr: "no such file",
		},
		{
			name:    "missing required file path",
			params:  `{"offset":0}`,
			wantErr: "missing required parameters: [file_path]",
		},
		{
			name:    "missing required offset",
			params:  fmt.Sprintf(`{"file_path":%q}`, existingFile),
			wantErr: "missing required parameters: [offset]",
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
		})
	}
}
