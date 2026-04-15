package tools

import (
	"encoding/json"
	"fmt"
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
			"offset": {
				Type:        "integer",
				Description: "The byte offset to start reading from",
			},
			"limit": {
				Type:        "integer",
				Description: "Optional limit on number of bytes to read from the offset; defaults to 2000 bytes",
			},
		},
		Required: []string{"file_path", "offset"},
	}
}

func (t *ReadFileTool) Execute(params string) (string, error) {
	var paramMap map[string]interface{}
	err := json.Unmarshal([]byte(params), &paramMap)
	if err != nil {
		return "", err
	}
	filePath, ok := paramMap["file_path"].(string)
	if !ok || filePath == "" {
		return "", fmt.Errorf("missing required parameter: file_path")
	}
	offsetValue, ok := paramMap["offset"].(float64)
	if !ok {
		return "", fmt.Errorf("missing required parameter: offset")
	}
	limit := 2000
	if rawLimit, ok := paramMap["limit"]; ok {
		if limitValue, ok := rawLimit.(float64); ok {
			limit = int(limitValue)
		}
	}

	content, err := read_file(filePath, int(offsetValue), limit)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func read_file(filePath string, offset int, limit int) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	if offset < 0 {
		return "", fmt.Errorf("offset must be non-negative")
	}
	if offset >= len(content) {
		return "", nil
	}
	content = content[offset:]
	if limit > 0 && len(content) > limit {
		content = content[:limit]
	}
	return string(content), nil
}
