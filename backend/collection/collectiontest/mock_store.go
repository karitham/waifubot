package collectiontest

import (
	"context"
	"time"

	"github.com/karitham/waifubot/catalog"
	"github.com/karitham/waifubot/collection"
)

type MockStore struct {
	GetUserFunc                  func(ctx context.Context, userID collection.UserID) (collection.User, error)
	CreateUserFunc               func(ctx context.Context, userID collection.UserID) error
	GetUserByAnilistFunc         func(ctx context.Context, anilistURL string) (collection.User, error)
	GetUserByDiscordUsernameFunc func(ctx context.Context, username string) (collection.User, error)
	UpdateLastRollFunc           func(ctx context.Context, userID collection.UserID, when time.Time) error
	SpendTokensFunc              func(ctx context.Context, userID collection.UserID, amount int32) (collection.User, error)
	AddTokensFunc                func(ctx context.Context, userID collection.UserID, amount int32) (collection.User, error)
	UpdateFavoriteFunc           func(ctx context.Context, userID collection.UserID, charID int64) error
	UpdateQuoteFunc              func(ctx context.Context, userID collection.UserID, quote string) error
	UpdateAnilistURLFunc         func(ctx context.Context, userID collection.UserID, url string) error
	UpdateDiscordInfoFunc        func(ctx context.Context, userID collection.UserID, username, avatar string, lastUpdated time.Time) error

	GetCollectionFunc        func(ctx context.Context, userID collection.UserID) ([]collection.OwnedCharacter, error)
	GetCollectionIDsFunc     func(ctx context.Context, userID collection.UserID) ([]int64, error)
	GetOwnedCharacterFunc    func(ctx context.Context, userID collection.UserID, charID int64) (collection.OwnedCharacter, error)
	AddToCollectionFunc      func(ctx context.Context, userID collection.UserID, char collection.Character, source string, acquiredAt time.Time) error
	RemoveFromCollectionFunc func(ctx context.Context, userID collection.UserID, charID int64) error
	GiveCharacterFunc        func(ctx context.Context, from, to collection.UserID, charID int64) (collection.OwnedCharacter, error)
	CountCollectionFunc      func(ctx context.Context, userID collection.UserID) (int64, error)
	RemoveFromWishlistFunc   func(ctx context.Context, userID collection.UserID, charID int64) error

	GetDropForUpdateFunc func(ctx context.Context, channelID uint64) (collection.Drop, error)
	DeleteDropFunc       func(ctx context.Context, channelID uint64) error

	IsGuildIndexedFunc          func(ctx context.Context, guildID uint64) (collection.GuildIndexStatus, error)
	StartIndexingJobFunc        func(ctx context.Context, guildID uint64) error
	CompleteIndexingJobFunc     func(ctx context.Context, guildID uint64) error
	UpsertGuildMembersFunc      func(ctx context.Context, guildID uint64, memberIDs []uint64, indexedAt time.Time) error
	DeleteGuildMembersNotInFunc func(ctx context.Context, guildID uint64, memberIDs []uint64) error

	UpsertCharacterFunc            func(ctx context.Context, char catalog.Character) error
	GetCharacterByIDFunc           func(ctx context.Context, charID int64) (catalog.Character, error)
	SearchCharactersFunc           func(ctx context.Context, userID uint64, term string) ([]catalog.Character, error)
	SearchGlobalCharactersFunc     func(ctx context.Context, term string) ([]catalog.Character, error)
	GetCharacterHoldersInGuildFunc func(ctx context.Context, guildID uint64, charID int64) ([]uint64, error)
	GetStaleCharactersFunc         func(ctx context.Context, cursorUpdatedAt time.Time, cursorID int64, limit int) ([]catalog.Character, error)
	UpdateCharacterSyncFunc        func(ctx context.Context, char catalog.Character) (catalog.Character, error)
	MarkCharacterInactiveFunc      func(ctx context.Context, charID int64) error
	GetActiveIDsFunc               func(ctx context.Context) ([]int64, error)

	RandomCharNotOwnedFunc func(ctx context.Context, userID collection.UserID) (catalog.Character, error)
	RandomActiveCharFunc   func(ctx context.Context) (catalog.Character, error)

	WithTxFunc   func(ctx context.Context) (collection.Store, error)
	CommitFunc   func(ctx context.Context) error
	RollbackFunc func(ctx context.Context) error

	CommitCalls   int
	RollbackCalls int
}

func (m *MockStore) GetUser(ctx context.Context, userID collection.UserID) (collection.User, error) {
	if m.GetUserFunc != nil {
		return m.GetUserFunc(ctx, userID)
	}
	return collection.User{UserID: userID}, nil
}

func (m *MockStore) CreateUser(ctx context.Context, userID collection.UserID) error {
	if m.CreateUserFunc != nil {
		return m.CreateUserFunc(ctx, userID)
	}
	return nil
}

func (m *MockStore) GetUserByAnilist(ctx context.Context, anilistURL string) (collection.User, error) {
	if m.GetUserByAnilistFunc != nil {
		return m.GetUserByAnilistFunc(ctx, anilistURL)
	}
	return collection.User{}, nil
}

func (m *MockStore) GetUserByDiscordUsername(ctx context.Context, username string) (collection.User, error) {
	if m.GetUserByDiscordUsernameFunc != nil {
		return m.GetUserByDiscordUsernameFunc(ctx, username)
	}
	return collection.User{}, nil
}

func (m *MockStore) UpdateLastRoll(ctx context.Context, userID collection.UserID, when time.Time) error {
	if m.UpdateLastRollFunc != nil {
		return m.UpdateLastRollFunc(ctx, userID, when)
	}
	return nil
}

func (m *MockStore) SpendTokens(ctx context.Context, userID collection.UserID, amount int32) (collection.User, error) {
	if m.SpendTokensFunc != nil {
		return m.SpendTokensFunc(ctx, userID, amount)
	}
	return collection.User{UserID: userID}, nil
}

func (m *MockStore) AddTokens(ctx context.Context, userID collection.UserID, amount int32) (collection.User, error) {
	if m.AddTokensFunc != nil {
		return m.AddTokensFunc(ctx, userID, amount)
	}
	return collection.User{UserID: userID}, nil
}

func (m *MockStore) UpdateFavorite(ctx context.Context, userID collection.UserID, charID int64) error {
	if m.UpdateFavoriteFunc != nil {
		return m.UpdateFavoriteFunc(ctx, userID, charID)
	}
	return nil
}

func (m *MockStore) UpdateQuote(ctx context.Context, userID collection.UserID, quote string) error {
	if m.UpdateQuoteFunc != nil {
		return m.UpdateQuoteFunc(ctx, userID, quote)
	}
	return nil
}

func (m *MockStore) UpdateAnilistURL(ctx context.Context, userID collection.UserID, url string) error {
	if m.UpdateAnilistURLFunc != nil {
		return m.UpdateAnilistURLFunc(ctx, userID, url)
	}
	return nil
}

func (m *MockStore) UpdateDiscordInfo(ctx context.Context, userID collection.UserID, username, avatar string, lastUpdated time.Time) error {
	if m.UpdateDiscordInfoFunc != nil {
		return m.UpdateDiscordInfoFunc(ctx, userID, username, avatar, lastUpdated)
	}
	return nil
}

func (m *MockStore) GetCollection(ctx context.Context, userID collection.UserID) ([]collection.OwnedCharacter, error) {
	if m.GetCollectionFunc != nil {
		return m.GetCollectionFunc(ctx, userID)
	}
	return nil, nil
}

func (m *MockStore) GetCollectionIDs(ctx context.Context, userID collection.UserID) ([]int64, error) {
	if m.GetCollectionIDsFunc != nil {
		return m.GetCollectionIDsFunc(ctx, userID)
	}
	return nil, nil
}

func (m *MockStore) GetOwnedCharacter(ctx context.Context, userID collection.UserID, charID int64) (collection.OwnedCharacter, error) {
	if m.GetOwnedCharacterFunc != nil {
		return m.GetOwnedCharacterFunc(ctx, userID, charID)
	}
	return collection.OwnedCharacter{}, nil
}

func (m *MockStore) AddToCollection(ctx context.Context, userID collection.UserID, char collection.Character, source string, acquiredAt time.Time) error {
	if m.AddToCollectionFunc != nil {
		return m.AddToCollectionFunc(ctx, userID, char, source, acquiredAt)
	}
	return nil
}

func (m *MockStore) RemoveFromCollection(ctx context.Context, userID collection.UserID, charID int64) error {
	if m.RemoveFromCollectionFunc != nil {
		return m.RemoveFromCollectionFunc(ctx, userID, charID)
	}
	return nil
}

func (m *MockStore) GiveCharacter(ctx context.Context, from, to collection.UserID, charID int64) (collection.OwnedCharacter, error) {
	if m.GiveCharacterFunc != nil {
		return m.GiveCharacterFunc(ctx, from, to, charID)
	}
	return collection.OwnedCharacter{}, nil
}

func (m *MockStore) CountCollection(ctx context.Context, userID collection.UserID) (int64, error) {
	if m.CountCollectionFunc != nil {
		return m.CountCollectionFunc(ctx, userID)
	}
	return 0, nil
}

func (m *MockStore) RemoveFromWishlist(ctx context.Context, userID collection.UserID, charID int64) error {
	if m.RemoveFromWishlistFunc != nil {
		return m.RemoveFromWishlistFunc(ctx, userID, charID)
	}
	return nil
}

func (m *MockStore) GetDropForUpdate(ctx context.Context, channelID uint64) (collection.Drop, error) {
	if m.GetDropForUpdateFunc != nil {
		return m.GetDropForUpdateFunc(ctx, channelID)
	}
	return collection.Drop{}, nil
}

func (m *MockStore) DeleteDrop(ctx context.Context, channelID uint64) error {
	if m.DeleteDropFunc != nil {
		return m.DeleteDropFunc(ctx, channelID)
	}
	return nil
}

func (m *MockStore) IsGuildIndexed(ctx context.Context, guildID uint64) (collection.GuildIndexStatus, error) {
	if m.IsGuildIndexedFunc != nil {
		return m.IsGuildIndexedFunc(ctx, guildID)
	}
	return collection.GuildIndexStatus{}, nil
}

func (m *MockStore) StartIndexingJob(ctx context.Context, guildID uint64) error {
	if m.StartIndexingJobFunc != nil {
		return m.StartIndexingJobFunc(ctx, guildID)
	}
	return nil
}

func (m *MockStore) CompleteIndexingJob(ctx context.Context, guildID uint64) error {
	if m.CompleteIndexingJobFunc != nil {
		return m.CompleteIndexingJobFunc(ctx, guildID)
	}
	return nil
}

func (m *MockStore) UpsertGuildMembers(ctx context.Context, guildID uint64, memberIDs []uint64, indexedAt time.Time) error {
	if m.UpsertGuildMembersFunc != nil {
		return m.UpsertGuildMembersFunc(ctx, guildID, memberIDs, indexedAt)
	}
	return nil
}

func (m *MockStore) DeleteGuildMembersNotIn(ctx context.Context, guildID uint64, memberIDs []uint64) error {
	if m.DeleteGuildMembersNotInFunc != nil {
		return m.DeleteGuildMembersNotInFunc(ctx, guildID, memberIDs)
	}
	return nil
}

func (m *MockStore) UpsertCharacter(ctx context.Context, char catalog.Character) error {
	if m.UpsertCharacterFunc != nil {
		return m.UpsertCharacterFunc(ctx, char)
	}
	return nil
}

func (m *MockStore) GetCharacterByID(ctx context.Context, charID int64) (catalog.Character, error) {
	if m.GetCharacterByIDFunc != nil {
		return m.GetCharacterByIDFunc(ctx, charID)
	}
	return catalog.Character{}, nil
}

func (m *MockStore) SearchCharacters(ctx context.Context, userID uint64, term string) ([]catalog.Character, error) {
	if m.SearchCharactersFunc != nil {
		return m.SearchCharactersFunc(ctx, userID, term)
	}
	return nil, nil
}

func (m *MockStore) SearchGlobalCharacters(ctx context.Context, term string) ([]catalog.Character, error) {
	if m.SearchGlobalCharactersFunc != nil {
		return m.SearchGlobalCharactersFunc(ctx, term)
	}
	return nil, nil
}

func (m *MockStore) GetCharacterHoldersInGuild(ctx context.Context, guildID uint64, charID int64) ([]uint64, error) {
	if m.GetCharacterHoldersInGuildFunc != nil {
		return m.GetCharacterHoldersInGuildFunc(ctx, guildID, charID)
	}
	return nil, nil
}

func (m *MockStore) GetStaleCharacters(ctx context.Context, cursorUpdatedAt time.Time, cursorID int64, limit int) ([]catalog.Character, error) {
	if m.GetStaleCharactersFunc != nil {
		return m.GetStaleCharactersFunc(ctx, cursorUpdatedAt, cursorID, limit)
	}
	return nil, nil
}

func (m *MockStore) UpdateCharacterSync(ctx context.Context, char catalog.Character) (catalog.Character, error) {
	if m.UpdateCharacterSyncFunc != nil {
		return m.UpdateCharacterSyncFunc(ctx, char)
	}
	return char, nil
}

func (m *MockStore) MarkCharacterInactive(ctx context.Context, charID int64) error {
	if m.MarkCharacterInactiveFunc != nil {
		return m.MarkCharacterInactiveFunc(ctx, charID)
	}
	return nil
}

func (m *MockStore) GetActiveIDs(ctx context.Context) ([]int64, error) {
	if m.GetActiveIDsFunc != nil {
		return m.GetActiveIDsFunc(ctx)
	}
	return nil, nil
}

func (m *MockStore) RandomCharNotOwned(ctx context.Context, userID collection.UserID) (catalog.Character, error) {
	if m.RandomCharNotOwnedFunc != nil {
		return m.RandomCharNotOwnedFunc(ctx, userID)
	}
	return catalog.Character{}, nil
}

func (m *MockStore) RandomActiveChar(ctx context.Context) (catalog.Character, error) {
	if m.RandomActiveCharFunc != nil {
		return m.RandomActiveCharFunc(ctx)
	}
	return catalog.Character{}, nil
}

func (m *MockStore) WithTx(ctx context.Context) (collection.Store, error) {
	if m.WithTxFunc != nil {
		return m.WithTxFunc(ctx)
	}
	return m, nil
}

func (m *MockStore) Commit(ctx context.Context) error {
	m.CommitCalls++
	if m.CommitFunc != nil {
		return m.CommitFunc(ctx)
	}
	return nil
}

func (m *MockStore) Rollback(ctx context.Context) error {
	m.RollbackCalls++
	if m.RollbackFunc != nil {
		return m.RollbackFunc(ctx)
	}
	return nil
}

var _ collection.Store = (*MockStore)(nil)
