package agent

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/Notailab/go-agent/agent/core"
)

type agent struct {
	LLM                *core.LLMClient
	Tools              *core.ToolRegistry
	Memory             *core.Memory
	Skills             *core.Skill
	Reporter           Reporter
	MaxSteps           int
	MaxTokens          int
	CurTokens          int
	Temperature        float64
	staticSystemPrompt string
}

type agentOption func(*agent)

func WithLLM(baseUrl, model, apiKey string) agentOption {
	return func(a *agent) {
		a.LLM = core.NewLLMClient(baseUrl, apiKey, model)
	}
}

func WithTools(tools ...core.Tool) agentOption {
	return func(a *agent) {
		a.Tools = core.NewToolRegistry(tools...)
	}
}

func WithMemory(memory *core.Memory) agentOption {
	return func(a *agent) {
		a.Memory = memory
	}
}

func WithSkills(paths ...string) agentOption {
	return func(a *agent) {
		a.Skills = core.NewSkill(paths...)
	}
}

func WithReporter(reporter Reporter) agentOption {
	return func(a *agent) {
		if reporter == nil {
			a.Reporter = NoopReporter{}
			return
		}
		a.Reporter = reporter
	}
}

func WithMaxSteps(maxSteps int) agentOption {
	return func(a *agent) {
		a.MaxSteps = maxSteps
	}
}

func WithMaxTokens(maxTokens int) agentOption {
	return func(a *agent) {
		a.MaxTokens = maxTokens
	}
}

func WithTemperature(temperature float64) agentOption {
	return func(a *agent) {
		a.Temperature = temperature
	}
}

func WithStaticSystemPrompt(prompt string) agentOption {
	return func(a *agent) {
		a.staticSystemPrompt = prompt
	}
}

func (a *agent) Fork() *agent {
	if a == nil {
		return &agent{}
	}

	var tools *core.ToolRegistry
	if a.Tools != nil {
		tools = a.Tools.Clone()
	}

	var memory *core.Memory
	if a.Memory != nil {
		memory = a.Memory.Clone()
	}

	var skills *core.Skill
	if a.Skills != nil {
		skills = a.Skills.Clone()
	}

	return &agent{
		LLM:                a.LLM,
		Tools:              tools,
		Memory:             memory,
		Skills:             skills,
		Reporter:           a.Reporter,
		MaxSteps:           a.MaxSteps,
		MaxTokens:          a.MaxTokens,
		CurTokens:          0,
		Temperature:        a.Temperature,
		staticSystemPrompt: a.staticSystemPrompt,
	}
}

func (a *agent) simpleRun(ctx context.Context, messages []core.ChatMessage) (string, error) {
	if a.LLM == nil {
		return "", errors.New("llm is empty")
	}

	res, err := a.LLM.Chat(ctx, messages, nil, a.Temperature)
	if err != nil {
		return "", err
	}
	if len(res.Choices) == 0 {
		return "", fmt.Errorf("no choices returned from llm")
	}
	a.CurTokens = res.Usage.TotalTokens
	return res.Choices[0].Message.Content, nil
}

func (a *agent) loopRun(ctx context.Context, messages []core.ChatMessage) (string, error) {
	return a.runLoop(ctx, messages, false)
}

func (a *agent) loopStreamRun(ctx context.Context, messages []core.ChatMessage) (string, error) {
	return a.runLoop(ctx, messages, true)
}

func (a *agent) runLoop(ctx context.Context, messages []core.ChatMessage, stream bool) (string, error) {
	if a.LLM == nil {
		return "", errors.New("llm is empty")
	}
	if a.Memory == nil {
		return "", errors.New("memory is empty")
	}
	if a.Tools == nil {
		return "", errors.New("tools is empty")
	}
	if a.Reporter == nil {
		a.Reporter = NoopReporter{}
	}

	for step := 0; step < a.MaxSteps; step++ {
		if ctx != nil && ctx.Err() != nil {
			return "", ctx.Err()
		}

		msgs := append(messages, a.Memory.ChatMemory()...)
		tools := a.Tools.Define()

		hookContext := HookContext{
			Step:     step,
			Stream:   stream,
			Messages: msgs,
			Tools:    tools,
		}

		a.Reporter.BeforeLLM(hookContext)

		var (
			res core.LLMResult
			err error
		)

		if stream {
			res, err = a.LLM.StreamChat(ctx, msgs, tools, a.Temperature,
				func(tokenType, token string) {
					a.Reporter.OnLLM(HookContext{
						Step:      step,
						Stream:    true,
						Messages:  msgs,
						Tools:     tools,
						TokenType: tokenType,
						Delta:     token,
					})
				},
			)
		} else {
			res, err = a.LLM.Chat(ctx, msgs, tools, a.Temperature)
		}

		hookContext.Result = res
		hookContext.Error = err
		a.Reporter.AfterLLM(hookContext)

		if err != nil {
			return "", err
		}
		if len(res.Choices) == 0 {
			return "", fmt.Errorf("no choices returned from llm")
		}

		msg := res.Choices[0].Message
		a.CurTokens = res.Usage.TotalTokens

		if msg.Content != "" {
			a.Memory.AddChat(core.RoleAssistant, msg.Content)
		}

		if len(msg.ToolCalls) == 0 {
			if msg.Content != "" {
				return msg.Content, nil
			}
			return msg.ReasoningContent, nil
		}

		a.Memory.AddToolCall(msg.ToolCalls)

		type toolExecutionResult struct {
			output string
			err    error
		}

		results := make([]toolExecutionResult, len(msg.ToolCalls))
		var wg sync.WaitGroup

		for i, tc := range msg.ToolCalls {
			name := tc.Function.Name
			args := tc.Function.Arguments

			toolHook := HookContext{
				Step:     step,
				Stream:   stream,
				Messages: msgs,
				Tools:    tools,
				Result:   res,
				ToolCall: tc,
			}

			a.Reporter.BeforeTool(toolHook)

			wg.Add(1)
			go func(index int, toolName, toolArgs string) {
				defer wg.Done()

				var (
					output string
					execErr error
				)

				tool, ok := a.Tools.Resolve(toolName)
				if ok {
					output, execErr = tool.Execute(toolArgs)
					if execErr != nil {
						output = fmt.Sprintf("error executing tool %s: %v", toolName, execErr)
					}
				} else {
					execErr = fmt.Errorf("tool %s not found", toolName)
					output = execErr.Error()
				}

				results[index] = toolExecutionResult{output: output, err: execErr}
			}(i, name, args)
		}

		wg.Wait()

		for i, tc := range msg.ToolCalls {
			result := results[i]
			toolCallID := tc.Id
			a.Memory.AddToolResult(toolCallID, result.output)
			a.Reporter.AfterTool(HookContext{
				Step:     step,
				Stream:   stream,
				Messages: msgs,
				Tools:    tools,
				Result:   res,
				ToolCall: tc,
				Output:   result.output,
				Error:    result.err,
			})
		}
	}

	return "", fmt.Errorf("max steps reached")
}
