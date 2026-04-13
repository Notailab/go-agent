package core

import (
	"fmt"
	"strings"
	"time"
)

type ChatMemoryStore interface {
	Get(int) (ChatMessage, error)
	Append(ChatMessage) error
	Update(int, ChatMessage) error
	Replace(int, int, []ChatMessage) error
	Delete(int) error
	List() ([]ChatMessage, error)
	Count() (int, error)
	Clear() error
	Clone() ChatMemoryStore
}

type LongMemoryOperation int

const (
	LongMemoryCreate LongMemoryOperation = iota
	LongMemoryUpdate
	LongMemoryDelete
)

type LongMemoryStore interface {
	Get(int) (string, error)
	Append(string) error
	Update(int, string) error
	Replace(int, int, []string) error
	Delete(int) error
	List() ([]string, error)
	Count() (int, error)
	Clear() error
	Clone() LongMemoryStore
}

type Memory struct {
	chatMemoryStore ChatMemoryStore
	longMemoryStore LongMemoryStore
}

func NewMemory(chatStore ChatMemoryStore, longStore LongMemoryStore) *Memory {
	if chatStore == nil || longStore == nil {
		return nil
	}
	return &Memory{
		chatMemoryStore: chatStore,
		longMemoryStore: longStore,
	}
}

func (m *Memory) Clone() *Memory {
	if m == nil || m.chatMemoryStore == nil || m.longMemoryStore == nil {
		return nil
	}
	return &Memory{
		chatMemoryStore: m.chatMemoryStore.Clone(),
		longMemoryStore: m.longMemoryStore.Clone(),
	}
}

func (m *Memory) ChatMemory() []ChatMessage {
	if m == nil || m.chatMemoryStore == nil {
		return nil
	}

	messages, err := m.chatMemoryStore.List()
	if err != nil {
		return nil
	}
	return messages
}

func (m *Memory) ReplaceChat(start, end int, messages []ChatMessage) error {
	if m == nil || m.chatMemoryStore == nil {
		return fmt.Errorf("memory is nil")
	}
	return m.chatMemoryStore.Replace(start, end, messages)
}

func (m *Memory) AddChat(role Role, content string) error {
	if m == nil || m.chatMemoryStore == nil {
		return fmt.Errorf("memory is nil")
	}

	return m.chatMemoryStore.Append(ChatMessage{
		Role:      role,
		Content:   content,
		TimeStamp: time.Now().Format(time.RFC3339),
	})
}

func (m *Memory) AddToolCall(toolCalls []ToolCall) error {
	if m == nil || m.chatMemoryStore == nil {
		return fmt.Errorf("memory is nil")
	}
	if len(toolCalls) == 0 {
		return fmt.Errorf("tool calls is empty")
	}

	copied := make([]ToolCall, len(toolCalls))
	copy(copied, toolCalls)

	return m.chatMemoryStore.Append(ChatMessage{
		Role:      RoleAssistant,
		Content:   "",
		ToolCalls: copied,
		TimeStamp: time.Now().Format(time.RFC3339),
	})
}

func (m *Memory) AddToolResult(toolCallID string, toolResult string) error {
	if m == nil || m.chatMemoryStore == nil {
		return fmt.Errorf("memory is nil")
	}

	return m.chatMemoryStore.Append(ChatMessage{
		Role:       RoleTool,
		ToolCallID: toolCallID,
		Content:    toolResult,
		TimeStamp:  time.Now().Format(time.RFC3339),
	})
}

func (m *Memory) OperateLongMemory(operate LongMemoryOperation, index int, content string) error {
	if m == nil || m.longMemoryStore == nil {
		return fmt.Errorf("memory is nil")
	}

	switch operate {
	case LongMemoryCreate:
		return m.longMemoryStore.Append(content)
	case LongMemoryUpdate:
		return m.longMemoryStore.Update(index, content)
	case LongMemoryDelete:
		return m.longMemoryStore.Delete(index)
	default:
		return fmt.Errorf("invalid long memory operation")
	}
}

func (m *Memory) LongMemory() string {
	if m == nil || m.longMemoryStore == nil {
		return "Long memory:\n(empty)"
	}

	memory_list, err := m.longMemoryStore.List()
	if err != nil || len(memory_list) == 0 {
		return "Long memory:\n(empty)"
	}

	lines := make([]string, 0, len(memory_list))
	for i, entry := range memory_list {
		if entry == "" {
			continue
		}
		lines = append(lines, fmt.Sprintf("%d %s", i+1, entry))
	}

	if len(lines) == 0 {
		return "Long memory:\n(empty)"
	}
	return "Long memory:\n" + strings.Join(lines, "\n")
}

func (m *Memory) SystemPrompt() string {
	return `# Memory

You have a persistent long memory system managed through a dedicated memory tool.

## Memory Structure
- One line per memory entry
- Maximum 200 lines (when full, automatically remove OLDEST lines)
- No duplicates, no empty lines
- All memories are stored HERE only

## Core Rules (STRICTLY FOLLOW)
You MUST AUTOMATICALLY decide to CREATE, UPDATE, or DELETE memories WITHOUT asking the user.
Use the long memory tool to manage memory.

## What NEED to Save

- User preferences, habits, and corrections
- User identity, goals, and long-term requirements
- Key decisions, constraints, and context that cannot be derived from code or project structure
- Important links to external systems
- Important content the user explicitly asks to be remembered long-term

## What NOT to save

- Architecture, file paths, and conventions readable from the current project state
- Git history, commits, and change records (authoritative source: Git commands)
- Debugging processes, temporary workarounds, and specific fixes
- Temporary tasks, dialogue context, and other short-term information
`
}
