package storage

import (
	"errors"

	"github.com/Notailab/go-agent/agent/core"
)

type InMemoryChatStore struct {
	messages []core.ChatMessage
}

func NewInMemoryChatStore() *InMemoryChatStore { return &InMemoryChatStore{} }

func (s *InMemoryChatStore) Get(index int) (core.ChatMessage, error) {
	if s == nil {
		return core.ChatMessage{}, errors.New("store is nil")
	}
	if index < 0 || index >= len(s.messages) {
		return core.ChatMessage{}, errors.New("index out of bounds")
	}
	return s.messages[index], nil
}

func (s *InMemoryChatStore) Append(message core.ChatMessage) error {
	if s == nil {
		return errors.New("store is nil")
	}
	if s.messages == nil {
		s.messages = make([]core.ChatMessage, 0, 4)
	}
	s.messages = append(s.messages, message)
	return nil
}

func (s *InMemoryChatStore) Update(index int, message core.ChatMessage) error {
	if s == nil {
		return errors.New("store is nil")
	}
	if index < 0 || index >= len(s.messages) {
		return errors.New("index out of bounds")
	}
	s.messages[index] = message
	return nil
}

func (s *InMemoryChatStore) Replace(start, end int, messages []core.ChatMessage) error {
	if s == nil {
		return errors.New("store is nil")
	}
	if start < 0 || start > len(s.messages) {
		return errors.New("start index out of bounds")
	}
	if end < 0 || end > len(s.messages) {
		return errors.New("end index out of bounds")
	}
	if start > end {
		return errors.New("start index cannot be greater than end index")
	}
	s.messages = append(s.messages[:start], append(messages, s.messages[end:]...)...)
	return nil
}

func (s *InMemoryChatStore) Delete(index int) error {
	if s == nil {
		return errors.New("store is nil")
	}
	if index < 0 || index >= len(s.messages) {
		return errors.New("index out of bounds")
	}
	s.messages = append(s.messages[:index], s.messages[index+1:]...)
	return nil
}

func (s *InMemoryChatStore) List() ([]core.ChatMessage, error) {
	if s == nil {
		return nil, errors.New("store is nil")
	}
	if s.messages == nil {
		return []core.ChatMessage{}, nil
	}
	return append([]core.ChatMessage(nil), s.messages...), nil
}

func (s *InMemoryChatStore) Count() (int, error) {
	if s == nil {
		return 0, errors.New("store is nil")
	}
	return len(s.messages), nil
}

func (s *InMemoryChatStore) Clear() error {
	if s == nil {
		return errors.New("store is nil")
	}
	s.messages = nil
	return nil
}

func (s *InMemoryChatStore) Clone() core.ChatMemoryStore {
	if s == nil {
		return &InMemoryChatStore{messages: []core.ChatMessage{}}
	}
	if s.messages == nil {
		return &InMemoryChatStore{messages: []core.ChatMessage{}}
	}
	return &InMemoryChatStore{
		messages: append([]core.ChatMessage(nil), s.messages...),
	}
}
