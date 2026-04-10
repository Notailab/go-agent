package core

import (
	"encoding/json"
	"fmt"
	"strings"
)

type Tool interface {
	Name() string
	Description() string
	Parameters() Parameters
	Execute(input string) (string, error)
}

type FunctionTool struct {
	Type     string `json:"type"`
	Function struct {
		Name        string     `json:"name"`
		Description string     `json:"description"`
		Parameters  Parameters `json:"parameters"`
	} `json:"function"`
}

func FunctionFromTool(tool Tool) FunctionTool {
	var function FunctionTool
	function.Type = "function"
	function.Function.Name = tool.Name()
	function.Function.Description = tool.Description()
	function.Function.Parameters = tool.Parameters()
	return function
}

type ToolRegistry struct {
	tools        map[string]Tool
	toolOrder    []string
	functionTool []FunctionTool
}

func ParseParams(params string, keys ...string) (map[string]interface{}, error) {
	var paramMap map[string]interface{}
	err := json.Unmarshal([]byte(params), &paramMap)
	if err != nil {
		return nil, fmt.Errorf("failed to parse parameters: %v", err)
	}

	result := make(map[string]interface{})
	for _, key := range keys {
		if val, ok := paramMap[key]; ok {
			result[key] = val
		} else {
			return nil, fmt.Errorf("missing required parameter: %s", key)
		}
	}
	return result, nil
}

func NewToolRegistry(tools ...Tool) *ToolRegistry {
	registry := ToolRegistry{tools: make(map[string]Tool)}
	for _, tool := range tools {
		_ = registry.Register(tool)
	}
	return &registry
}

func (r *ToolRegistry) Register(tool Tool) error {
	if r == nil {
		return fmt.Errorf("tool registry is nil")
	}
	if tool == nil {
		return fmt.Errorf("tool is nil")
	}

	name := strings.TrimSpace(tool.Name())
	if name == "" {
		return fmt.Errorf("tool name is empty")
	}
	if r.tools == nil {
		r.tools = make(map[string]Tool)
	}
	r.tools[name] = tool
	r.functionTool = nil
	return nil
}

func (r *ToolRegistry) Resolve(name string) (Tool, bool) {
	if r == nil {
		return nil, false
	}
	tool, ok := r.tools[strings.TrimSpace(name)]
	return tool, ok
}

func (r *ToolRegistry) Define() []FunctionTool {
	if r == nil {
		return []FunctionTool{}
	}

	if r.functionTool != nil && len(r.functionTool) == len(r.tools) {
		return r.functionTool
	}

	r.functionTool = make([]FunctionTool, 0, len(r.tools))
	for _, tool := range r.tools {
		r.functionTool = append(r.functionTool, FunctionFromTool(tool))
	}
	return r.functionTool
}

func (r *ToolRegistry) Clone() *ToolRegistry {
	if r == nil {
		return nil
	}

	clone := &ToolRegistry{
		tools: make(map[string]Tool, len(r.tools)),
	}
	for name, tool := range r.tools {
		clone.tools[name] = tool
	}
	if len(r.functionTool) > 0 {
		clone.functionTool = append([]FunctionTool(nil), r.functionTool...)
	}
	return clone
}
