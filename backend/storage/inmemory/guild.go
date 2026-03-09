package inmemory

import (
	"context"
	"sync"
	"time"

	"github.com/karitham/waifubot/storage/interfaces"
)

type GuildStore struct {
	mu             sync.RWMutex
	members        map[uint64]map[uint64]struct{}
	indexingStatus map[uint64]interfaces.IndexStatus
}

func NewGuildStore() *GuildStore {
	return &GuildStore{
		members:        make(map[uint64]map[uint64]struct{}),
		indexingStatus: make(map[uint64]interfaces.IndexStatus),
	}
}

func (s *GuildStore) CompleteIndexingJob(ctx context.Context, guildID uint64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.indexingStatus[guildID] = interfaces.IndexStatus{
		Status:    interfaces.IndexingStatusCompleted,
		UpdatedAt: time.Now(),
	}
	return nil
}

func (s *GuildStore) DeleteGuildMembers(ctx context.Context, guildID uint64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.members, guildID)
	return nil
}

func (s *GuildStore) DeleteGuildMembersNotIn(ctx context.Context, guildID uint64, memberIDs []uint64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	guildMembers, ok := s.members[guildID]
	if !ok {
		return nil
	}

	memberSet := make(map[uint64]struct{})
	for _, id := range memberIDs {
		memberSet[id] = struct{}{}
	}

	for id := range guildMembers {
		if _, keep := memberSet[id]; !keep {
			delete(guildMembers, id)
		}
	}
	return nil
}

func (s *GuildStore) GetGuildMembers(ctx context.Context, guildID uint64) ([]uint64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	guildMembers, ok := s.members[guildID]
	if !ok {
		return []uint64{}, nil
	}

	ids := make([]uint64, 0, len(guildMembers))
	for id := range guildMembers {
		ids = append(ids, id)
	}
	return ids, nil
}

func (s *GuildStore) GetIndexingStatus(ctx context.Context, guildID uint64) (interfaces.IndexStatus, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	status, ok := s.indexingStatus[guildID]
	if !ok {
		return interfaces.IndexStatus{
			Status:    interfaces.IndexingStatusPending,
			UpdatedAt: time.Now(),
		}, nil
	}
	return status, nil
}

func (s *GuildStore) IsGuildIndexed(ctx context.Context, guildID uint64) (bool, time.Time, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	status, ok := s.indexingStatus[guildID]
	if !ok {
		return false, time.Time{}, nil
	}

	return status.Status == interfaces.IndexingStatusCompleted, status.UpdatedAt, nil
}

func (s *GuildStore) StartIndexingJob(ctx context.Context, guildID uint64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.indexingStatus[guildID] = interfaces.IndexStatus{
		Status:    interfaces.IndexingStatusRunning,
		UpdatedAt: time.Now(),
	}
	return nil
}

func (s *GuildStore) UpsertGuildMembers(ctx context.Context, guildID uint64, memberIDs []uint64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.members[guildID] == nil {
		s.members[guildID] = make(map[uint64]struct{})
	}

	for _, id := range memberIDs {
		s.members[guildID][id] = struct{}{}
	}
	return nil
}

func (s *GuildStore) UsersOwningCharInGuild(ctx context.Context, guildID uint64, charID int64) ([]uint64, error) {
	return []uint64{}, nil
}
