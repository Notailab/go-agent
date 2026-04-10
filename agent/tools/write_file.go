package tools

import (
	"fmt"
	"os"

	"github.com/Notailab/go-agent/agent/core"
)

type WriteFileTool struct{}

func (t *WriteFileTool) Name() string {
	return "Write_file"
}

func (t *WriteFileTool) Description() string {
	return "Write content to a file."
}

func (t *WriteFileTool) Parameters() core.Parameters {
	return core.Parameters{
		Type: "object",
		Properties: map[string]core.Param{
			"file_path": {
				Type:        "string",
				Description: "The path to the file to write",
			},
			"content": {
				Type:        "string",
				Description: "The content to write to the file",
			},
		},
		Required: []string{"file_path", "content"},
	}
}

func (t *WriteFileTool) Execute(params string) (string, error) {
	paramMap, err := core.ParseParams(params, "file_path", "content")
	if err != nil {
		return "", err
	}
	filePath, _ := paramMap["file_path"].(string)
	content, _ := paramMap["content"].(string)

	err = os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Content written to %s", filePath), nil
}
