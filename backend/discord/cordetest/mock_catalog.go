package cordetest

import (
	"context"

	"github.com/karitham/waifubot/catalog"
)

// MockCatalogStore implements catalog.Store for testing.
type MockCatalogStore struct {
	UpsertCharacterFunc            func(ctx context.Context, char catalog.Character) error
	GetCharacterByIDFunc           func(ctx context.Context, charID int64) (catalog.Character, error)
	SearchCharactersFunc           func(ctx context.Context, userID uint64, term string) ([]catalog.Character, error)
	SearchGlobalCharactersFunc     func(ctx context.Context, term string) ([]catalog.Character, error)
	GetCharacterHoldersInGuildFunc func(ctx context.Context, guildID uint64, charID int64) ([]uint64, error)
	GetActiveIDsFunc               func(ctx context.Context) ([]int64, error)
	MarkCharactersInactiveFunc     func(ctx context.Context, ids []int64) error
}

var _ catalog.Store = (*MockCatalogStore)(nil)

func (m *MockCatalogStore) UpsertCharacter(ctx context.Context, char catalog.Character) error {
	if m.UpsertCharacterFunc != nil {
		return m.UpsertCharacterFunc(ctx, char)
	}
	return nil
}

func (m *MockCatalogStore) GetCharacterByID(ctx context.Context, charID int64) (catalog.Character, error) {
	if m.GetCharacterByIDFunc != nil {
		return m.GetCharacterByIDFunc(ctx, charID)
	}
	return catalog.Character{}, nil
}

func (m *MockCatalogStore) SearchCharacters(ctx context.Context, userID uint64, term string) ([]catalog.Character, error) {
	if m.SearchCharactersFunc != nil {
		return m.SearchCharactersFunc(ctx, userID, term)
	}
	return nil, nil
}

func (m *MockCatalogStore) SearchGlobalCharacters(ctx context.Context, term string) ([]catalog.Character, error) {
	if m.SearchGlobalCharactersFunc != nil {
		return m.SearchGlobalCharactersFunc(ctx, term)
	}
	return nil, nil
}

func (m *MockCatalogStore) GetCharacterHoldersInGuild(ctx context.Context, guildID uint64, charID int64) ([]uint64, error) {
	if m.GetCharacterHoldersInGuildFunc != nil {
		return m.GetCharacterHoldersInGuildFunc(ctx, guildID, charID)
	}
	return nil, nil
}

func (m *MockCatalogStore) MarkCharactersInactive(ctx context.Context, ids []int64) error {
	if m.MarkCharactersInactiveFunc != nil {
		return m.MarkCharactersInactiveFunc(ctx, ids)
	}
	return nil
}

func (m *MockCatalogStore) GetActiveIDs(ctx context.Context) ([]int64, error) {
	if m.GetActiveIDsFunc != nil {
		return m.GetActiveIDsFunc(ctx)
	}
	return nil, nil
}
