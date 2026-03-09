package inmemory

import (
	"context"
	"sync"

	"github.com/Karitham/corde"
)

type InteractionStore struct {
	mu           sync.RWMutex
	interactions map[corde.Snowflake]int64
}

func NewInteractionStore() *InteractionStore {
	return &InteractionStore{
		interactions: make(map[corde.Snowflake]int64),
	}
}

func (s *InteractionStore) Increment(ctx context.Context, channelID corde.Snowflake) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.interactions[channelID]++
	return nil
}

func (s *InteractionStore) Get(ctx context.Context, channelID corde.Snowflake) (int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.interactions[channelID], nil
}

func (s *InteractionStore) Reset(ctx context.Context, channelID corde.Snowflake) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.interactions[channelID] = 0
	return nil
}
