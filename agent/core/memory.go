package core

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type MemoryStore interface {
	Save(messages []ChatMessage) error
	Load(messages *[]ChatMessage) error
}

type LongMemoryStore interface {
	Save(entries []string) error
	Load(entries *[]string) error
}

type InMemoryStore struct{}

func NewInMemoryStore() *InMemoryStore { return &InMemoryStore{} }

func (s *InMemoryStore) Save(messages []ChatMessage) error {
	return nil
}

func (s *InMemoryStore) Load(messages *[]ChatMessage) error {
	return nil
}

type InMemoryLongMemoryStore struct {
	mu      sync.RWMutex
	entries []string
}

func NewInMemoryLongMemoryStore() *InMemoryLongMemoryStore {
	return &InMemoryLongMemoryStore{entries: []string{}}
}

func (s *InMemoryLongMemoryStore) Save(entries []string) error {
	if s == nil {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries = append([]string(nil), entries...)
	return nil
}

func (s *InMemoryLongMemoryStore) Load(entries *[]string) error {
	if s == nil {
		return nil
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	*entries = append([]string(nil), s.entries...)
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

type FileLongMemoryStore struct {
	path string
}

func NewFileLongMemoryStore(path string) *FileLongMemoryStore {
	return &FileLongMemoryStore{path: path}
}

func (f *FileLongMemoryStore) Save(entries []string) error {
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
	for _, entry := range entries {
		if err := json.NewEncoder(writer).Encode(entry); err != nil {
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

func (f *FileLongMemoryStore) Load(entries *[]string) error {
	file, err := os.Open(f.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()

	loaded := make([]string, 0)
	reader := bufio.NewReader(file)
	for {
		line, err := reader.ReadBytes('\n')
		if len(line) > 0 {
			line = bytes.TrimSpace(line)
			if len(line) > 0 {
				var entry string
				if err := json.Unmarshal(line, &entry); err != nil {
					return err
				}
				if strings.TrimSpace(entry) != "" {
					loaded = append(loaded, entry)
				}
			}
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
	}
	*entries = loaded
	return nil
}

type Memory struct {
	mu              sync.RWMutex
	chatMemory      []ChatMessage
	longMemory      []string
	DialogIndex     []int
	DialogLimit     int
	memoryStore     MemoryStore
	longMemoryStore LongMemoryStore
}

func NewMemory() *Memory {
	return NewMemoryWithStores(NewInMemoryStore(), NewFileLongMemoryStore(".memory/MEMORY.md"))
}

func NewMemoryWithStore(store MemoryStore) *Memory {
	return NewMemoryWithStores(store, NewFileLongMemoryStore(".memory/MEMORY.md"))
}

func NewMemoryWithStores(store MemoryStore, longMemoryStore LongMemoryStore) *Memory {
	if store == nil {
		store = NewInMemoryStore()
	}
	if longMemoryStore == nil {
		longMemoryStore = NewInMemoryLongMemoryStore()
	}

	memory := &Memory{
		chatMemory:      []ChatMessage{},
		longMemory:      []string{},
		DialogIndex:     []int{},
		DialogLimit:     100,
		memoryStore:     store,
		longMemoryStore: longMemoryStore,
	}
	return memory
}

func NewFileBackedMemory(path string) *Memory {
	return NewMemoryWithStores(NewFileMemoryStore(path), NewFileLongMemoryStore(".memory/MEMORY.md"))
}

func NewFileBackedMemoryWithLongStore(path string, longMemoryStore LongMemoryStore) *Memory {
	return NewMemoryWithStores(NewFileMemoryStore(path), longMemoryStore)
}

func (m *Memory) Save() error {
	if m == nil {
		return nil
	}

	m.CheckDialogLimit()
	if m.memoryStore != nil {
		if err := m.memoryStore.Save(m.Snapshot()); err != nil {
			return err
		}
	}
	return m.SaveLongMemory()
}

func (m *Memory) Load() error {
	if m == nil {
		return nil
	}

	loaded := make([]ChatMessage, 0)
	if m.memoryStore != nil {
		if err := m.memoryStore.Load(&loaded); err != nil {
			return err
		}
	}

	m.mu.Lock()
	m.chatMemory = loaded
	m.rebuildDialogIndex()
	m.mu.Unlock()
	if err := m.LoadLongMemory(); err != nil {
		return err
	}
	m.CheckDialogLimit()
	return nil
}

func (m *Memory) LoadLongMemory() error {
	if m == nil {
		return nil
	}

	loaded := make([]string, 0)
	if m.longMemoryStore != nil {
		if err := m.longMemoryStore.Load(&loaded); err != nil {
			return err
		}
	}

	m.mu.Lock()
	m.longMemory = loaded
	m.mu.Unlock()
	return nil
}

func (m *Memory) SaveLongMemory() error {
	if m == nil {
		return nil
	}
	if m.longMemoryStore == nil {
		return nil
	}

	return m.longMemoryStore.Save(m.LongMemorySnapshot())
}

func (m *Memory) SetLongMemoryStore(store LongMemoryStore) {
	if m == nil {
		return
	}

	m.mu.Lock()
	if store == nil {
		store = NewInMemoryLongMemoryStore()
	}
	m.longMemoryStore = store
	m.mu.Unlock()
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
		DialogLimit:     m.DialogLimit,
		memoryStore:     m.memoryStore,
		longMemoryStore: m.longMemoryStore,
	}
	if len(m.chatMemory) > 0 {
		clone.chatMemory = append([]ChatMessage(nil), m.chatMemory...)
	} else {
		clone.chatMemory = []ChatMessage{}
	}
	if len(m.longMemory) > 0 {
		clone.longMemory = append([]string(nil), m.longMemory...)
	} else {
		clone.longMemory = []string{}
	}
	if len(m.DialogIndex) > 0 {
		clone.DialogIndex = append([]int(nil), m.DialogIndex...)
	} else {
		clone.DialogIndex = []int{}
	}
	return clone
}

func normalizeLongMemoryContent(content string) string {
	content = strings.ReplaceAll(content, "\r\n", " ")
	content = strings.ReplaceAll(content, "\n", " ")
	return strings.TrimSpace(content)
}

func (m *Memory) LongMemorySnapshot() []string {
	if m == nil {
		return []string{}
	}

	m.mu.RLock()
	defer m.mu.RUnlock()
	if len(m.longMemory) == 0 {
		return []string{}
	}
	return append([]string(nil), m.longMemory...)
}

func (m *Memory) ApplyLongMemoryOperation(operation string, index int, content string) (string, error) {
	if m == nil {
		return "", fmt.Errorf("memory is nil")
	}

	op := strings.ToLower(strings.TrimSpace(operation))
	content = normalizeLongMemoryContent(content)

	m.mu.Lock()
	switch op {
	case "create":
		if content == "" {
			m.mu.Unlock()
			return "", fmt.Errorf("context is required for create")
		}
		insertAt := len(m.longMemory)
		if index > 0 && index <= len(m.longMemory)+1 {
			insertAt = index - 1
		}
		m.longMemory = append(m.longMemory, "")
		copy(m.longMemory[insertAt+1:], m.longMemory[insertAt:])
		m.longMemory[insertAt] = content
		snapshot := append([]string(nil), m.longMemory...)
		m.mu.Unlock()
		if m.longMemoryStore != nil {
			if err := m.longMemoryStore.Save(snapshot); err != nil {
				return "", err
			}
		}
		return fmt.Sprintf("created long memory at index %d", insertAt+1), nil
	case "update":
		if index <= 0 || index > len(m.longMemory) {
			m.mu.Unlock()
			return "", fmt.Errorf("index out of range")
		}
		if content == "" {
			m.mu.Unlock()
			return "", fmt.Errorf("context is required for update")
		}
		m.longMemory[index-1] = content
		snapshot := append([]string(nil), m.longMemory...)
		m.mu.Unlock()
		if m.longMemoryStore != nil {
			if err := m.longMemoryStore.Save(snapshot); err != nil {
				return "", err
			}
		}
		return fmt.Sprintf("updated long memory at index %d", index), nil
	case "delete":
		if index <= 0 || index > len(m.longMemory) {
			m.mu.Unlock()
			return "", fmt.Errorf("index out of range")
		}
		m.longMemory = append(m.longMemory[:index-1], m.longMemory[index:]...)
		snapshot := append([]string(nil), m.longMemory...)
		m.mu.Unlock()
		if m.longMemoryStore != nil {
			if err := m.longMemoryStore.Save(snapshot); err != nil {
				return "", err
			}
		}
		return fmt.Sprintf("deleted long memory at index %d", index), nil
	default:
		m.mu.Unlock()
		return "", fmt.Errorf("unsupported operation: %s", operation)
	}
}

func (m *Memory) LongMemory() string {
	if m == nil {
		return "Long memory:\n(empty)"
	}

	snapshot := m.LongMemorySnapshot()
	if len(snapshot) == 0 {
		return "Long memory:\n(empty)"
	}

	lines := make([]string, 0, len(snapshot))
	for i, entry := range snapshot {
		entry = normalizeLongMemoryContent(entry)
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
	return `# Long Memory

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
