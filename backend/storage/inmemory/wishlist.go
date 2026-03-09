package inmemory

import (
	"context"
	"sync"

	"github.com/karitham/waifubot/storage/interfaces"
)

type WishlistStore struct {
	mu        sync.RWMutex
	wishlists map[uint64]map[int64]struct{}
}

func NewWishlistStore() *WishlistStore {
	return &WishlistStore{
		wishlists: make(map[uint64]map[int64]struct{}),
	}
}

func (s *WishlistStore) AddCharacter(ctx context.Context, userID uint64, charID int64) error {
	return s.AddMultipleCharacters(ctx, userID, []int64{charID})
}

func (s *WishlistStore) AddMultipleCharacters(ctx context.Context, userID uint64, charIDs []int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.wishlists[userID] == nil {
		s.wishlists[userID] = make(map[int64]struct{})
	}

	for _, charID := range charIDs {
		s.wishlists[userID][charID] = struct{}{}
	}
	return nil
}

func (s *WishlistStore) RemoveCharacter(ctx context.Context, userID uint64, charID int64) error {
	return s.RemoveMultipleCharacters(ctx, userID, []int64{charID})
}

func (s *WishlistStore) RemoveMultipleCharacters(ctx context.Context, userID uint64, charIDs []int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.wishlists[userID] == nil {
		return nil
	}

	for _, charID := range charIDs {
		delete(s.wishlists[userID], charID)
	}
	return nil
}

func (s *WishlistStore) RemoveAll(ctx context.Context, userID uint64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.wishlists, userID)
	return nil
}

func (s *WishlistStore) GetUserWishlist(ctx context.Context, userID uint64) ([]int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	userWL, ok := s.wishlists[userID]
	if !ok {
		return []int64{}, nil
	}

	ids := make([]int64, 0, len(userWL))
	for id := range userWL {
		ids = append(ids, id)
	}
	return ids, nil
}

func (s *WishlistStore) GetWantedCharacters(ctx context.Context, charID int64) ([]uint64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var owners []uint64
	for userID, wishlist := range s.wishlists {
		if _, wants := wishlist[charID]; wants {
			owners = append(owners, userID)
		}
	}
	return owners, nil
}

func (s *WishlistStore) GetWishlistHolders(ctx context.Context, charIDs []int64) (map[int64][]uint64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[int64][]uint64)
	charSet := make(map[int64]struct{})
	for _, id := range charIDs {
		charSet[id] = struct{}{}
	}

	for userID, wishlist := range s.wishlists {
		for charID := range wishlist {
			if _, ok := charSet[charID]; ok {
				result[charID] = append(result[charID], userID)
			}
		}
	}
	return result, nil
}

func (s *WishlistStore) CompareWithUser(ctx context.Context, userID uint64, otherUserID uint64) (interfaces.WishlistComparison, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	userWL := s.wishlists[userID]
	otherWL := s.wishlists[otherUserID]

	userHas := make([]int64, 0)
	userWants := make([]int64, 0)
	mutual := make([]int64, 0)

	for charID := range otherWL {
		userHas = append(userHas, charID)
	}

	for charID := range userWL {
		userWants = append(userWants, charID)
		if _, ok := otherWL[charID]; ok {
			mutual = append(mutual, charID)
		}
	}

	return interfaces.WishlistComparison{
		UserHas:       userHas,
		UserWants:     userWants,
		MutualMatches: mutual,
	}, nil
}
