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
	registry := core.NewToolRegistry(&tools.EditFileTool{})
	tool, ok := registry.Resolve("Edit_file")
	if !ok {
		t.Fatal("Edit_file tool not found")
	}

	tests := []struct {
		name    string
		content string
		params  func(filePath string) string
		want    string
		wantErr string
		check   func(t *testing.T, filePath string)
	}{
		{
			name:    "replace once",
			content: "Hello, world!",
			params: func(filePath string) string {
				return fmt.Sprintf(`{"file_path":%q,"old_text":%q,"new_text":%q}`, filePath, "Hello", "Hi")
			},
			want: "Content written to ",
			check: func(t *testing.T, filePath string) {
				data, err := os.ReadFile(filePath)
				if err != nil {
					t.Fatalf("read file failed: %v", err)
				}
				if got := string(data); got != "Hi, world!" {
					t.Fatalf("unexpected file content: %q", got)
				}
			},
		},
		{
			name:    "replace multiple occurrences",
			content: "foo foo foo",
			params: func(filePath string) string {
				return fmt.Sprintf(`{"file_path":%q,"old_text":%q,"new_text":%q}`, filePath, "foo", "bar")
			},
			want: "Content written to ",
			check: func(t *testing.T, filePath string) {
				data, err := os.ReadFile(filePath)
				if err != nil {
					t.Fatalf("read file failed: %v", err)
				}
				if got := string(data); got != "bar bar bar" {
					t.Fatalf("unexpected file content: %q", got)
				}
			},
		},
		{
			name:    "remove text",
			content: "Hello, world!",
			params: func(filePath string) string {
				return fmt.Sprintf(`{"file_path":%q,"old_text":%q,"new_text":%q}`, filePath, "Hello", "")
			},
			want: "Content written to ",
			check: func(t *testing.T, filePath string) {
				data, err := os.ReadFile(filePath)
				if err != nil {
					t.Fatalf("read file failed: %v", err)
				}
				if got := string(data); got != ", world!" {
					t.Fatalf("unexpected file content: %q", got)
				}
			},
		},
		{
			name:    "old text not found",
			content: "Hello, world!",
			params: func(filePath string) string {
				return fmt.Sprintf(`{"file_path":%q,"old_text":%q,"new_text":%q}`, filePath, "bye", "hi")
			},
			want: "Content written to ",
			check: func(t *testing.T, filePath string) {
				data, err := os.ReadFile(filePath)
				if err != nil {
					t.Fatalf("read file failed: %v", err)
				}
				if got := string(data); got != "Hello, world!" {
					t.Fatalf("unexpected file content: %q", got)
				}
			},
		},
		{
			name:    "missing file",
			content: "",
			params: func(filePath string) string {
				return fmt.Sprintf(`{"file_path":%q,"old_text":%q,"new_text":%q}`, filePath, "Hello", "Hi")
			},
			wantErr: "no such file or directory",
		},
		{
			name:    "missing file path",
			content: "Hello, world!",
			params: func(filePath string) string {
				return `{"old_text":"Hello","new_text":"Hi"}`
			},
			wantErr: "missing required parameters: [file_path]",
		},
		{
			name:    "missing old text",
			content: "Hello, world!",
			params: func(filePath string) string {
				return fmt.Sprintf(`{"file_path":%q,"new_text":%q}`, filePath, "Hi")
			},
			wantErr: "missing required parameters: [old_text]",
		},
		{
			name:    "missing new text",
			content: "Hello, world!",
			params: func(filePath string) string {
				return fmt.Sprintf(`{"file_path":%q,"old_text":%q}`, filePath, "Hello")
			},
			wantErr: "missing required parameters: [new_text]",
		},
		{
			name:    "invalid file path type",
			content: "Hello, world!",
			params: func(filePath string) string {
				return `{"file_path":123,"old_text":"Hello","new_text":"Hi"}`
			},
			wantErr: `invalid type for parameter "file_path"`,
		},
		{
			name:    "invalid old text type",
			content: "Hello, world!",
			params: func(filePath string) string {
				return fmt.Sprintf(`{"file_path":%q,"old_text":123,"new_text":%q}`, filePath, "Hi")
			},
			wantErr: `invalid type for parameter "old_text"`,
		},
		{
			name:    "invalid new text type",
			content: "Hello, world!",
			params: func(filePath string) string {
				return fmt.Sprintf(`{"file_path":%q,"old_text":%q,"new_text":123}`, filePath, "Hello")
			},
			wantErr: `invalid type for parameter "new_text"`,
		},
		{
			name:    "empty params",
			content: "Hello, world!",
			params: func(filePath string) string {
				return ""
			},
			wantErr: "parameters JSON is empty",
		},
		{
			name:    "invalid json",
			content: "Hello, world!",
			params: func(filePath string) string {
				return `{"file_path":`
			},
			wantErr: "failed to parse parameters JSON:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			filePath := filepath.Join(dir, "edit.txt")
			if tt.content != "" {
				if err := os.WriteFile(filePath, []byte(tt.content), 0o644); err != nil {
					t.Fatalf("write file failed: %v", err)
				}
			}

			output, err := tool.Execute(tt.params(filePath))
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
			if !strings.Contains(output, filePath) {
				t.Fatalf("unexpected output: %q", output)
			}
			if tt.check != nil {
				tt.check(t, filePath)
			}
		})
	}
}
