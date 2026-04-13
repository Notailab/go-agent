package storage

import (
	"errors"

	"github.com/Notailab/go-agent/agent/core"
)

type InMemoryLongStore struct {
	memories []string
}

func NewInMemoryLongStore() *InMemoryLongStore { return &InMemoryLongStore{} }

func (s *InMemoryLongStore) Get(index int) (string, error) {
	if s == nil {
		return "", errors.New("store is nil")
	}
	if index < 0 || index >= len(s.memories) {
		return "", errors.New("index out of bounds")
	}
	return s.memories[index], nil
}

func (s *InMemoryLongStore) Append(memory string) error {
	if s == nil {
		return errors.New("store is nil")
	}
	if s.memories == nil {
		s.memories = make([]string, 0, 4)
	}
	s.memories = append(s.memories, memory)
	return nil
}

func (s *InMemoryLongStore) Update(index int, memory string) error {
	if s == nil {
		return errors.New("store is nil")
	}
	if index < 0 || index >= len(s.memories) {
		return errors.New("index out of bounds")
	}
	s.memories[index] = memory
	return nil
}

func (s *InMemoryLongStore) Replace(start, end int, memories []string) error {
	if s == nil {
		return errors.New("store is nil")
	}
	if start < 0 || start > len(s.memories) {
		return errors.New("start index out of bounds")
	}
	if end < 0 || end > len(s.memories) {
		return errors.New("end index out of bounds")
	}
	if start > end {
		return errors.New("start index cannot be greater than end index")
	}
	s.memories = append(s.memories[:start], append(memories, s.memories[end:]...)...)
	return nil
}

func (s *InMemoryLongStore) Delete(index int) error {
	if s == nil {
		return errors.New("store is nil")
	}
	if index < 0 || index >= len(s.memories) {
		return errors.New("index out of bounds")
	}
	s.memories = append(s.memories[:index], s.memories[index+1:]...)
	return nil
}

func (s *InMemoryLongStore) List() ([]string, error) {
	if s == nil {
		return nil, errors.New("store is nil")
	}
	if s.memories == nil {
		return []string{}, nil
	}
	return append([]string(nil), s.memories...), nil
}

func (s *InMemoryLongStore) Count() (int, error) {
	if s == nil {
		return 0, errors.New("store is nil")
	}
	return len(s.memories), nil
}

func (s *InMemoryLongStore) Clear() error {
	if s == nil {
		return errors.New("store is nil")
	}
	s.memories = nil
	return nil
}

func (s *InMemoryLongStore) Clone() core.LongMemoryStore {
	if s == nil {
		return &InMemoryLongStore{memories: []string{}}
	}
	if s.memories == nil {
		return &InMemoryLongStore{memories: []string{}}
	}
	return &InMemoryLongStore{
		memories: append([]string(nil), s.memories...),
	}
}
