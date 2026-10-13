package collectiontest

import (
	"context"
	"time"

	"github.com/karitham/waifubot/catalog"
	"github.com/karitham/waifubot/collection"
	"github.com/karitham/waifubot/guild"
)

type MockStore struct {
	GetUserFunc                  func(ctx context.Context, userID collection.UserID) (collection.User, error)
	CreateUserFunc               func(ctx context.Context, userID collection.UserID) error
	LockUserFunc                 func(ctx context.Context, userID collection.UserID) error
	GetUserByAnilistFunc         func(ctx context.Context, anilistURL string) (collection.User, error)
	GetUserByDiscordUsernameFunc func(ctx context.Context, username string) (collection.User, error)
	RecordRollFunc               func(ctx context.Context, userID collection.UserID, now time.Time, cooldown time.Duration) error
	SpendTokensFunc              func(ctx context.Context, userID collection.UserID, amount collection.TokenAmount) (collection.User, error)
	AddTokensFunc                func(ctx context.Context, userID collection.UserID, amount collection.TokenAmount) (collection.User, error)
	UpdateQuoteFunc              func(ctx context.Context, userID collection.UserID, quote string) error
	UpdateAnilistURLFunc         func(ctx context.Context, userID collection.UserID, url string) error
	UpdateDiscordInfoFunc        func(ctx context.Context, userID collection.UserID, username, avatar string, lastUpdated time.Time) error

	GetCollectionFunc           func(ctx context.Context, userID collection.UserID) ([]collection.OwnedCharacter, error)
	GetCollectionIDsFunc        func(ctx context.Context, userID collection.UserID) ([]catalog.CharacterID, error)
	GetOwnedCharacterFunc       func(ctx context.Context, userID collection.UserID, charID catalog.CharacterID) (collection.OwnedCharacter, error)
	AddToCollectionFunc         func(ctx context.Context, userID collection.UserID, charID catalog.CharacterID, source collection.Source, acquiredAt time.Time) error
	RemoveFromCollectionFunc    func(ctx context.Context, userID collection.UserID, charID catalog.CharacterID) error
	GiveCharacterFunc           func(ctx context.Context, from, to collection.UserID, charID catalog.CharacterID) (collection.OwnedCharacter, error)
	CountCollectionFunc         func(ctx context.Context, userID collection.UserID) (int64, error)
	SetFavoriteFunc             func(ctx context.Context, userID collection.UserID, charID catalog.CharacterID) error
	SearchCharactersFunc        func(ctx context.Context, userID collection.UserID, term string) ([]catalog.Character, error)
	RemoveFromWishlistFunc      func(ctx context.Context, userID collection.UserID, charID catalog.CharacterID) error
	RemoveAllFromWishlistFunc   func(ctx context.Context, userID collection.UserID) error
	AddCharactersToWishlistFunc func(ctx context.Context, userID collection.UserID, characterIDs []catalog.CharacterID) (int64, error)

	GetUserCharacterWishlistFunc func(ctx context.Context, userID collection.UserID) ([]collection.WishlistEntry, error)
	GetWishlistHoldersFunc       func(ctx context.Context, characterIDs []catalog.CharacterID, userID collection.UserID, guildID guild.ID) ([]collection.WishlistUserSet, error)
	GetWantedCharactersFunc      func(ctx context.Context, userID collection.UserID, guildID guild.ID) ([]collection.WishlistUserSet, error)
	CompareWithUserFunc          func(ctx context.Context, userID1, userID2 collection.UserID) (collection.WishlistComparison, error)
	GetUsersWantingCharacterFunc func(ctx context.Context, charID catalog.CharacterID, guildID guild.ID, excludeUserID collection.UserID) ([]guild.MemberID, error)

	SetDropFunc          func(ctx context.Context, channelID collection.ChannelID, char catalog.Character) error
	GetDropForUpdateFunc func(ctx context.Context, channelID collection.ChannelID) (catalog.Character, error)
	DeleteDropFunc       func(ctx context.Context, channelID collection.ChannelID) error

	IsGuildIndexedFunc             func(ctx context.Context, guildID guild.ID) (guild.IndexStatus, error)
	StartIndexingJobFunc           func(ctx context.Context, guildID guild.ID) error
	CompleteIndexingJobFunc        func(ctx context.Context, guildID guild.ID) error
	UpsertGuildMembersFunc         func(ctx context.Context, guildID guild.ID, memberIDs []guild.MemberID, indexedAt time.Time) error
	DeleteGuildMembersNotInFunc    func(ctx context.Context, guildID guild.ID, memberIDs []guild.MemberID) error
	GetCharacterHoldersInGuildFunc func(ctx context.Context, guildID guild.ID, charID catalog.CharacterID) ([]guild.MemberID, error)

	UpsertCharacterFunc        func(ctx context.Context, char catalog.Character) error
	GetCharacterByIDFunc       func(ctx context.Context, charID catalog.CharacterID) (catalog.Character, error)
	SearchGlobalCharactersFunc func(ctx context.Context, term string) ([]catalog.Character, error)
	GetActiveIDsFunc           func(ctx context.Context) ([]catalog.CharacterID, error)
	KnownCharacterIDsFunc      func(ctx context.Context, ids []catalog.CharacterID) ([]catalog.CharacterID, error)
	MarkCharactersInactiveFunc func(ctx context.Context, charIDs []catalog.CharacterID) error

	RandomCharNotOwnedFunc func(ctx context.Context, userID collection.UserID, weightExponent float64) (catalog.Character, error)
	RandomActiveCharFunc   func(ctx context.Context, weightExponent float64) (catalog.Character, error)

	BeginFunc    func(ctx context.Context) (collection.Tx, error)
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

func (m *MockStore) LockUser(ctx context.Context, userID collection.UserID) error {
	if m.LockUserFunc != nil {
		return m.LockUserFunc(ctx, userID)
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

func (m *MockStore) RecordRoll(ctx context.Context, userID collection.UserID, now time.Time, cooldown time.Duration) error {
	if m.RecordRollFunc != nil {
		return m.RecordRollFunc(ctx, userID, now, cooldown)
	}
	return nil
}

func (m *MockStore) SpendTokens(ctx context.Context, userID collection.UserID, amount collection.TokenAmount) (collection.User, error) {
	if m.SpendTokensFunc != nil {
		return m.SpendTokensFunc(ctx, userID, amount)
	}
	return collection.User{UserID: userID}, nil
}

func (m *MockStore) AddTokens(ctx context.Context, userID collection.UserID, amount collection.TokenAmount) (collection.User, error) {
	if m.AddTokensFunc != nil {
		return m.AddTokensFunc(ctx, userID, amount)
	}
	return collection.User{UserID: userID}, nil
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

func (m *MockStore) GetCollectionIDs(ctx context.Context, userID collection.UserID) ([]catalog.CharacterID, error) {
	if m.GetCollectionIDsFunc != nil {
		return m.GetCollectionIDsFunc(ctx, userID)
	}
	return nil, nil
}

func (m *MockStore) GetOwnedCharacter(ctx context.Context, userID collection.UserID, charID catalog.CharacterID) (collection.OwnedCharacter, error) {
	if m.GetOwnedCharacterFunc != nil {
		return m.GetOwnedCharacterFunc(ctx, userID, charID)
	}
	return collection.OwnedCharacter{}, nil
}

func (m *MockStore) AddToCollection(ctx context.Context, userID collection.UserID, charID catalog.CharacterID, source collection.Source, acquiredAt time.Time) error {
	if m.AddToCollectionFunc != nil {
		return m.AddToCollectionFunc(ctx, userID, charID, source, acquiredAt)
	}
	return nil
}

func (m *MockStore) RemoveFromCollection(ctx context.Context, userID collection.UserID, charID catalog.CharacterID) error {
	if m.RemoveFromCollectionFunc != nil {
		return m.RemoveFromCollectionFunc(ctx, userID, charID)
	}
	return nil
}

func (m *MockStore) GiveCharacter(ctx context.Context, from, to collection.UserID, charID catalog.CharacterID) (collection.OwnedCharacter, error) {
	if m.GiveCharacterFunc != nil {
		return m.GiveCharacterFunc(ctx, from, to, charID)
	}
	return collection.OwnedCharacter{}, nil
}

func (m *MockStore) SetFavorite(ctx context.Context, userID collection.UserID, charID catalog.CharacterID) error {
	if m.SetFavoriteFunc != nil {
		return m.SetFavoriteFunc(ctx, userID, charID)
	}
	return nil
}

func (m *MockStore) CountCollection(ctx context.Context, userID collection.UserID) (int64, error) {
	if m.CountCollectionFunc != nil {
		return m.CountCollectionFunc(ctx, userID)
	}
	return 0, nil
}

func (m *MockStore) SearchCharacters(ctx context.Context, userID collection.UserID, term string) ([]catalog.Character, error) {
	if m.SearchCharactersFunc != nil {
		return m.SearchCharactersFunc(ctx, userID, term)
	}
	return nil, nil
}

func (m *MockStore) RemoveFromWishlist(ctx context.Context, userID collection.UserID, charID catalog.CharacterID) error {
	if m.RemoveFromWishlistFunc != nil {
		return m.RemoveFromWishlistFunc(ctx, userID, charID)
	}
	return nil
}

func (m *MockStore) RemoveAllFromWishlist(ctx context.Context, userID collection.UserID) error {
	if m.RemoveAllFromWishlistFunc != nil {
		return m.RemoveAllFromWishlistFunc(ctx, userID)
	}
	return nil
}

func (m *MockStore) AddCharactersToWishlist(ctx context.Context, userID collection.UserID, characterIDs []catalog.CharacterID) (int64, error) {
	if m.AddCharactersToWishlistFunc != nil {
		return m.AddCharactersToWishlistFunc(ctx, userID, characterIDs)
	}
	return 0, nil
}

func (m *MockStore) SetDrop(ctx context.Context, channelID collection.ChannelID, char catalog.Character) error {
	if m.SetDropFunc != nil {
		return m.SetDropFunc(ctx, channelID, char)
	}
	return nil
}

func (m *MockStore) GetDropForUpdate(ctx context.Context, channelID collection.ChannelID) (catalog.Character, error) {
	if m.GetDropForUpdateFunc != nil {
		return m.GetDropForUpdateFunc(ctx, channelID)
	}
	return catalog.Character{}, nil
}

func (m *MockStore) DeleteDrop(ctx context.Context, channelID collection.ChannelID) error {
	if m.DeleteDropFunc != nil {
		return m.DeleteDropFunc(ctx, channelID)
	}
	return nil
}

func (m *MockStore) IsGuildIndexed(ctx context.Context, guildID guild.ID) (guild.IndexStatus, error) {
	if m.IsGuildIndexedFunc != nil {
		return m.IsGuildIndexedFunc(ctx, guildID)
	}
	return guild.IndexStatus{}, nil
}

func (m *MockStore) StartIndexingJob(ctx context.Context, guildID guild.ID) error {
	if m.StartIndexingJobFunc != nil {
		return m.StartIndexingJobFunc(ctx, guildID)
	}
	return nil
}

func (m *MockStore) CompleteIndexingJob(ctx context.Context, guildID guild.ID) error {
	if m.CompleteIndexingJobFunc != nil {
		return m.CompleteIndexingJobFunc(ctx, guildID)
	}
	return nil
}

func (m *MockStore) UpsertGuildMembers(ctx context.Context, guildID guild.ID, memberIDs []guild.MemberID, indexedAt time.Time) error {
	if m.UpsertGuildMembersFunc != nil {
		return m.UpsertGuildMembersFunc(ctx, guildID, memberIDs, indexedAt)
	}
	return nil
}

func (m *MockStore) DeleteGuildMembersNotIn(ctx context.Context, guildID guild.ID, memberIDs []guild.MemberID) error {
	if m.DeleteGuildMembersNotInFunc != nil {
		return m.DeleteGuildMembersNotInFunc(ctx, guildID, memberIDs)
	}
	return nil
}

func (m *MockStore) GetCharacterHoldersInGuild(ctx context.Context, guildID guild.ID, charID catalog.CharacterID) ([]guild.MemberID, error) {
	if m.GetCharacterHoldersInGuildFunc != nil {
		return m.GetCharacterHoldersInGuildFunc(ctx, guildID, charID)
	}
	return nil, nil
}

func (m *MockStore) UpsertCharacter(ctx context.Context, char catalog.Character) error {
	if m.UpsertCharacterFunc != nil {
		return m.UpsertCharacterFunc(ctx, char)
	}
	return nil
}

func (m *MockStore) GetCharacterByID(ctx context.Context, charID catalog.CharacterID) (catalog.Character, error) {
	if m.GetCharacterByIDFunc != nil {
		return m.GetCharacterByIDFunc(ctx, charID)
	}
	return catalog.Character{}, nil
}

func (m *MockStore) SearchGlobalCharacters(ctx context.Context, term string) ([]catalog.Character, error) {
	if m.SearchGlobalCharactersFunc != nil {
		return m.SearchGlobalCharactersFunc(ctx, term)
	}
	return nil, nil
}

func (m *MockStore) GetUserCharacterWishlist(ctx context.Context, userID collection.UserID) ([]collection.WishlistEntry, error) {
	if m.GetUserCharacterWishlistFunc != nil {
		return m.GetUserCharacterWishlistFunc(ctx, userID)
	}
	return nil, nil
}

func (m *MockStore) GetWishlistHolders(ctx context.Context, characterIDs []catalog.CharacterID, userID collection.UserID, guildID guild.ID) ([]collection.WishlistUserSet, error) {
	if m.GetWishlistHoldersFunc != nil {
		return m.GetWishlistHoldersFunc(ctx, characterIDs, userID, guildID)
	}
	return nil, nil
}

func (m *MockStore) GetWantedCharacters(ctx context.Context, userID collection.UserID, guildID guild.ID) ([]collection.WishlistUserSet, error) {
	if m.GetWantedCharactersFunc != nil {
		return m.GetWantedCharactersFunc(ctx, userID, guildID)
	}
	return nil, nil
}

func (m *MockStore) CompareWithUser(ctx context.Context, userID1, userID2 collection.UserID) (collection.WishlistComparison, error) {
	if m.CompareWithUserFunc != nil {
		return m.CompareWithUserFunc(ctx, userID1, userID2)
	}
	return collection.WishlistComparison{}, nil
}

func (m *MockStore) GetUsersWantingCharacter(ctx context.Context, charID catalog.CharacterID, guildID guild.ID, excludeUserID collection.UserID) ([]guild.MemberID, error) {
	if m.GetUsersWantingCharacterFunc != nil {
		return m.GetUsersWantingCharacterFunc(ctx, charID, guildID, excludeUserID)
	}
	return nil, nil
}

func (m *MockStore) KnownCharacterIDs(ctx context.Context, ids []catalog.CharacterID) ([]catalog.CharacterID, error) {
	if m.KnownCharacterIDsFunc != nil {
		return m.KnownCharacterIDsFunc(ctx, ids)
	}
	return ids, nil
}

func (m *MockStore) GetActiveIDs(ctx context.Context) ([]catalog.CharacterID, error) {
	if m.GetActiveIDsFunc != nil {
		return m.GetActiveIDsFunc(ctx)
	}
	return nil, nil
}

func (m *MockStore) MarkCharactersInactive(ctx context.Context, charIDs []catalog.CharacterID) error {
	if m.MarkCharactersInactiveFunc != nil {
		return m.MarkCharactersInactiveFunc(ctx, charIDs)
	}
	return nil
}

func (m *MockStore) RandomCharNotOwned(ctx context.Context, userID collection.UserID, weightExponent float64) (catalog.Character, error) {
	if m.RandomCharNotOwnedFunc != nil {
		return m.RandomCharNotOwnedFunc(ctx, userID, weightExponent)
	}
	return catalog.Character{}, nil
}

func (m *MockStore) RandomActiveChar(ctx context.Context, weightExponent float64) (catalog.Character, error) {
	if m.RandomActiveCharFunc != nil {
		return m.RandomActiveCharFunc(ctx, weightExponent)
	}
	return catalog.Character{}, nil
}

func (m *MockStore) Begin(ctx context.Context) (collection.Tx, error) {
	if m.BeginFunc != nil {
		return m.BeginFunc(ctx)
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

var (
	_ collection.Store = (*MockStore)(nil)
	_ collection.Tx    = (*MockStore)(nil)
	_ collection.DB    = (*MockStore)(nil)
)
