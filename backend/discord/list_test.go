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

func TestListHandler_List(t *testing.T) {
	tests := []struct {
		name           string
		cmd            CommandContext
		store          *collectiontest.MockStore
		wantContent    string
		wantEmbedTitle string
	}{
		{
			name: "empty collection",
			cmd: &MockCommandContext{
				UserIDVal:    1,
				UsernameVal:  "testuser",
				AvatarPNGVal: "https://example.com/avatar.png",
			},
			store: &collectiontest.MockStore{
				GetCollectionFunc: func(ctx context.Context, userID collection.UserID) ([]collection.OwnedCharacter, error) {
					return nil, nil
				},
			},
			wantContent: "No characters in collection",
		},
		{
			name: "collection with characters",
			cmd: &MockCommandContext{
				UserIDVal:    1,
				UsernameVal:  "testuser",
				AvatarPNGVal: "https://example.com/avatar.png",
			},
			store: &collectiontest.MockStore{
				GetCollectionFunc: func(ctx context.Context, userID collection.UserID) ([]collection.OwnedCharacter, error) {
					return []collection.OwnedCharacter{
						{
							Character: collection.Character{
								ID:    42,
								Name:  "Sakura",
								Image: "https://example.com/sakura.png",
							},
							Date:   time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
							Source: "drop",
						},
					}, nil
				},
			},
			wantEmbedTitle: "testuser's List",
		},
		{
			name: "store error",
			cmd: &MockCommandContext{
				UserIDVal:   1,
				UsernameVal: "testuser",
			},
			store: &collectiontest.MockStore{
				GetCollectionFunc: func(ctx context.Context, userID collection.UserID) ([]collection.OwnedCharacter, error) {
					return nil, errors.New("db error")
				},
			},
			wantContent: "An error occurred dialing the database, please try again later",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &cordetest.MockResponseWriter{}
			h := &ListHandler{store: tt.store}

			h.List(t.Context(), w, tt.cmd)

			assert.True(t, w.RespondCalled)
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
