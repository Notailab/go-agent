package tools

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Notailab/go-agent/agent/core"
)

type LongMemoryTool struct {
	Memory *core.Memory
}

func NewLongMemoryTool(memory *core.Memory) *LongMemoryTool {
	return &LongMemoryTool{Memory: memory}
}

func (t *LongMemoryTool) Name() string {
	return "Long_memory"
}

func (t *LongMemoryTool) Description() string {
	return "Create, update, or delete long-term memory entries."
}

func (t *LongMemoryTool) Parameters() core.Parameters {
	return core.Parameters{
		Type: "object",
		Properties: map[string]core.Param{
			"operation": {
				Type:        "string",
				Description: "The operation to perform: create, update, or delete",
			},
			"index": {
				Type:        "integer",
				Description: "The 1-based index of the memory entry",
			},
			"context": {
				Type:        "string",
				Description: "The memory content to write",
			},
		},
		Required: []string{"operation", "index"},
	}
}

func (t *LongMemoryTool) Execute(params string) (string, error) {
	if t == nil || t.Memory == nil {
		return "", fmt.Errorf("memory is nil")
	}

	var req struct {
		Operation string  `json:"operation"`
		Index     int     `json:"index"`
		Context   *string `json:"context"`
	}
	if err := json.Unmarshal([]byte(params), &req); err != nil {
		return "", fmt.Errorf("failed to parse parameters: %v", err)
	}

	operation := strings.ToLower(strings.TrimSpace(req.Operation))
	context := ""
	if req.Context != nil {
		context = *req.Context
	}

	switch operation {
	case "create":
		if strings.TrimSpace(context) == "" {
			return "", fmt.Errorf("context is required for create")
		}
	case "update":
		if strings.TrimSpace(context) == "" {
			return "", fmt.Errorf("context is required for update")
		}
	case "delete":
		context = ""
	default:
		return "", fmt.Errorf("unsupported operation: %s", req.Operation)
	}

	return t.Memory.ApplyLongMemoryOperation(operation, req.Index, context)
}
