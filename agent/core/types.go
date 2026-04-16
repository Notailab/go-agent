package core

import (
	"fmt"
	"math"
)

type Role string

const (
	RoleSystem    Role = "system"
	RoleAssistant Role = "assistant"
	RoleUser      Role = "user"
	RoleTool      Role = "tool"
)

type ChatMessage struct {
	Role       Role       `json:"role"`
	Content    string     `json:"content"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
	TimeStamp  string     `json:"timestamp,omitempty"`
}

type ChatRequest struct {
	Model       string         `json:"model"`
	Messages    []ChatMessage  `json:"messages"`
	Tools       []FunctionTool `json:"tools,omitempty"`
	Stream      bool           `json:"stream,omitempty"`
	Temperature float64        `json:"temperature,omitempty"`
}

type ToolCall struct {
	Id       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

type LLMMessage struct {
	Content          string     `json:"content"`
	ReasoningContent string     `json:"reasoning_content"`
	ToolCalls        []ToolCall `json:"tool_calls"`
}

type LLMChoice struct {
	Message LLMMessage `json:"message"`
}

type LLMResult struct {
	Choices []LLMChoice `json:"choices"`
	Usage   struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

type Parameters struct {
	Type       string           `json:"type"`
	Properties map[string]Param `json:"properties"`
	Required   []string         `json:"required"`
}

type Param struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	Default     any    `json:"default,omitempty"`
}

func convertToType(value any, targetType string) (any, error) {
	switch targetType {
	case "string":
		return asString(value)
	case "integer":
		return asInteger(value)
	case "number":
		return asNumber(value)
	case "boolean":
		return asBoolean(value)
	case "object":
		return asObject(value)
	case "array":
		return asArray(value)
	default:
		return value, nil
	}
}

func asString(value any) (string, error) {
	if value == nil {
		return "", nil
	}
	v, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("expected string, got %T", value)
	}
	return v, nil
}

func asInteger(value any) (any, error) {
	if value == nil {
		return 0, nil
	}
	switch v := value.(type) {
	case int:
		return v, nil
	case int64:
		return int(v), nil
	case float64:
		if math.Trunc(v) != v {
			return nil, fmt.Errorf("expected integer, got float %v", v)
		}
		return int(v), nil
	default:
		return nil, fmt.Errorf("expected integer, got %T", value)
	}
}

func asNumber(value any) (any, error) {
	if value == nil {
		return 0.0, nil
	}
	switch v := value.(type) {
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case float64:
		return v, nil
	default:
		return nil, fmt.Errorf("expected number, got %T", value)
	}
}

func asBoolean(value any) (any, error) {
	if value == nil {
		return false, nil
	}
	v, ok := value.(bool)
	if !ok {
		return nil, fmt.Errorf("expected boolean, got %T", value)
	}
	return v, nil
}

func asObject(value any) (any, error) {
	if value == nil {
		return map[string]any{}, nil
	}
	v, ok := value.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("expected object, got %T", value)
	}
	return v, nil
}

func asArray(value any) (any, error) {
	if value == nil {
		return []any{}, nil
	}
	v, ok := value.([]any)
	if !ok {
		return nil, fmt.Errorf("expected array, got %T", value)
	}
	return v, nil
}