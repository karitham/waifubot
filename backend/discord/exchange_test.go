package discord

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Karitham/corde"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/karitham/waifubot/catalog"
	"github.com/karitham/waifubot/collection"
	"github.com/karitham/waifubot/collection/collectiontest"
	"github.com/karitham/waifubot/discord/cordetest"
)

func TestRollHandler_Exchange(t *testing.T) {
	t.Run("rolling user exchanges the character", func(t *testing.T) {
		store := storeWithOwnedCharacter()
		cleaner := &recordingCleaner{called: make(chan struct{})}
		h := &RollHandler{store: store, cleaner: cleaner}
		w := &cordetest.MockResponseWriter{}
		i := buttonInteraction(1, exchangeCustomID(1, 42))
		i.Token = "click-token"

		h.Exchange(t.Context(), w, i)

		assert.False(t, w.RespondCalled)
		assert.True(t, w.DeferedUpdateCalled)
		assert.Equal(t, 1, store.CommitCalls)
		requireCleanerCalled(t, cleaner)
		assert.Equal(t, "click-token", cleaner.token)
		assert.Contains(t, cleaner.content, "Sold Rem for 1 token")
	})

	t.Run("another user cannot exchange the character", func(t *testing.T) {
		store := storeWithOwnedCharacter()
		h := &RollHandler{store: store}
		w := &cordetest.MockResponseWriter{}

		h.Exchange(t.Context(), w, buttonInteraction(2, exchangeCustomID(1, 42)))

		assert.False(t, w.DeferedUpdateCalled)
		require.True(t, w.RespondCalled)
		assert.Contains(t, w.LastRespond.InteractionRespData().Content, "Only the user who rolled")
		assert.Equal(t, 0, store.CommitCalls)
	})

	t.Run("character no longer owned", func(t *testing.T) {
		store := storeWithOwnedCharacter()
		store.GetOwnedCharacterFunc = func(ctx context.Context, userID collection.UserID, charID catalog.CharacterID) (collection.OwnedCharacter, error) {
			return collection.OwnedCharacter{}, fmt.Errorf("%w %d", collection.ErrUserDoesNotOwnCharacter, charID)
		}
		h := &RollHandler{store: store}
		w := &cordetest.MockResponseWriter{}

		h.Exchange(t.Context(), w, buttonInteraction(1, exchangeCustomID(1, 42)))

		assert.False(t, w.DeferedUpdateCalled)
		require.True(t, w.RespondCalled)
		assert.Contains(t, w.LastRespond.InteractionRespData().Content, "don't own that character")
	})

	t.Run("malformed custom ID", func(t *testing.T) {
		h := &RollHandler{store: storeWithOwnedCharacter()}
		w := &cordetest.MockResponseWriter{}

		h.Exchange(t.Context(), w, buttonInteraction(1, "exchange:not-a-user"))

		assert.False(t, w.DeferedUpdateCalled)
		require.True(t, w.RespondCalled)
		assert.Contains(t, w.LastRespond.InteractionRespData().Content, "no longer valid")
	})
}

func TestRollHandler_ExchangeDMCaller(t *testing.T) {
	cleaner := &recordingCleaner{called: make(chan struct{})}
	h := &RollHandler{store: storeWithOwnedCharacter(), cleaner: cleaner}
	w := &cordetest.MockResponseWriter{}
	i := buttonInteraction(1, exchangeCustomID(1, 42))
	i.Member = corde.Member{}
	i.User = &corde.User{ID: 1}

	h.Exchange(t.Context(), w, i)

	require.True(t, w.DeferedUpdateCalled)
	requireCleanerCalled(t, cleaner)
}

func TestParseExchangeCustomID(t *testing.T) {
	tests := []struct {
		name     string
		customID string
		owner    collection.UserID
		charID   catalog.CharacterID
		wantErr  bool
	}{
		{name: "valid", customID: "exchange:1:42", owner: 1, charID: 42},
		{name: "wrong prefix", customID: "sell:1:42", wantErr: true},
		{name: "missing character", customID: "exchange:1", wantErr: true},
		{name: "bad owner", customID: "exchange:nope:42", wantErr: true},
		{name: "bad character", customID: "exchange:1:nope", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			owner, charID, err := parseExchangeCustomID(tt.customID)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.owner, owner)
			assert.Equal(t, tt.charID, charID)
		})
	}
}

func storeWithOwnedCharacter() *collectiontest.MockStore {
	return &collectiontest.MockStore{
		GetCharacterByIDFunc: func(ctx context.Context, charID catalog.CharacterID) (catalog.Character, error) {
			return catalog.Character{ID: charID, Name: "Rem"}, nil
		},
		GetOwnedCharacterFunc: func(ctx context.Context, userID collection.UserID, charID catalog.CharacterID) (collection.OwnedCharacter, error) {
			return collection.OwnedCharacter{}, nil
		},
	}
}

func buttonInteraction(userID uint64, customID string) *corde.Interaction[corde.ButtonInteractionData] {
	return &corde.Interaction[corde.ButtonInteractionData]{
		Member: corde.Member{User: corde.User{ID: corde.Snowflake(userID)}},
		Data: corde.ButtonInteractionData{
			CustomID:      customID,
			ComponentType: corde.COMPONENT_BUTTON,
		},
	}
}

func requireCleanerCalled(t *testing.T, cleaner *recordingCleaner) {
	t.Helper()
	select {
	case <-cleaner.called:
	case <-time.After(time.Second):
		t.Fatal("exchange button was not cleared")
	}
}

// recordingCleaner captures the request to strip the used button.
type recordingCleaner struct {
	called  chan struct{}
	token   string
	content string
}

func (c *recordingCleaner) ClearComponents(token string, content string) error {
	c.token = token
	c.content = content
	close(c.called)
	return nil
}
