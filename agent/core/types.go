package core

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
}
