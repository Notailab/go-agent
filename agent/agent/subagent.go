package agent

import (
	"context"
	"fmt"

	"github.com/Notailab/go-agent/agent/core"
)

type SubAgent struct {
	agent *agent
}

func NewSubAgent(r *ReactAgent) *SubAgent {
	if r == nil || r.agent == nil {
		return &SubAgent{}
	}
	return &SubAgent{agent: r.agent.Fork()}
}

func SummarizeMessageSubAgent(r *ReactAgent) *SubAgent {
	if r == nil || r.agent == nil {
		return &SubAgent{}
	}
	return &SubAgent{
		agent: &agent{
			LLM:         r.agent.LLM,
			Tools:       nil,
			Memory:      nil,
			Skills:      nil,
			Temperature: r.agent.Temperature,
			staticSystemPrompt: `You are a summarize messages subagent.
Please summarize the following conversation history concisely,
keeping all key user requirements, preferences, decisions and context.
Do not omit important details.
You MUST NOT output any tool calls, function calls, or []-style commands.
Output ONLY plain text the summary, no extra explanation.`,
		},
	}
}

func (s *SubAgent) Run(ctx context.Context, messages []core.ChatMessage) (string, error) {
	if s == nil || s.agent == nil {
		return "", fmt.Errorf("subagent is nil")
	}
	msgs := []core.ChatMessage{
		{Role: core.RoleSystem, Content: s.agent.staticSystemPrompt},
	}
	msgs = append(msgs, messages...)
	return s.agent.simpleRun(ctx, msgs)
}
