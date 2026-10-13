package wishlisttest

import (
	"context"

	"github.com/karitham/waifubot/storage"
)

// MockQuerier embeds the generated querier and overrides the wishlist
// methods the adapter tests exercise.
type MockQuerier struct {
	storage.Querier
	AddCharactersToWishlistFunc      func(ctx context.Context, arg storage.AddCharactersToWishlistParams) (int64, error)
	CompareWithUserFunc              func(ctx context.Context, arg storage.CompareWithUserParams) ([]storage.CompareWithUserRow, error)
	GetUserCharacterWishlistFunc     func(ctx context.Context, userID uint64) ([]storage.GetUserCharacterWishlistRow, error)
	GetUsersWantingCharacterFunc     func(ctx context.Context, arg storage.GetUsersWantingCharacterParams) ([]uint64, error)
	GetWantedCharactersFunc          func(ctx context.Context, arg storage.GetWantedCharactersParams) ([]storage.GetWantedCharactersRow, error)
	GetWishlistHoldersFunc           func(ctx context.Context, arg storage.GetWishlistHoldersParams) ([]storage.GetWishlistHoldersRow, error)
	RemoveAllFromWishlistFunc        func(ctx context.Context, userID uint64) error
	RemoveCharactersFromWishlistFunc func(ctx context.Context, arg storage.RemoveCharactersFromWishlistParams) error
}

func (m *MockQuerier) AddCharactersToWishlist(ctx context.Context, arg storage.AddCharactersToWishlistParams) (int64, error) {
	if m.AddCharactersToWishlistFunc != nil {
		return m.AddCharactersToWishlistFunc(ctx, arg)
	}
	return 0, nil
}

func (m *MockQuerier) CompareWithUser(ctx context.Context, arg storage.CompareWithUserParams) ([]storage.CompareWithUserRow, error) {
	if m.CompareWithUserFunc != nil {
		return m.CompareWithUserFunc(ctx, arg)
	}
	return nil, nil
}

func (m *MockQuerier) GetUserCharacterWishlist(ctx context.Context, userID uint64) ([]storage.GetUserCharacterWishlistRow, error) {
	if m.GetUserCharacterWishlistFunc != nil {
		return m.GetUserCharacterWishlistFunc(ctx, userID)
	}
	return nil, nil
}

func (m *MockQuerier) GetWantedCharacters(ctx context.Context, arg storage.GetWantedCharactersParams) ([]storage.GetWantedCharactersRow, error) {
	if m.GetWantedCharactersFunc != nil {
		return m.GetWantedCharactersFunc(ctx, arg)
	}
	return nil, nil
}

func (m *MockQuerier) GetWishlistHolders(ctx context.Context, arg storage.GetWishlistHoldersParams) ([]storage.GetWishlistHoldersRow, error) {
	if m.GetWishlistHoldersFunc != nil {
		return m.GetWishlistHoldersFunc(ctx, arg)
	}
	return nil, nil
}

func (m *MockQuerier) RemoveAllFromWishlist(ctx context.Context, userID uint64) error {
	if m.RemoveAllFromWishlistFunc != nil {
		return m.RemoveAllFromWishlistFunc(ctx, userID)
	}
	return nil
}

func (m *MockQuerier) RemoveCharactersFromWishlist(ctx context.Context, arg storage.RemoveCharactersFromWishlistParams) error {
	if m.RemoveCharactersFromWishlistFunc != nil {
		return m.RemoveCharactersFromWishlistFunc(ctx, arg)
	}
	return nil
}

func (m *MockQuerier) GetUsersWantingCharacter(ctx context.Context, arg storage.GetUsersWantingCharacterParams) ([]uint64, error) {
	if m.GetUsersWantingCharacterFunc != nil {
		return m.GetUsersWantingCharacterFunc(ctx, arg)
	}
	return nil, nil
}

var _ storage.Querier = (*MockQuerier)(nil)
