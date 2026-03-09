package inmemory

import (
	"context"
	"sync"

	"github.com/Karitham/corde"
	"github.com/karitham/waifubot/storage/interfaces"
)

type DropStore struct {
	mu    sync.RWMutex
	drops map[corde.Snowflake]interfaces.Drop
}

func NewDropStore() *DropStore {
	return &DropStore{
		drops: make(map[corde.Snowflake]interfaces.Drop),
	}
}

func (s *DropStore) Delete(ctx context.Context, id corde.Snowflake) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.drops, id)
	return nil
}

func (s *DropStore) Get(ctx context.Context, id corde.Snowflake) (*interfaces.Drop, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	drop, ok := s.drops[id]
	if !ok {
		return nil, ErrNotFound
	}
	return &drop, nil
}

func (s *DropStore) Set(ctx context.Context, id corde.Snowflake, data interfaces.Drop) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.drops[id] = data
	return nil
}
