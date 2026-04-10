package tools

import (
	"os"

	"github.com/Notailab/go-agent/agent/core"
)

type ReadFileTool struct{}

func (t *ReadFileTool) Name() string {
	return "Read_file"
}

func (t *ReadFileTool) Description() string {
	return "Read the contents of a file."
}

func (t *ReadFileTool) Parameters() core.Parameters {
	return core.Parameters{
		Type: "object",
		Properties: map[string]core.Param{
			"file_path": {
				Type:        "string",
				Description: "The path to the file to read",
			},
			"limit": {
				Type:        "integer",
				Description: "Optional limit on number of bytes to read",
			},
		},
		Required: []string{"file_path", "limit"},
	}
}

func (t *ReadFileTool) Execute(params string) (string, error) {
	paramMap, err := core.ParseParams(params, "file_path", "limit")
	if err != nil {
		return "", err
	}
	filePath, _ := paramMap["file_path"].(string)
	limit, _ := paramMap["limit"].(float64)

	content, err := read_file(filePath, int(limit))
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func read_file(filePath string, limit int) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	if limit > 0 && len(content) > limit {
		content = content[:limit]
	}
	return string(content), nil
}
