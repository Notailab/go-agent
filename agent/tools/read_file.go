package tools

import (
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
				Default:     2000,
			},
		},
		Required: []string{"file_path", "offset"},
	}
}

func (t *ReadFileTool) Execute(paramsJson string) (string, error) {
	params, err := core.ParseToolParams(paramsJson, t.Parameters())
	if err != nil {
		return "", err
	}

	filePath := params["file_path"].(string)
	offset := params["offset"].(int)
	limit := params["limit"].(int)

	content, err := read_file(filePath, offset, limit)
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
