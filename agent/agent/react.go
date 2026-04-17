package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Notailab/go-agent/agent/core"
	"github.com/Notailab/go-agent/agent/storage"
	"github.com/Notailab/go-agent/agent/tools"
)

type ReactAgent struct {
	agent *agent
}

func NewReactAgent(opts ...agentOption) *ReactAgent {
	react := &ReactAgent{
		agent: &agent{
			Tools:              &core.ToolRegistry{},
			Reporter:           NoopReporter{},
			MaxSteps:           50,
			MaxTokens:          65536,
			staticSystemPrompt: core.GetStaticSystemPrompt(),
		},
	}
	for _, opt := range opts {
		opt(react.agent)
	}
	return react
}

func DefaultReactAgent(baseUrl, model, apiKey string) *ReactAgent {
	memory := core.NewMemory(
		storage.NewInMemoryChatStore(),
		storage.NewInMemoryLongStore(),
	)
	return NewReactAgent(
		WithLLM(baseUrl, model, apiKey),
		WithTools(
			&tools.BashTool{},
			&tools.EditFileTool{},
			&tools.ReadFileTool{},
			&tools.WriteFileTool{},
			tools.NewLongMemoryTool(memory),
		),
		WithMemory(memory),
		WithSkills("skills"),
	)
}

func (r *ReactAgent) CurTokens() int {
	return r.agent.CurTokens
}

func (r *ReactAgent) SystemPrompt() string {
	if r.agent == nil {
		return ""
	}

	parts := []string{r.agent.staticSystemPrompt}
	if r.agent.Skills != nil {
		parts = append(parts, r.agent.Skills.SystemPrompt())
	}
	if r.agent.Memory != nil {
		parts = append(parts, r.agent.Memory.SystemPrompt())
	}
	return strings.Join(parts, "\n\n")
}

func (r *ReactAgent) SystemReminder() string {
	if r.agent == nil || r.agent.Memory == nil {
		return ""
	}

	parts := []string{
		"<core-rules>",
		r.agent.Memory.LongMemory(),
		"Current time: " + time.Now().Format(time.RFC3339),
		"</core-rules>",
	}
	return strings.Join(parts, "\n\n")
}

func MessagesToPlainText(msgs []core.ChatMessage) string {
	var sb strings.Builder

	for _, msg := range msgs {
		switch msg.Role {
		case core.RoleUser:
			sb.WriteString("[user] ")
			sb.WriteString(strings.TrimSpace(msg.Content))
			sb.WriteString("\n")
		case core.RoleAssistant:
			sb.WriteString("[assistant] ")

			if strings.TrimSpace(msg.Content) != "" {
				sb.WriteString(strings.TrimSpace(msg.Content))
			}

			for _, tc := range msg.ToolCalls {
				funcName := tc.Function.Name
				args := tc.Function.Arguments

				var argMap map[string]interface{}
				if err := json.Unmarshal([]byte(args), &argMap); err == nil {
					if cmd, ok := argMap["command"].(string); ok {
						args = cmd
					}
				}

				sb.WriteString("(tool call: ")
				sb.WriteString(funcName)
				sb.WriteString(" → ")
				sb.WriteString(args)
				sb.WriteString(")")
			}

			sb.WriteString("\n")

		case core.RoleTool:
			sb.WriteString("[tool] ")
			content := strings.TrimSpace(msg.Content)
			if content == "" {
				content = "(no result)"
			}
			sb.WriteString(content)
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

func (r *ReactAgent) CompactHistory() {
	if r == nil {
		return
	}
	if r.agent.Memory == nil {
		return
	}

	if float64(r.agent.CurTokens) < 0.75*float64(r.agent.MaxTokens) {
		return
	}

	snapshot := r.agent.Memory.ChatMemory()
	if len(snapshot) == 0 {
		return
	}

	targetIndex := len(snapshot) * 7 / 10
	if targetIndex >= len(snapshot) {
		targetIndex = len(snapshot) - 1
	}

	cutIndex := -1
	for i := targetIndex; i >= 0; i-- {
		if snapshot[i].Role == core.RoleUser {
			cutIndex = i
			break
		}
	}
	if cutIndex <= 0 || cutIndex >= len(snapshot) {
		return
	}

	prefix := snapshot[:cutIndex]

	go func(prefix []core.ChatMessage) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
		defer cancel()

		messages := []core.ChatMessage{{
			Role: core.RoleUser,
			Content: "Below is a real conversation:\n\n" +
				MessagesToPlainText(prefix) +
				"\n\nPlease summarize the above conversation history concisely.",
		}}

		summary, err := SummarizeMessageSubAgent(r).Run(ctx, messages)
		if err != nil || summary == "" {
			return
		}

		summaryMessage := []core.ChatMessage{{
			Role:    core.RoleUser,
			Content: summary,
		}}
		r.agent.Memory.ReplaceChat(0, len(prefix), summaryMessage)
	}(prefix)
}

func (r *ReactAgent) Run(ctx context.Context, userInput string) (string, error) {
	if r.agent.LLM == nil {
		return "", fmt.Errorf("llm client is nil")
	}
	if r.agent.Memory == nil {
		return "", fmt.Errorf("memory is nil")
	}
	if r.agent.Reporter == nil {
		r.agent.Reporter = NoopReporter{}
	}
	if resetter, ok := r.agent.Reporter.(interface{ ResetDialog() }); ok {
		resetter.ResetDialog()
	}

	var messages []core.ChatMessage
	messages = append(messages, core.ChatMessage{Role: core.RoleSystem, Content: r.SystemPrompt()})
	messages = append(messages, core.ChatMessage{Role: core.RoleUser, Content: r.SystemReminder()})

	r.agent.Memory.AddChat(core.RoleUser, userInput)

	defer r.CompactHistory()
	return r.agent.loopRun(ctx, messages)
}

func (r *ReactAgent) StreamRun(ctx context.Context, userInput string) (string, error) {
	if r.agent.LLM == nil {
		return "", fmt.Errorf("llm client is nil")
	}
	if r.agent.Memory == nil {
		return "", fmt.Errorf("memory is nil")
	}
	if r.agent.Reporter == nil {
		r.agent.Reporter = NoopReporter{}
	}
	if resetter, ok := r.agent.Reporter.(interface{ ResetDialog() }); ok {
		resetter.ResetDialog()
	}

	var messages []core.ChatMessage
	messages = append(messages, core.ChatMessage{Role: core.RoleSystem, Content: r.SystemPrompt()})
	messages = append(messages, core.ChatMessage{Role: core.RoleUser, Content: r.SystemReminder()})

	r.agent.Memory.AddChat(core.RoleUser, userInput)

	defer r.CompactHistory()
	return r.agent.loopStreamRun(ctx, messages)
}
