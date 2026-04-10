package core

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type MemoryStore interface {
	Save(messages []ChatMessage) error
	Load(messages *[]ChatMessage) error
}

type InMemoryStore struct{}

func NewInMemoryStore() *InMemoryStore { return &InMemoryStore{} }

func (s *InMemoryStore) Save(messages []ChatMessage) error {
	return nil
}

func (s *InMemoryStore) Load(messages *[]ChatMessage) error {
	return nil
}

type FileMemoryStore struct {
	path string
}

func NewFileMemoryStore(path string) *FileMemoryStore { return &FileMemoryStore{path: path} }

func (f *FileMemoryStore) Save(messages []ChatMessage) error {
	dir := filepath.Dir(f.path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	tempFile, err := os.CreateTemp(dir, filepath.Base(f.path)+".tmp-*")
	if err != nil {
		return err
	}
	tempPath := tempFile.Name()
	defer func() {
		tempFile.Close()
		_ = os.Remove(tempPath)
	}()

	writer := bufio.NewWriter(tempFile)
	for _, m := range messages {
		if err := json.NewEncoder(writer).Encode(m); err != nil {
			return err
		}
	}
	if err := writer.Flush(); err != nil {
		return err
	}
	if err := tempFile.Sync(); err != nil {
		return err
	}
	if err := tempFile.Close(); err != nil {
		return err
	}
	if err := os.Rename(tempPath, f.path); err != nil {
		return err
	}
	return nil
}

func (f *FileMemoryStore) Load(messages *[]ChatMessage) error {
	file, err := os.Open(f.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()

	loaded := make([]ChatMessage, 0)
	reader := bufio.NewReader(file)
	for {
		line, err := reader.ReadBytes('\n')
		if len(line) > 0 {
			line = bytes.TrimSpace(line)
			if len(line) > 0 {
				var m ChatMessage
				if err := json.Unmarshal(line, &m); err != nil {
					return err
				}
				loaded = append(loaded, m)
			}
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
	}
	*messages = loaded
	return nil
}

type Memory struct {
	mu          sync.RWMutex
	chatMemory  []ChatMessage
	DialogIndex []int
	DialogLimit int
	memoryStore MemoryStore
}

func NewMemory() *Memory {
	return NewMemoryWithStore(NewInMemoryStore())
}

func NewMemoryWithStore(store MemoryStore) *Memory {
	if store == nil {
		store = NewInMemoryStore()
	}

	memory := &Memory{
		chatMemory:  []ChatMessage{},
		DialogIndex: []int{},
		DialogLimit: 100,
		memoryStore: store,
	}
	return memory
}

func NewFileBackedMemory(path string) *Memory {
	return NewMemoryWithStore(NewFileMemoryStore(path))
}

func (m *Memory) Save() error {
	if m == nil {
		return nil
	}

	snapshot := m.Snapshot()
	return m.memoryStore.Save(snapshot)
}

func (m *Memory) Load() error {
	if m == nil {
		return nil
	}

	loaded := make([]ChatMessage, 0)
	if err := m.memoryStore.Load(&loaded); err != nil {
		return err
	}

	m.mu.Lock()
	m.chatMemory = loaded
	m.rebuildDialogIndex()
	m.mu.Unlock()
	return nil
}

func (m *Memory) rebuildDialogIndex() {
	m.DialogIndex = m.DialogIndex[:0]
	for i, message := range m.chatMemory {
		if message.Role == RoleUser {
			m.DialogIndex = append(m.DialogIndex, i)
		}
	}
}

func (m *Memory) CheckDialogLimit() {
	if m == nil || m.DialogLimit <= 0 {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	excessDialogs := len(m.DialogIndex) - m.DialogLimit
	if excessDialogs <= 0 {
		return
	}

	cutIndex := len(m.chatMemory)
	if excessDialogs < len(m.DialogIndex) {
		cutIndex = m.DialogIndex[excessDialogs]
	}
	if cutIndex < 0 || cutIndex > len(m.chatMemory) {
		return
	}

	m.chatMemory = append([]ChatMessage(nil), m.chatMemory[cutIndex:]...)
	m.DialogIndex = append([]int(nil), m.DialogIndex[excessDialogs:]...)
	for i := range m.DialogIndex {
		m.DialogIndex[i] -= cutIndex
	}
}

func (m *Memory) SetDialogLimit(n int) {
	if m == nil {
		return
	}

	m.mu.Lock()
	m.DialogLimit = n
	m.mu.Unlock()
	m.CheckDialogLimit()
}

func (m *Memory) AddChat(role Role, content string) {
	if m == nil {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if role == RoleUser {
		m.DialogIndex = append(m.DialogIndex, len(m.chatMemory))
	}
	m.chatMemory = append(m.chatMemory, ChatMessage{
		Role:      role,
		Content:   content,
		TimeStamp: time.Now().Format(time.RFC3339),
	})
}

func (m *Memory) AddToolCall(toolCalls []ToolCall) {
	if len(toolCalls) == 0 {
		return
	}
	if m == nil {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	copied := make([]ToolCall, len(toolCalls))
	copy(copied, toolCalls)
	m.chatMemory = append(m.chatMemory, ChatMessage{
		Role:      RoleAssistant,
		Content:   "",
		ToolCalls: copied,
		TimeStamp: time.Now().Format(time.RFC3339),
	})
}

func (m *Memory) AddToolResult(toolCallID string, toolResult string) {
	if m == nil {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.chatMemory = append(m.chatMemory, ChatMessage{
		Role:       RoleTool,
		ToolCallID: toolCallID,
		Content:    toolResult,
		TimeStamp:  time.Now().Format(time.RFC3339),
	})
}

func (m *Memory) ChatMemory() []ChatMessage {
	if m == nil {
		return nil
	}

	m.mu.RLock()
	defer m.mu.RUnlock()
	return append([]ChatMessage(nil), m.chatMemory...)
}

func (m *Memory) Snapshot() []ChatMessage {
	if m == nil {
		return []ChatMessage{}
	}

	m.mu.RLock()
	defer m.mu.RUnlock()
	if len(m.chatMemory) == 0 {
		return []ChatMessage{}
	}

	return append([]ChatMessage(nil), m.chatMemory...)
}

func (m *Memory) Replace(start, end int, messages []ChatMessage) {
	if m == nil {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if start < 0 {
		start = 0
	}
	if end < start {
		end = start
	}
	if start > len(m.chatMemory) {
		start = len(m.chatMemory)
	}
	if end > len(m.chatMemory) {
		end = len(m.chatMemory)
	}

	replacement := append([]ChatMessage(nil), messages...)
	updated := make([]ChatMessage, 0, len(m.chatMemory)-(end-start)+len(replacement))
	updated = append(updated, m.chatMemory[:start]...)
	updated = append(updated, replacement...)
	updated = append(updated, m.chatMemory[end:]...)
	m.chatMemory = updated
	if len(m.chatMemory) == 0 {
		m.DialogIndex = []int{}
		return
	}
	m.rebuildDialogIndex()
}

func (m *Memory) Clone() *Memory {
	if m == nil {
		return nil
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	clone := &Memory{
		DialogLimit: m.DialogLimit,
		memoryStore: m.memoryStore,
	}
	if len(m.chatMemory) > 0 {
		clone.chatMemory = append([]ChatMessage(nil), m.chatMemory...)
	} else {
		clone.chatMemory = []ChatMessage{}
	}
	if len(m.DialogIndex) > 0 {
		clone.DialogIndex = append([]int(nil), m.DialogIndex...)
	} else {
		clone.DialogIndex = []int{}
	}
	return clone
}

func (m *Memory) LongMemory() string {
	b, err := os.ReadFile(".memory/MEMORY.md")
	if err != nil {
		return "Memory from .memory/MEMORY.md:\n(empty)"
	}
	return "Memory from .memory/MEMORY.md:\n" + string(b)
}

func (m *Memory) SystemPrompt() string {
	return `# Memory

You have a persistent file-based memory system stored in the .memory/MEMORY.md.

## Memory Structure
- MEMORY.md: ONE LINE = ONE COMPLETE MEMORY
- Maximum 200 lines (when full, automatically remove OLDEST lines)
- No duplicates, no empty lines
- All memories are stored HERE only

## Core Rules (STRICTLY FOLLOW)
You MUST AUTOMATICALLY decide to CREATE, UPDATE, or DELETE memories WITHOUT asking the user.
Use read_file, write_file, edit_file tools to manage memory.

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
