package tools

import (
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
				Description: "The operation to perform: create, update, or delete.\nAll three parameters must be present in the request.\n- create: set `index` empty/0 and provide a non-empty `context`.\n- update: provide a valid 1-based `index` and a non-empty `context`.\n- delete: provide a valid 1-based `index` and set `context` to empty.",
			},
			"index": {
				Type:        "integer",
				Description: "The 1-based index of the memory entry.\nFor `create` pass 0 (or an empty/zero value). Required and must be valid for `update` and `delete`.",
			},
			"context": {
				Type:        "string",
				Description: "The memory content. Required and must be non-empty for `create` and `update`. For `delete` pass an empty string.",
			},
		},
		Required: []string{"operation", "index", "context"},
	}
}

func (t *LongMemoryTool) Execute(params string) (string, error) {
	paramMap, err := core.ParseParams(params, "operation", "index", "context")

	if err != nil {
		return "", err
	}

	operation := strings.ToLower(strings.TrimSpace(paramMap["operation"].(string)))
	index := int(paramMap["index"].(float64))
	context := strings.TrimSpace(paramMap["context"].(string))

	var op core.LongMemoryOperation

	switch operation {
	case "create":
		if strings.TrimSpace(context) == "" {
			return "", fmt.Errorf("context is required for create")
		}
		op = core.LongMemoryCreate
	case "update":
		if strings.TrimSpace(context) == "" {
			return "", fmt.Errorf("context is required for update")
		}
		op = core.LongMemoryUpdate
	case "delete":
		context = ""
		op = core.LongMemoryDelete
	default:
		return "", fmt.Errorf("unsupported operation: %s", operation)
	}

	err = t.Memory.OperateLongMemory(op, index, context)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Long-term memory %s operation successful", operation), nil
}
