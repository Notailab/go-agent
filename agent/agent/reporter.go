package agent

import "github.com/Notailab/go-agent/agent/core"

type HookContext struct {
	Step      int
	Stream    bool
	Messages  []core.ChatMessage
	Tools     []core.FunctionTool
	Result    core.LLMResult
	TokenType string
	Delta     string
	ToolCall  core.ToolCall
	Output    string
	Error     error
}

type Reporter interface {
	BeforeLLM(ctx HookContext)
	OnLLM(ctx HookContext)
	AfterLLM(ctx HookContext)
	BeforeTool(ctx HookContext)
	AfterTool(ctx HookContext)
}

type NoopReporter struct{}

func (NoopReporter) BeforeLLM(ctx HookContext) {}

func (NoopReporter) OnLLM(ctx HookContext) {}

func (NoopReporter) AfterLLM(ctx HookContext) {}

func (NoopReporter) BeforeTool(ctx HookContext) {}

func (NoopReporter) AfterTool(ctx HookContext) {}
