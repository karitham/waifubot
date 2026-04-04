package discord

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/karitham/waifubot/collection"
	"github.com/karitham/waifubot/collection/collectiontest"
	"github.com/karitham/waifubot/discord/cordetest"
	"github.com/karitham/waifubot/wishlist"
)

type mockWishlistStore struct {
	AddCharactersToWishlistFunc      func(ctx context.Context, userID uint64, characterIDs []int64) error
	RemoveCharactersFromWishlistFunc func(ctx context.Context, userID uint64, characterIDs []int64) error
	RemoveAllFromWishlistFunc        func(ctx context.Context, userID uint64) error
	GetUserCharacterWishlistFunc     func(ctx context.Context, userID uint64) ([]wishlist.Character, error)
	GetWishlistHoldersFunc           func(ctx context.Context, characterIDs []int64, userID, guildID uint64) ([]wishlist.UserCharacterSet, error)
	GetWantedCharactersFunc          func(ctx context.Context, userID, guildID uint64) ([]wishlist.UserCharacterSet, error)
	CompareWithUserFunc              func(ctx context.Context, userID1, userID2 uint64) (wishlist.WishlistComparison, error)
	GetUsersWantingCharacterFunc     func(ctx context.Context, charID int64, guildID, excludeUserID uint64) ([]uint64, error)
}

func (m *mockWishlistStore) AddCharactersToWishlist(ctx context.Context, userID uint64, characterIDs []int64) error {
	if m.AddCharactersToWishlistFunc != nil {
		return m.AddCharactersToWishlistFunc(ctx, userID, characterIDs)
	}
	return nil
}
func (m *mockWishlistStore) RemoveCharactersFromWishlist(ctx context.Context, userID uint64, characterIDs []int64) error {
	if m.RemoveCharactersFromWishlistFunc != nil {
		return m.RemoveCharactersFromWishlistFunc(ctx, userID, characterIDs)
	}
	return nil
}
func (m *mockWishlistStore) RemoveAllFromWishlist(ctx context.Context, userID uint64) error {
	if m.RemoveAllFromWishlistFunc != nil {
		return m.RemoveAllFromWishlistFunc(ctx, userID)
	}
	return nil
}
func (m *mockWishlistStore) GetUserCharacterWishlist(ctx context.Context, userID uint64) ([]wishlist.Character, error) {
	if m.GetUserCharacterWishlistFunc != nil {
		return m.GetUserCharacterWishlistFunc(ctx, userID)
	}
	return nil, nil
}
func (m *mockWishlistStore) GetWishlistHolders(ctx context.Context, characterIDs []int64, userID, guildID uint64) ([]wishlist.UserCharacterSet, error) {
	if m.GetWishlistHoldersFunc != nil {
		return m.GetWishlistHoldersFunc(ctx, characterIDs, userID, guildID)
	}
	return nil, nil
}
func (m *mockWishlistStore) GetWantedCharacters(ctx context.Context, userID, guildID uint64) ([]wishlist.UserCharacterSet, error) {
	if m.GetWantedCharactersFunc != nil {
		return m.GetWantedCharactersFunc(ctx, userID, guildID)
	}
	return nil, nil
}
func (m *mockWishlistStore) CompareWithUser(ctx context.Context, userID1, userID2 uint64) (wishlist.WishlistComparison, error) {
	if m.CompareWithUserFunc != nil {
		return m.CompareWithUserFunc(ctx, userID1, userID2)
	}
	return wishlist.WishlistComparison{}, nil
}
func (m *mockWishlistStore) GetUsersWantingCharacter(ctx context.Context, charID int64, guildID, excludeUserID uint64) ([]uint64, error) {
	if m.GetUsersWantingCharacterFunc != nil {
		return m.GetUsersWantingCharacterFunc(ctx, charID, guildID, excludeUserID)
	}
	return nil, nil
}

func TestRollHandler_Roll(t *testing.T) {
	tests := []struct {
		name           string
		store          *collectiontest.MockStore
		animeService   *collectiontest.MockAnimeService
		wishlist       *mockWishlistStore
		cmd            CommandContext
		config         collection.Config
		wantContent    string
		wantEmbedTitle string
		wantEmbedDesc  string
		wantResponded  bool
	}{
		{
			name: "roll cooldown",
			store: &collectiontest.MockStore{
				GetUserFunc: func(ctx context.Context, userID collection.UserID) (collection.User, error) {
					return collection.User{
						UserID: userID,
						Date:   time.Now().Add(30 * time.Minute),
					}, nil
				},
			},
			animeService: &collectiontest.MockAnimeService{},
			wishlist:     &mockWishlistStore{},
			cmd: &MockCommandContext{
				UserIDVal:  1,
				GuildIDVal: 2,
			},
			config:        collection.Config{RollCooldown: time.Hour, TokensNeeded: 10},
			wantContent:   "roll again",
			wantResponded: true,
		},
		{
			name: "store error",
			store: &collectiontest.MockStore{
				GetUserFunc: func(ctx context.Context, userID collection.UserID) (collection.User, error) {
					return collection.User{}, errors.New("database on fire")
				},
			},
			animeService: &collectiontest.MockAnimeService{},
			wishlist:     &mockWishlistStore{},
			cmd: &MockCommandContext{
				UserIDVal:  1,
				GuildIDVal: 2,
			},
			config:        collection.Config{RollCooldown: time.Hour, TokensNeeded: 10},
			wantContent:   "error occurred",
			wantResponded: true,
		},
		{
			name: "success no wishlist",
			store: &collectiontest.MockStore{
				GetUserFunc: func(ctx context.Context, userID collection.UserID) (collection.User, error) {
					return collection.User{UserID: userID}, nil
				},
				GetCollectionIDsFunc: func(ctx context.Context, userID collection.UserID) ([]int64, error) {
					return nil, nil
				},
				UpsertCharacterFunc: func(ctx context.Context, char collection.Character) error { return nil },
				AddToCollectionFunc: func(ctx context.Context, userID collection.UserID, char collection.Character, source string, acquiredAt time.Time) error {
					return nil
				},
				RemoveFromWishlistFunc: func(ctx context.Context, userID collection.UserID, charID int64) error { return nil },
				UpdateLastRollFunc:     func(ctx context.Context, userID collection.UserID, when time.Time) error { return nil },
				CommitFunc:             func(ctx context.Context) error { return nil },
			},
			animeService: &collectiontest.MockAnimeService{
				RandomCharFunc: func(ctx context.Context, notIn ...int64) (collection.MediaCharacter, error) {
					return collection.MediaCharacter{
						ID:         42,
						Name:       "Rem",
						ImageURL:   "https://example.com/rem.png",
						MediaTitle: "Re:Zero",
						Favorites:  5000,
					}, nil
				},
			},
			wishlist: &mockWishlistStore{
				GetUsersWantingCharacterFunc: func(ctx context.Context, charID int64, guildID, excludeUserID uint64) ([]uint64, error) {
					return nil, nil
				},
			},
			cmd: &MockCommandContext{
				UserIDVal:  1,
				GuildIDVal: 2,
			},
			config:         collection.Config{RollCooldown: time.Hour, TokensNeeded: 10},
			wantEmbedTitle: "Rem",
			wantResponded:  true,
		},
		{
			name: "success with wishlist",
			store: &collectiontest.MockStore{
				GetUserFunc: func(ctx context.Context, userID collection.UserID) (collection.User, error) {
					return collection.User{UserID: userID}, nil
				},
				GetCollectionIDsFunc: func(ctx context.Context, userID collection.UserID) ([]int64, error) {
					return nil, nil
				},
				UpsertCharacterFunc: func(ctx context.Context, char collection.Character) error { return nil },
				AddToCollectionFunc: func(ctx context.Context, userID collection.UserID, char collection.Character, source string, acquiredAt time.Time) error {
					return nil
				},
				RemoveFromWishlistFunc: func(ctx context.Context, userID collection.UserID, charID int64) error { return nil },
				UpdateLastRollFunc:     func(ctx context.Context, userID collection.UserID, when time.Time) error { return nil },
				CommitFunc:             func(ctx context.Context) error { return nil },
			},
			animeService: &collectiontest.MockAnimeService{
				RandomCharFunc: func(ctx context.Context, notIn ...int64) (collection.MediaCharacter, error) {
					return collection.MediaCharacter{
						ID:         42,
						Name:       "Rem",
						ImageURL:   "https://example.com/rem.png",
						MediaTitle: "Re:Zero",
						Favorites:  5000,
					}, nil
				},
			},
			wishlist: &mockWishlistStore{
				GetUsersWantingCharacterFunc: func(ctx context.Context, charID int64, guildID, excludeUserID uint64) ([]uint64, error) {
					return []uint64{10, 20, 30}, nil
				},
			},
			cmd: &MockCommandContext{
				UserIDVal:  1,
				GuildIDVal: 2,
			},
			config:         collection.Config{RollCooldown: time.Hour, TokensNeeded: 10},
			wantEmbedTitle: "Rem",
			wantEmbedDesc:  "<@10> <@20> <@30>",
			wantResponded:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &cordetest.MockResponseWriter{}
			h := &RollHandler{
				store:        tt.store,
				animeService: tt.animeService,
				wishlist:     tt.wishlist,
				config:       tt.config,
			}

			h.Roll(t.Context(), w, tt.cmd)

			assert.Equal(t, tt.wantResponded, w.RespondCalled)
			if tt.wantContent != "" {
				data := w.LastRespond.InteractionRespData()
				assert.Contains(t, data.Content, tt.wantContent)
			}
			if tt.wantEmbedTitle != "" {
				data := w.LastRespond.InteractionRespData()
				if assert.Len(t, data.Embeds, 1) {
					assert.Equal(t, tt.wantEmbedTitle, data.Embeds[0].Title)
				}
			}
			if tt.wantEmbedDesc != "" {
				data := w.LastRespond.InteractionRespData()
				if assert.Len(t, data.Embeds, 1) {
					assert.Contains(t, data.Embeds[0].Description, tt.wantEmbedDesc)
				}
			}
		})
	}
}
