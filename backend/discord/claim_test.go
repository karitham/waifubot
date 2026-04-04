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
)

func TestClaimHandler_Claim(t *testing.T) {
	tests := []struct {
		name              string
		cmd               CommandContext
		store             *collectiontest.MockStore
		wantContent       string
		wantEmbedTitle    string
		wantRespondCalled bool
	}{
		{
			name: "no name provided",
			cmd: &MockCommandContext{
				UserIDVal:     1,
				ChannelIDVal:  2,
				GuildIDVal:    3,
				OptStringVals: map[string]string{"name": ""},
			},
			store:             &collectiontest.MockStore{},
			wantContent:       "enter a name to claim the character",
			wantRespondCalled: true,
		},
		{
			name: "no drop in channel",
			cmd: &MockCommandContext{
				UserIDVal:     1,
				ChannelIDVal:  2,
				GuildIDVal:    3,
				OptStringVals: map[string]string{"name": "Sakura"},
			},
			store: &collectiontest.MockStore{
				GetDropForUpdateFunc: func(ctx context.Context, channelID uint64) (collection.Drop, error) {
					return collection.Drop{}, collection.ErrNotFound
				},
			},
			wantContent:       "No character to claim in this channel",
			wantRespondCalled: true,
		},
		{
			name: "wrong character name",
			cmd: &MockCommandContext{
				UserIDVal:     1,
				ChannelIDVal:  2,
				GuildIDVal:    3,
				OptStringVals: map[string]string{"name": "Sakura"},
			},
			store: &collectiontest.MockStore{
				GetDropForUpdateFunc: func(ctx context.Context, channelID uint64) (collection.Drop, error) {
					return collection.Drop{ID: 42, Name: "Naruto"}, nil
				},
			},
			wantContent:       "Wrong name! Check the hint",
			wantRespondCalled: true,
		},
		{
			name: "already owned",
			cmd: &MockCommandContext{
				UserIDVal:     1,
				ChannelIDVal:  2,
				GuildIDVal:    3,
				OptStringVals: map[string]string{"name": "Sakura"},
			},
			store: &collectiontest.MockStore{
				GetDropForUpdateFunc: func(ctx context.Context, channelID uint64) (collection.Drop, error) {
					return collection.Drop{ID: 42, Name: "Sakura"}, nil
				},
				AddToCollectionFunc: func(ctx context.Context, userID collection.UserID, char collection.Character, source string, acquiredAt time.Time) error {
					return collection.ErrAlreadyOwned
				},
			},
			wantContent:       "You already have this character",
			wantRespondCalled: true,
		},
		{
			name: "generic store error",
			cmd: &MockCommandContext{
				UserIDVal:     1,
				ChannelIDVal:  2,
				GuildIDVal:    3,
				OptStringVals: map[string]string{"name": "Sakura"},
			},
			store: &collectiontest.MockStore{
				GetDropForUpdateFunc: func(ctx context.Context, channelID uint64) (collection.Drop, error) {
					return collection.Drop{}, errors.New("database on fire")
				},
			},
			wantContent:       "Failed to claim character",
			wantRespondCalled: true,
		},
		{
			name: "successful claim",
			cmd: &MockCommandContext{
				UserIDVal:     1,
				ChannelIDVal:  2,
				GuildIDVal:    3,
				OptStringVals: map[string]string{"name": "Sakura"},
			},
			store: &collectiontest.MockStore{
				GetDropForUpdateFunc: func(ctx context.Context, channelID uint64) (collection.Drop, error) {
					return collection.Drop{ID: 42, Name: "Sakura", Image: "https://example.com/sakura.png", Favorites: 1000}, nil
				},
				AddToCollectionFunc: func(ctx context.Context, userID collection.UserID, char collection.Character, source string, acquiredAt time.Time) error {
					return nil
				},
				DeleteDropFunc: func(ctx context.Context, channelID uint64) error {
					return nil
				},
				CommitFunc: func(ctx context.Context) error {
					return nil
				},
			},
			wantEmbedTitle:    "Sakura",
			wantRespondCalled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &cordetest.MockResponseWriter{}
			h := &ClaimHandler{store: tt.store}

			h.Claim(t.Context(), w, tt.cmd)

			assert.Equal(t, tt.wantRespondCalled, w.RespondCalled)
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
		})
	}
}
