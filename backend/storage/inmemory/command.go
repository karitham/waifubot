package inmemory

import (
	"context"
	"sync"
)

type CommandStore struct {
	mu   sync.RWMutex
	hash string
}

func NewCommandStore() *CommandStore {
	return &CommandStore{}
}

func (s *CommandStore) GetCommandHash(ctx context.Context) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.hash, nil
}

func (s *CommandStore) SetCommandHash(ctx context.Context, hash string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.hash = hash
	return nil
}

func (s *CommandStore) UpdateCommandHash(ctx context.Context, hash string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.hash = hash
	return nil
}
