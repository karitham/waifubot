package wishlisttest

import (
	"context"

	"github.com/karitham/waifubot/storage/wishliststore"
)

type MockQuerier struct {
	AddCharactersToWishlistFunc      func(ctx context.Context, arg wishliststore.AddCharactersToWishlistParams) error
	CompareWithUserFunc              func(ctx context.Context, arg wishliststore.CompareWithUserParams) ([]wishliststore.CompareWithUserRow, error)
	GetUserCharacterWishlistFunc     func(ctx context.Context, userID uint64) ([]wishliststore.GetUserCharacterWishlistRow, error)
	GetWantedCharactersFunc          func(ctx context.Context, arg wishliststore.GetWantedCharactersParams) ([]wishliststore.GetWantedCharactersRow, error)
	GetWishlistHoldersFunc           func(ctx context.Context, arg wishliststore.GetWishlistHoldersParams) ([]wishliststore.GetWishlistHoldersRow, error)
	RemoveAllFromWishlistFunc        func(ctx context.Context, userID uint64) error
	RemoveCharactersFromWishlistFunc func(ctx context.Context, arg wishliststore.RemoveCharactersFromWishlistParams) error
}

func (m *MockQuerier) AddCharactersToWishlist(ctx context.Context, arg wishliststore.AddCharactersToWishlistParams) error {
	if m.AddCharactersToWishlistFunc != nil {
		return m.AddCharactersToWishlistFunc(ctx, arg)
	}
	return nil
}

func (m *MockQuerier) CompareWithUser(ctx context.Context, arg wishliststore.CompareWithUserParams) ([]wishliststore.CompareWithUserRow, error) {
	if m.CompareWithUserFunc != nil {
		return m.CompareWithUserFunc(ctx, arg)
	}
	return nil, nil
}

func (m *MockQuerier) GetUserCharacterWishlist(ctx context.Context, userID uint64) ([]wishliststore.GetUserCharacterWishlistRow, error) {
	if m.GetUserCharacterWishlistFunc != nil {
		return m.GetUserCharacterWishlistFunc(ctx, userID)
	}
	return nil, nil
}

func (m *MockQuerier) GetWantedCharacters(ctx context.Context, arg wishliststore.GetWantedCharactersParams) ([]wishliststore.GetWantedCharactersRow, error) {
	if m.GetWantedCharactersFunc != nil {
		return m.GetWantedCharactersFunc(ctx, arg)
	}
	return nil, nil
}

func (m *MockQuerier) GetWishlistHolders(ctx context.Context, arg wishliststore.GetWishlistHoldersParams) ([]wishliststore.GetWishlistHoldersRow, error) {
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

func (m *MockQuerier) RemoveCharactersFromWishlist(ctx context.Context, arg wishliststore.RemoveCharactersFromWishlistParams) error {
	if m.RemoveCharactersFromWishlistFunc != nil {
		return m.RemoveCharactersFromWishlistFunc(ctx, arg)
	}
	return nil
}

var _ wishliststore.Querier = (*MockQuerier)(nil)
