package wishlisttest

import (
	"context"

	"github.com/karitham/waifubot/wishlist"
)

// MockStore implements wishlist.Store for testing.
type MockStore struct {
	AddCharactersToWishlistFunc      func(ctx context.Context, userID uint64, characterIDs []int64) error
	RemoveCharactersFromWishlistFunc func(ctx context.Context, userID uint64, characterIDs []int64) error
	RemoveAllFromWishlistFunc        func(ctx context.Context, userID uint64) error
	GetUserCharacterWishlistFunc     func(ctx context.Context, userID uint64) ([]wishlist.Character, error)
	GetWishlistHoldersFunc           func(ctx context.Context, characterIDs []int64, userID, guildID uint64) ([]wishlist.UserCharacterSet, error)
	GetWantedCharactersFunc          func(ctx context.Context, userID, guildID uint64) ([]wishlist.UserCharacterSet, error)
	CompareWithUserFunc              func(ctx context.Context, userID1, userID2 uint64) (wishlist.WishlistComparison, error)
	GetUsersWantingCharacterFunc     func(ctx context.Context, charID int64, guildID, excludeUserID uint64) ([]uint64, error)
}

var _ wishlist.Store = (*MockStore)(nil)

func (m *MockStore) AddCharactersToWishlist(ctx context.Context, userID uint64, characterIDs []int64) error {
	if m.AddCharactersToWishlistFunc != nil {
		return m.AddCharactersToWishlistFunc(ctx, userID, characterIDs)
	}
	return nil
}

func (m *MockStore) RemoveCharactersFromWishlist(ctx context.Context, userID uint64, characterIDs []int64) error {
	if m.RemoveCharactersFromWishlistFunc != nil {
		return m.RemoveCharactersFromWishlistFunc(ctx, userID, characterIDs)
	}
	return nil
}

func (m *MockStore) RemoveAllFromWishlist(ctx context.Context, userID uint64) error {
	if m.RemoveAllFromWishlistFunc != nil {
		return m.RemoveAllFromWishlistFunc(ctx, userID)
	}
	return nil
}

func (m *MockStore) GetUserCharacterWishlist(ctx context.Context, userID uint64) ([]wishlist.Character, error) {
	if m.GetUserCharacterWishlistFunc != nil {
		return m.GetUserCharacterWishlistFunc(ctx, userID)
	}
	return nil, nil
}

func (m *MockStore) GetWishlistHolders(ctx context.Context, characterIDs []int64, userID, guildID uint64) ([]wishlist.UserCharacterSet, error) {
	if m.GetWishlistHoldersFunc != nil {
		return m.GetWishlistHoldersFunc(ctx, characterIDs, userID, guildID)
	}
	return nil, nil
}

func (m *MockStore) GetWantedCharacters(ctx context.Context, userID, guildID uint64) ([]wishlist.UserCharacterSet, error) {
	if m.GetWantedCharactersFunc != nil {
		return m.GetWantedCharactersFunc(ctx, userID, guildID)
	}
	return nil, nil
}

func (m *MockStore) CompareWithUser(ctx context.Context, userID1, userID2 uint64) (wishlist.WishlistComparison, error) {
	if m.CompareWithUserFunc != nil {
		return m.CompareWithUserFunc(ctx, userID1, userID2)
	}
	return wishlist.WishlistComparison{}, nil
}

func (m *MockStore) GetUsersWantingCharacter(ctx context.Context, charID int64, guildID, excludeUserID uint64) ([]uint64, error) {
	if m.GetUsersWantingCharacterFunc != nil {
		return m.GetUsersWantingCharacterFunc(ctx, charID, guildID, excludeUserID)
	}
	return nil, nil
}
