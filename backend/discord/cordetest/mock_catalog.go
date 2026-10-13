package cordetest

import (
	"context"

	"github.com/karitham/waifubot/catalog"
)

// MockCatalogStore implements catalog.Store for testing.
type MockCatalogStore struct {
	UpsertCharacterFunc        func(ctx context.Context, char catalog.Character) error
	GetCharacterByIDFunc       func(ctx context.Context, charID catalog.CharacterID) (catalog.Character, error)
	SearchGlobalCharactersFunc func(ctx context.Context, term string) ([]catalog.Character, error)
	GetActiveIDsFunc           func(ctx context.Context) ([]catalog.CharacterID, error)
	KnownCharacterIDsFunc      func(ctx context.Context, ids []catalog.CharacterID) ([]catalog.CharacterID, error)
	MarkCharactersInactiveFunc func(ctx context.Context, charIDs []catalog.CharacterID) error
}

var _ catalog.Store = (*MockCatalogStore)(nil)

func (m *MockCatalogStore) UpsertCharacter(ctx context.Context, char catalog.Character) error {
	if m.UpsertCharacterFunc != nil {
		return m.UpsertCharacterFunc(ctx, char)
	}
	return nil
}

func (m *MockCatalogStore) GetCharacterByID(ctx context.Context, charID catalog.CharacterID) (catalog.Character, error) {
	if m.GetCharacterByIDFunc != nil {
		return m.GetCharacterByIDFunc(ctx, charID)
	}
	return catalog.Character{}, nil
}

func (m *MockCatalogStore) SearchGlobalCharacters(ctx context.Context, term string) ([]catalog.Character, error) {
	if m.SearchGlobalCharactersFunc != nil {
		return m.SearchGlobalCharactersFunc(ctx, term)
	}
	return nil, nil
}

func (m *MockCatalogStore) KnownCharacterIDs(ctx context.Context, ids []catalog.CharacterID) ([]catalog.CharacterID, error) {
	if m.KnownCharacterIDsFunc != nil {
		return m.KnownCharacterIDsFunc(ctx, ids)
	}
	return ids, nil
}

func (m *MockCatalogStore) MarkCharactersInactive(ctx context.Context, charIDs []catalog.CharacterID) error {
	if m.MarkCharactersInactiveFunc != nil {
		return m.MarkCharactersInactiveFunc(ctx, charIDs)
	}
	return nil
}

func (m *MockCatalogStore) GetActiveIDs(ctx context.Context) ([]catalog.CharacterID, error) {
	if m.GetActiveIDsFunc != nil {
		return m.GetActiveIDsFunc(ctx)
	}
	return nil, nil
}
