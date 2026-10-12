package wishlisttest

import (
	"context"

	"github.com/karitham/waifubot/catalog"
	"github.com/karitham/waifubot/collection"
	"github.com/karitham/waifubot/guild"
	"github.com/karitham/waifubot/wishlist"
)

// MockStore implements wishlist.Store for testing.
type MockStore struct {
	RemoveCharactersFromWishlistFunc func(ctx context.Context, userID collection.UserID, characterIDs []catalog.CharacterID) error
	RemoveAllFromWishlistFunc        func(ctx context.Context, userID collection.UserID) error
	GetUserCharacterWishlistFunc     func(ctx context.Context, userID collection.UserID) ([]wishlist.Entry, error)
	GetWishlistHoldersFunc           func(ctx context.Context, characterIDs []catalog.CharacterID, userID collection.UserID, guildID guild.ID) ([]wishlist.UserCharacterSet, error)
	GetWantedCharactersFunc          func(ctx context.Context, userID collection.UserID, guildID guild.ID) ([]wishlist.UserCharacterSet, error)
	CompareWithUserFunc              func(ctx context.Context, userID1, userID2 collection.UserID) (wishlist.WishlistComparison, error)
	GetUsersWantingCharacterFunc     func(ctx context.Context, charID catalog.CharacterID, guildID guild.ID, excludeUserID collection.UserID) ([]guild.MemberID, error)
}

var _ wishlist.Store = (*MockStore)(nil)

func (m *MockStore) RemoveCharactersFromWishlist(ctx context.Context, userID collection.UserID, characterIDs []catalog.CharacterID) error {
	if m.RemoveCharactersFromWishlistFunc != nil {
		return m.RemoveCharactersFromWishlistFunc(ctx, userID, characterIDs)
	}
	return nil
}

func (m *MockStore) RemoveAllFromWishlist(ctx context.Context, userID collection.UserID) error {
	if m.RemoveAllFromWishlistFunc != nil {
		return m.RemoveAllFromWishlistFunc(ctx, userID)
	}
	return nil
}

func (m *MockStore) GetUserCharacterWishlist(ctx context.Context, userID collection.UserID) ([]wishlist.Entry, error) {
	if m.GetUserCharacterWishlistFunc != nil {
		return m.GetUserCharacterWishlistFunc(ctx, userID)
	}
	return nil, nil
}

func (m *MockStore) GetWishlistHolders(ctx context.Context, characterIDs []catalog.CharacterID, userID collection.UserID, guildID guild.ID) ([]wishlist.UserCharacterSet, error) {
	if m.GetWishlistHoldersFunc != nil {
		return m.GetWishlistHoldersFunc(ctx, characterIDs, userID, guildID)
	}
	return nil, nil
}

func (m *MockStore) GetWantedCharacters(ctx context.Context, userID collection.UserID, guildID guild.ID) ([]wishlist.UserCharacterSet, error) {
	if m.GetWantedCharactersFunc != nil {
		return m.GetWantedCharactersFunc(ctx, userID, guildID)
	}
	return nil, nil
}

func (m *MockStore) CompareWithUser(ctx context.Context, userID1, userID2 collection.UserID) (wishlist.WishlistComparison, error) {
	if m.CompareWithUserFunc != nil {
		return m.CompareWithUserFunc(ctx, userID1, userID2)
	}
	return wishlist.WishlistComparison{}, nil
}

func (m *MockStore) GetUsersWantingCharacter(ctx context.Context, charID catalog.CharacterID, guildID guild.ID, excludeUserID collection.UserID) ([]guild.MemberID, error) {
	if m.GetUsersWantingCharacterFunc != nil {
		return m.GetUsersWantingCharacterFunc(ctx, charID, guildID, excludeUserID)
	}
	return nil, nil
}
