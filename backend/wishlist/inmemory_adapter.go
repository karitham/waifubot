package wishlist

import (
	"context"

	"github.com/karitham/waifubot/storage/inmemory"
)

type InMemoryStore struct {
	wishlist *inmemory.WishlistStore
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		wishlist: inmemory.NewWishlistStore(),
	}
}

func (s *InMemoryStore) AddMultipleCharactersToWishlist(ctx context.Context, userID uint64, characterIDs []int64) error {
	return s.wishlist.AddMultipleCharacters(ctx, userID, characterIDs)
}

func (s *InMemoryStore) RemoveMultipleCharactersFromWishlist(ctx context.Context, userID uint64, characterIDs []int64) error {
	return s.wishlist.RemoveMultipleCharacters(ctx, userID, characterIDs)
}

func (s *InMemoryStore) RemoveAllFromWishlist(ctx context.Context, userID uint64) error {
	return s.wishlist.RemoveAll(ctx, userID)
}

func (s *InMemoryStore) GetUserCharacterWishlist(ctx context.Context, userID uint64) ([]Character, error) {
	ids, err := s.wishlist.GetUserWishlist(ctx, userID)
	if err != nil {
		return nil, err
	}

	characters := make([]Character, len(ids))
	for i, id := range ids {
		characters[i] = Character{ID: id}
	}
	return characters, nil
}

func (s *InMemoryStore) GetWishlistHolders(ctx context.Context, userID, guildID uint64) ([]WishlistHolder, error) {
	return nil, nil
}

func (s *InMemoryStore) GetWantedCharacters(ctx context.Context, userID, guildID uint64) ([]WantedCharacter, error) {
	return nil, nil
}

func (s *InMemoryStore) CompareWithUser(ctx context.Context, userID1, userID2 uint64) (WishlistComparison, error) {
	comp, err := s.wishlist.CompareWithUser(ctx, userID1, userID2)
	if err != nil {
		return WishlistComparison{}, err
	}

	return WishlistComparison{
		UserHasCharacters:   charsFromIDs(comp.UserHas),
		UserWantsCharacters: charsFromIDs(comp.UserWants),
		MutualMatches:       len(comp.MutualMatches),
	}, nil
}

func charsFromIDs(ids []int64) []Character {
	result := make([]Character, len(ids))
	for i, id := range ids {
		result[i] = Character{ID: id}
	}
	return result
}
