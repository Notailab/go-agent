package tools

import (
	"fmt"
	"os"
	"strings"

	"github.com/Notailab/go-agent/agent/core"
)

type EditFileTool struct{}

func (t *EditFileTool) Name() string {
	return "Edit_file"
}

func (t *EditFileTool) Description() string {
	return "Edit the contents of a file."
}

func (t *EditFileTool) Parameters() core.Parameters {
	return core.Parameters{
		Type: "object",
		Properties: map[string]core.Param{
			"file_path": {
				Type:        "string",
				Description: "The path to the file to edit",
			},
			"old_text": {
				Type:        "string",
				Description: "The text to be replaced in the file",
			},
			"new_text": {
				Type:        "string",
				Description: "The new text to replace the old text in the file",
			},
		},
		Required: []string{"file_path", "old_text", "new_text"},
	}
}

func (t *EditFileTool) Execute(params string) (string, error) {
	paramMap, err := core.ParseParams(params, "file_path", "old_text", "new_text")
	if err != nil {
		return "", err
	}
	filePath, _ := paramMap["file_path"].(string)
	oldText, _ := paramMap["old_text"].(string)
	newText, _ := paramMap["new_text"].(string)

	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	replacedContent := strings.ReplaceAll(string(content), oldText, newText)
	err = os.WriteFile(filePath, []byte(replacedContent), 0644)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Content written to %s", filePath), nil
}
