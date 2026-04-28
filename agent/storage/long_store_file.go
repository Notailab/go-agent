package storage

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"os"

	"github.com/Notailab/go-agent/agent/core"
)

type FileLongStore struct {
	path     string
	memories []string
}

func NewFileLongStore(path string) *FileLongStore {
	store := &FileLongStore{path: path, memories: []string{}}

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
		var m string
		if err := json.Unmarshal(line, &m); err != nil {
			continue
		}
		store.memories = append(store.memories, m)
	}
	return store
}

func (s *FileLongStore) Get(index int) (string, error) {
	if s == nil || s.memories == nil {
		return "", errors.New("memories is nil")
	}
	if index < 0 || index >= len(s.memories) {
		return "", errors.New("index out of bounds")
	}
	return s.memories[index], nil
}

func (s *FileLongStore) Append(memory string) error {
	if s == nil || s.memories == nil {
		return errors.New("memories is nil")
	}
	s.memories = append(s.memories, memory)
	if err := ensureParentDir(s.path); err != nil {
		return err
	}

	f, err := os.OpenFile(s.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	line, err := json.Marshal(memory)
	if err != nil {
		return err
	}
	if _, err := f.Write(append(line, '\n')); err != nil {
		return err
	}
	return nil
}

func (s *FileLongStore) Update(index int, memory string) error {
	if s == nil || s.memories == nil {
		return errors.New("memories is nil")
	}
	if index < 0 || index >= len(s.memories) {
		return errors.New("index out of bounds")
	}
	s.memories[index] = memory
	if err := ensureParentDir(s.path); err != nil {
		return err
	}
	f, err := os.OpenFile(s.path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	for _, m := range s.memories {
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

func (s *FileLongStore) Replace(start, end int, memories []string) error {
	if s == nil || s.memories == nil {
		return errors.New("memories is nil")
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
	if err := ensureParentDir(s.path); err != nil {
		return err
	}
	f, err := os.OpenFile(s.path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	for _, m := range s.memories {
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

func (s *FileLongStore) Delete(index int) error {
	if s == nil || s.memories == nil {
		return errors.New("memories is nil")
	}
	if index < 0 || index >= len(s.memories) {
		return errors.New("index out of bounds")
	}
	s.memories = append(s.memories[:index], s.memories[index+1:]...)
	if err := ensureParentDir(s.path); err != nil {
		return err
	}
	f, err := os.OpenFile(s.path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	for _, m := range s.memories {
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

func (s *FileLongStore) List() ([]string, error) {
	if s == nil || s.memories == nil {
		return nil, errors.New("memories is nil")
	}
	return s.memories, nil
}

func (s *FileLongStore) Count() (int, error) {
	if s == nil || s.memories == nil {
		return 0, errors.New("memories is nil")
	}
	return len(s.memories), nil
}

func (s *FileLongStore) Clear() error {
	if s == nil || s.memories == nil {
		return errors.New("memories is nil")
	}
	s.memories = nil
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

func (s *FileLongStore) Clone() core.LongMemoryStore {
	if s == nil || s.memories == nil {
		return &FileLongStore{}
	}
	clone := &FileLongStore{path: s.path, memories: []string{}}
	if len(s.memories) == 0 {
		return clone
	}

	clone.memories = make([]string, 0, len(s.memories))
	for _, m := range s.memories {
		clone.memories = append(clone.memories, m)
	}
	return clone
}
