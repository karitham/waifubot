package discord

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/Karitham/corde"
	"github.com/failsafe-go/failsafe-go/bulkhead"
	"golang.org/x/sync/singleflight"

	"github.com/karitham/waifubot/collection"
)

const (
	refreshTTL             = 24 * time.Hour
	refreshTimeout         = 5 * time.Second
	refreshFailureTTL      = 5 * time.Minute
	maxConcurrentRefreshes = 8
	maxRefreshFailures     = 4096
)

type fetchDiscordUser func(ctx context.Context, userID string) (*DiscordUser, error)

// Service refreshes stored Discord user information.
type Service struct {
	fetchUser fetchDiscordUser
	permits   bulkhead.Bulkhead[struct{}]
	group     singleflight.Group

	mu       sync.Mutex
	failures map[corde.Snowflake]time.Time
}

func NewService(botToken string) *Service {
	client := NewClient(botToken)
	return newService(client.GetUser)
}

func newService(fetchUser fetchDiscordUser) *Service {
	return &Service{
		fetchUser: fetchUser,
		permits:   bulkhead.New[struct{}](maxConcurrentRefreshes),
		failures:  make(map[corde.Snowflake]time.Time),
	}
}

func (s *Service) updateUserInfo(ctx context.Context, store collection.DB, userID corde.Snowflake) error {
	if userID == 0 {
		return collection.ErrInvalidUserID
	}

	discordUser, err := s.fetchUser(ctx, fmt.Sprintf("%d", uint64(userID)))
	if err != nil {
		return fmt.Errorf("failed to fetch Discord user: %w", err)
	}

	return collection.SetDiscordInfo(ctx, store, collection.UserID(userID), discordUser.Username, discordUser.Avatar, time.Now())
}

func (s *Service) UpdateIfNeeded(ctx context.Context, store collection.DB, userID corde.Snowflake) error {
	if userID == 0 {
		return collection.ErrInvalidUserID
	}

	key := strconv.FormatUint(uint64(userID), 10)
	_, err, _ := s.group.Do(key, func() (any, error) {
		// The shared refresh owns the whole flight's deadline: a client
		// disconnect must not cancel the freshness check, the semaphore wait,
		// or the fetch, nor poison the failure cache for other readers.
		refreshCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), refreshTimeout)
		defer cancel()

		if s.recentFailure(userID) {
			return nil, nil
		}

		shouldUpdate, err := shouldUpdateInfo(refreshCtx, store, userID)
		if err != nil {
			if errors.Is(err, collection.ErrNotFound) {
				return nil, nil
			}
			return nil, err
		}
		if !shouldUpdate {
			return nil, nil
		}

		if err := s.permits.AcquirePermit(refreshCtx); err != nil {
			return nil, err
		}
		defer s.permits.ReleasePermit()

		if err := s.updateUserInfo(refreshCtx, store, userID); err != nil {
			s.recordFailure(userID)
			return nil, err
		}
		s.clearFailure(userID)
		return nil, nil
	})
	return err
}

// shouldUpdateInfo reports whether Discord info is older than 24 hours.
func shouldUpdateInfo(ctx context.Context, store collection.Store, userID corde.Snowflake) (bool, error) {
	u, err := store.GetUser(ctx, collection.UserID(userID))
	if err != nil {
		return false, err
	}

	return u.LastUpdated.IsZero() || time.Since(u.LastUpdated) > refreshTTL, nil
}

func (s *Service) recentFailure(userID corde.Snowflake) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	at, ok := s.failures[userID]
	if !ok {
		return false
	}
	if time.Since(at) > refreshFailureTTL {
		delete(s.failures, userID)
		return false
	}
	return true
}

func (s *Service) recordFailure(userID corde.Snowflake) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	if len(s.failures) >= maxRefreshFailures {
		s.pruneFailuresLocked(now)
		if len(s.failures) >= maxRefreshFailures {
			clear(s.failures)
		}
	}
	s.failures[userID] = now
}

func (s *Service) pruneFailuresLocked(now time.Time) {
	for id, at := range s.failures {
		if now.Sub(at) > refreshFailureTTL {
			delete(s.failures, id)
		}
	}
}

func (s *Service) clearFailure(userID corde.Snowflake) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.failures, userID)
}
