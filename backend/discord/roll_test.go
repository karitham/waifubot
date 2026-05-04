package discord

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/karitham/waifubot/catalog"
	"github.com/karitham/waifubot/collection"
	"github.com/karitham/waifubot/collection/collectiontest"
	"github.com/karitham/waifubot/discord/cordetest"
	"github.com/karitham/waifubot/wishlist/wishlisttest"
)

func TestRollHandler_Roll(t *testing.T) {
	tests := []struct {
		name           string
		store          *collectiontest.MockStore
		wishlist       *wishlisttest.MockStore
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
			wishlist: &wishlisttest.MockStore{},
			cmd: &MockCommandContext{
				UserIDVal:  1,
				GuildIDVal: 2,
			},
			config:        collection.Config{RollCooldown: time.Hour},
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
			wishlist: &wishlisttest.MockStore{},
			cmd: &MockCommandContext{
				UserIDVal:  1,
				GuildIDVal: 2,
			},
			config:        collection.Config{RollCooldown: time.Hour},
			wantContent:   "error occurred",
			wantResponded: true,
		},
		{
			name: "success no wishlist",
			store: &collectiontest.MockStore{
				GetUserFunc: func(ctx context.Context, userID collection.UserID) (collection.User, error) {
					return collection.User{UserID: userID}, nil
				},
				RandomCharNotOwnedFunc: func(ctx context.Context, userID collection.UserID, _ float64) (catalog.Character, error) {
					return catalog.Character{
						ID:         42,
						Name:       "Rem",
						Image:      "https://example.com/rem.png",
						MediaTitle: "Re:Zero",
						Favorites:  5000,
					}, nil
				},
				UpsertCharacterFunc: func(ctx context.Context, char catalog.Character) error { return nil },
				AddToCollectionFunc: func(ctx context.Context, userID collection.UserID, char collection.Character, source string, acquiredAt time.Time) error {
					return nil
				},
				RemoveFromWishlistFunc: func(ctx context.Context, userID collection.UserID, charID int64) error { return nil },
				UpdateLastRollFunc:     func(ctx context.Context, userID collection.UserID, when time.Time) error { return nil },
				CommitFunc:             func(ctx context.Context) error { return nil },
			},
			wishlist: &wishlisttest.MockStore{
				GetUsersWantingCharacterFunc: func(ctx context.Context, charID int64, guildID, excludeUserID uint64) ([]uint64, error) {
					return nil, nil
				},
			},
			cmd: &MockCommandContext{
				UserIDVal:  1,
				GuildIDVal: 2,
			},
			config:         collection.Config{RollCooldown: time.Hour},
			wantEmbedTitle: "Rem",
			wantResponded:  true,
		},
		{
			name: "success with wishlist",
			store: &collectiontest.MockStore{
				GetUserFunc: func(ctx context.Context, userID collection.UserID) (collection.User, error) {
					return collection.User{UserID: userID}, nil
				},
				RandomCharNotOwnedFunc: func(ctx context.Context, userID collection.UserID, _ float64) (catalog.Character, error) {
					return catalog.Character{
						ID:         42,
						Name:       "Rem",
						Image:      "https://example.com/rem.png",
						MediaTitle: "Re:Zero",
						Favorites:  5000,
					}, nil
				},
				UpsertCharacterFunc: func(ctx context.Context, char catalog.Character) error { return nil },
				AddToCollectionFunc: func(ctx context.Context, userID collection.UserID, char collection.Character, source string, acquiredAt time.Time) error {
					return nil
				},
				RemoveFromWishlistFunc: func(ctx context.Context, userID collection.UserID, charID int64) error { return nil },
				UpdateLastRollFunc:     func(ctx context.Context, userID collection.UserID, when time.Time) error { return nil },
				CommitFunc:             func(ctx context.Context) error { return nil },
			},
			wishlist: &wishlisttest.MockStore{
				GetUsersWantingCharacterFunc: func(ctx context.Context, charID int64, guildID, excludeUserID uint64) ([]uint64, error) {
					return []uint64{10, 20, 30}, nil
				},
			},
			cmd: &MockCommandContext{
				UserIDVal:  1,
				GuildIDVal: 2,
			},
			config:         collection.Config{RollCooldown: time.Hour},
			wantEmbedTitle: "Rem",
			wantEmbedDesc:  "<@10> <@20> <@30>",
			wantResponded:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &cordetest.MockResponseWriter{}
			svc := collection.NewRollService(tt.store, collection.RollConfig{RollCooldown: tt.config.RollCooldown})
			h := &RollHandler{
				rollService: svc,
				wishlist:    tt.wishlist,
			}

			h.Roll(t.Context(), w, tt.cmd)

			assert.Equal(t, tt.wantResponded, w.RespondCalled)
			if tt.wantContent != "" {
				w.AssertContains(t, tt.wantContent)
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
