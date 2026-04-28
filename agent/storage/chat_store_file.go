package storage

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/Notailab/go-agent/agent/core"
)

func ensureParentDir(path string) error {
	dir := filepath.Dir(path)
	if dir == "." || dir == "" {
		return nil
	}
	return os.MkdirAll(dir, 0755)
}

type FileChatStore struct {
	path     string
	messages []core.ChatMessage
}

func NewFileChatStore(path string) *FileChatStore {
	store := &FileChatStore{path: path, messages: []core.ChatMessage{}}

	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return store
		}
		return store
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}
		var m core.ChatMessage
		if err := json.Unmarshal(line, &m); err != nil {
			continue
		}
		store.messages = append(store.messages, m)
	}
	return store
}

func (s *FileChatStore) Get(index int) (core.ChatMessage, error) {
	if s == nil || s.messages == nil {
		return core.ChatMessage{}, errors.New("messages is nil")
	}
	if index < 0 || index >= len(s.messages) {
		return core.ChatMessage{}, errors.New("index out of bounds")
	}
	return s.messages[index], nil
}

func (s *FileChatStore) Append(message core.ChatMessage) error {
	if s == nil || s.messages == nil {
		return errors.New("messages is nil")
	}
	s.messages = append(s.messages, message)
	if err := ensureParentDir(s.path); err != nil {
		return err
	}

	f, err := os.OpenFile(s.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	line, err := json.Marshal(message)
	if err != nil {
		return err
	}
	if _, err := f.Write(append(line, '\n')); err != nil {
		return err
	}
	return nil
}

func (s *FileChatStore) Update(index int, message core.ChatMessage) error {
	if s == nil || s.messages == nil {
		return errors.New("messages is nil")
	}
	if index < 0 || index >= len(s.messages) {
		return errors.New("index out of bounds")
	}
	s.messages[index] = message
	if err := ensureParentDir(s.path); err != nil {
		return err
	}
	f, err := os.OpenFile(s.path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	for _, m := range s.messages {
		line, err := json.Marshal(m)
		if err != nil {
			return err
		}
		if _, err := f.Write(append(line, '\n')); err != nil {
			return err
		}
	}
	return nil
}

func (s *FileChatStore) Replace(start, end int, messages []core.ChatMessage) error {
	if s == nil || s.messages == nil {
		return errors.New("messages is nil")
	}
	if start < 0 || end > len(s.messages) || start > end {
		return errors.New("index out of bounds")
	}
	s.messages = append(s.messages[:start], append(messages, s.messages[end:]...)...)
	if err := ensureParentDir(s.path); err != nil {
		return err
	}
	f, err := os.OpenFile(s.path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	for _, m := range s.messages {
		line, err := json.Marshal(m)
		if err != nil {
			return err
		}
		if _, err := f.Write(append(line, '\n')); err != nil {
			return err
		}
	}
	return nil
}

func (s *FileChatStore) Delete(index int) error {
	if s == nil || s.messages == nil {
		return errors.New("messages is nil")
	}
	if index < 0 || index >= len(s.messages) {
		return errors.New("index out of bounds")
	}
	s.messages = append(s.messages[:index], s.messages[index+1:]...)
	if err := ensureParentDir(s.path); err != nil {
		return err
	}
	f, err := os.OpenFile(s.path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	for _, m := range s.messages {
		line, err := json.Marshal(m)
		if err != nil {
			return err
		}
		if _, err := f.Write(append(line, '\n')); err != nil {
			return err
		}
	}
	return nil
}

func (s *FileChatStore) List() ([]core.ChatMessage, error) {
	if s == nil || s.messages == nil {
		return nil, errors.New("messages is nil")
	}
	return s.messages, nil
}

func (s *FileChatStore) Count() (int, error) {
	if s == nil || s.messages == nil {
		return 0, errors.New("messages is nil")
	}
	return len(s.messages), nil
}

func (s *FileChatStore) Clear() error {
	if s == nil || s.messages == nil {
		return errors.New("messages is nil")
	}
	s.messages = nil
	if err := ensureParentDir(s.path); err != nil {
		return err
	}
	f, err := os.OpenFile(s.path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	return nil
}

func (s *FileChatStore) Clone() core.ChatMemoryStore {
	if s == nil || s.messages == nil {
		return &FileChatStore{}
	}
	clone := &FileChatStore{path: s.path, messages: []core.ChatMessage{}}
	if len(s.messages) == 0 {
		return clone
	}

	clone.messages = make([]core.ChatMessage, 0, len(s.messages))
	for _, m := range s.messages {
		nm := m
		if m.ToolCalls != nil {
			nm.ToolCalls = make([]core.ToolCall, len(m.ToolCalls))
			copy(nm.ToolCalls, m.ToolCalls)
		}
		clone.messages = append(clone.messages, nm)
	}
	return clone
}
