package discord

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Karitham/corde"
	"github.com/stretchr/testify/assert"

	"github.com/karitham/waifubot/catalog"
	"github.com/karitham/waifubot/collection"
	"github.com/karitham/waifubot/collection/collectiontest"
	"github.com/karitham/waifubot/discord/cordetest"
	"github.com/karitham/waifubot/guild"
)

func TestRollHandler_Roll(t *testing.T) {
	tests := []struct {
		name               string
		store              *collectiontest.MockStore
		cmd                CommandContext
		config             collection.Config
		wantContent        string
		wantEmbedTitle     string
		wantEmbedDesc      string
		wantResponded      bool
		wantExchangePrompt bool
	}{
		{
			name: "roll cooldown",
			store: &collectiontest.MockStore{
				GetUserFunc: func(ctx context.Context, userID collection.UserID) (collection.User, error) {
					return collection.User{
						UserID:   userID,
						LastRoll: time.Now().Add(30 * time.Minute),
					}, nil
				},
			},
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
						Popularity: 5000,
					}, nil
				},
				UpsertCharacterFunc: func(ctx context.Context, char catalog.Character) error { return nil },
				AddToCollectionFunc: func(ctx context.Context, userID collection.UserID, charID catalog.CharacterID, source collection.Source, acquiredAt time.Time) error {
					return nil
				},
				RemoveFromWishlistFunc: func(ctx context.Context, userID collection.UserID, charID catalog.CharacterID) error { return nil },
				RecordRollFunc: func(ctx context.Context, userID collection.UserID, now time.Time, cooldown time.Duration) error {
					return nil
				},
				CommitFunc: func(ctx context.Context) error { return nil },
				GetUsersWantingCharacterFunc: func(ctx context.Context, charID catalog.CharacterID, guildID guild.ID, excludeUserID collection.UserID) ([]guild.MemberID, error) {
					return nil, nil
				},
			},
			cmd: &MockCommandContext{
				UserIDVal:  1,
				TokenVal:   "roll-token",
				GuildIDVal: 2,
			},
			config:             collection.Config{RollCooldown: time.Hour},
			wantEmbedTitle:     "Rem",
			wantResponded:      true,
			wantExchangePrompt: true,
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
						Popularity: 5000,
					}, nil
				},
				UpsertCharacterFunc: func(ctx context.Context, char catalog.Character) error { return nil },
				AddToCollectionFunc: func(ctx context.Context, userID collection.UserID, charID catalog.CharacterID, source collection.Source, acquiredAt time.Time) error {
					return nil
				},
				RemoveFromWishlistFunc: func(ctx context.Context, userID collection.UserID, charID catalog.CharacterID) error { return nil },
				RecordRollFunc: func(ctx context.Context, userID collection.UserID, now time.Time, cooldown time.Duration) error {
					return nil
				},
				CommitFunc: func(ctx context.Context) error { return nil },
				GetUsersWantingCharacterFunc: func(ctx context.Context, charID catalog.CharacterID, guildID guild.ID, excludeUserID collection.UserID) ([]guild.MemberID, error) {
					return []guild.MemberID{10, 20, 30}, nil
				},
			},
			cmd: &MockCommandContext{
				UserIDVal:  1,
				TokenVal:   "roll-token",
				GuildIDVal: 2,
			},
			config:             collection.Config{RollCooldown: time.Hour},
			wantEmbedTitle:     "Rem",
			wantEmbedDesc:      "<@10> <@20> <@30>",
			wantResponded:      true,
			wantExchangePrompt: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &cordetest.MockResponseWriter{}
			followUp := &recordingFollowUp{called: make(chan struct{})}
			h := &RollHandler{
				store:    tt.store,
				config:   tt.config,
				followUp: followUp,
			}

			h.Roll(t.Context(), w, tt.cmd)

			assert.Equal(t, tt.wantResponded, w.RespondCalled)
			if !tt.wantExchangePrompt {
				select {
				case <-followUp.called:
					t.Fatal("exchange prompt sent for a failed roll")
				default:
				}
			}
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
			if tt.wantExchangePrompt {
				select {
				case <-followUp.called:
				case <-time.After(time.Second):
					t.Fatal("exchange prompt was not sent")
				}
				assert.Equal(t, tt.cmd.Token(), followUp.token)
				if assert.NotNil(t, followUp.data) {
					data := followUp.data.InteractionRespData()
					assert.Equal(t, corde.RESPONSE_FLAGS_EPHEMERAL, data.Flags)
					if assert.Len(t, data.Components, 1) {
						row := data.Components[0]
						assert.Equal(t, corde.COMPONENT_ACTION_ROW, row.Type)
						if assert.Len(t, row.Components, 1) {
							assert.Equal(t, exchangeCustomID(tt.cmd.UserID(), 42), row.Components[0].CustomID)
						}
					}
				}
			}
		})
	}
}

// recordingFollowUp captures the ephemeral exchange prompt sent after a roll.
type recordingFollowUp struct {
	called chan struct{}
	token  string
	data   corde.InteractionResponder
}

func (f *recordingFollowUp) FollowUpInteraction(token string, data corde.InteractionResponder) error {
	f.token = token
	f.data = data
	close(f.called)
	return nil
}
