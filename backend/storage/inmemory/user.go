package inmemory

import (
	"context"
	"sync"
	"time"

	"github.com/karitham/waifubot/storage/interfaces"
)

type UserStore struct {
	mu    sync.RWMutex
	users map[uint64]interfaces.User
}

func NewUserStore() *UserStore {
	return &UserStore{
		users: make(map[uint64]interfaces.User),
	}
}

func (s *UserStore) Create(ctx context.Context, userID uint64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.users[userID]; exists {
		return nil
	}

	s.users[userID] = interfaces.User{
		ID:     uint64(len(s.users) + 1),
		UserID: userID,
		Date:   time.Now(),
		Tokens: 0,
	}
	return nil
}

func (s *UserStore) Get(ctx context.Context, userID uint64) (interfaces.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, ok := s.users[userID]
	if !ok {
		return interfaces.User{}, ErrNotFound
	}
	return user, nil
}

func (s *UserStore) GetByAnilist(ctx context.Context, url string) (interfaces.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, user := range s.users {
		if user.AnilistURL == url {
			return user, nil
		}
	}
	return interfaces.User{}, ErrNotFound
}

func (s *UserStore) GetByDiscordUsername(ctx context.Context, username string) (interfaces.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, user := range s.users {
		if user.DiscordUsername == username {
			return user, nil
		}
	}
	return interfaces.User{}, ErrNotFound
}

func (s *UserStore) List(ctx context.Context, limit, offset int) ([]interfaces.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	users := make([]interfaces.User, 0, len(s.users))
	for _, user := range s.users {
		users = append(users, user)
	}

	if offset >= len(users) {
		return []interfaces.User{}, nil
	}
	if offset+limit > len(users) {
		limit = len(users) - offset
	}
	return users[offset : offset+limit], nil
}

func (s *UserStore) UpdateTokens(ctx context.Context, userID uint64, tokens int) (interfaces.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, ok := s.users[userID]
	if !ok {
		return interfaces.User{}, ErrNotFound
	}
	user.Tokens = tokens
	user.LastUpdated = time.Now()
	s.users[userID] = user
	return user, nil
}

func (s *UserStore) UpdateQuote(ctx context.Context, userID uint64, quote string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, ok := s.users[userID]
	if !ok {
		return ErrNotFound
	}
	user.Quote = quote
	user.LastUpdated = time.Now()
	s.users[userID] = user
	return nil
}

func (s *UserStore) UpdateFavorite(ctx context.Context, userID uint64, favorite int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, ok := s.users[userID]
	if !ok {
		return ErrNotFound
	}
	user.Favorite = favorite
	user.LastUpdated = time.Now()
	s.users[userID] = user
	return nil
}

func (s *UserStore) UpdateAnilistURL(ctx context.Context, userID uint64, url string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, ok := s.users[userID]
	if !ok {
		return ErrNotFound
	}
	user.AnilistURL = url
	user.LastUpdated = time.Now()
	s.users[userID] = user
	return nil
}

func (s *UserStore) UpdateDiscordInfo(ctx context.Context, userID uint64, username, avatar string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, ok := s.users[userID]
	if !ok {
		return ErrNotFound
	}
	user.DiscordUsername = username
	user.DiscordAvatar = avatar
	user.LastUpdated = time.Now()
	s.users[userID] = user
	return nil
}

func (s *UserStore) CountFiltered(ctx context.Context, search string) (int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return int64(len(s.users)), nil
}
