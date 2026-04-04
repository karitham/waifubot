package discord

import (
	"context"
	"errors"
	"testing"

	"github.com/Karitham/corde"
	"github.com/stretchr/testify/assert"

	"github.com/karitham/waifubot/catalog"
	"github.com/karitham/waifubot/collection"
	"github.com/karitham/waifubot/collection/collectiontest"
	"github.com/karitham/waifubot/discord/cordetest"
)

func TestGiveHandler_Give(t *testing.T) {
	tests := []struct {
		name        string
		cmd         CommandContext
		store       *collectiontest.MockStore
		wantContent string
	}{
		{
			name: "missing user option",
			cmd: &MockCommandContext{
				UserIDVal: 1,
				ErrVal:    errors.New("option not found"),
			},
			store:       &collectiontest.MockStore{},
			wantContent: "select a user",
		},
		{
			name: "missing character option",
			cmd: &MockCommandContext{
				UserIDVal: 1,
				OptUserVals: map[string]corde.User{
					"user": {ID: 99, Username: "recipient"},
				},
				ErrVal: errors.New("option not found"),
			},
			store:       &collectiontest.MockStore{},
			wantContent: "select a character",
		},
		{
			name: "don't own character",
			cmd: &MockCommandContext{
				UserIDVal: 1,
				OptUserVals: map[string]corde.User{
					"user": {ID: 99, Username: "recipient"},
				},
				OptInt64Vals: map[string]int64{"id": 42},
			},
			store: &collectiontest.MockStore{
				GetOwnedCharacterFunc: func(ctx context.Context, userID collection.UserID, charID int64) (collection.OwnedCharacter, error) {
					return collection.OwnedCharacter{}, collection.ErrNotFound
				},
			},
			wantContent: "don't own that character",
		},
		{
			name: "generic store error",
			cmd: &MockCommandContext{
				UserIDVal: 1,
				OptUserVals: map[string]corde.User{
					"user": {ID: 99, Username: "recipient"},
				},
				OptInt64Vals: map[string]int64{"id": 42},
			},
			store: &collectiontest.MockStore{
				GetOwnedCharacterFunc: func(ctx context.Context, userID collection.UserID, charID int64) (collection.OwnedCharacter, error) {
					if userID == 1 {
						return collection.OwnedCharacter{}, nil
					}
					return collection.OwnedCharacter{}, collection.ErrNotFound
				},
				GiveCharacterFunc: func(ctx context.Context, from, to collection.UserID, charID int64) (collection.OwnedCharacter, error) {
					return collection.OwnedCharacter{}, errors.New("database on fire")
				},
			},
			wantContent: "Failed to give",
		},
		{
			name: "successful give",
			cmd: &MockCommandContext{
				UserIDVal: 1,
				OptUserVals: map[string]corde.User{
					"user": {ID: 99, Username: "recipient"},
				},
				OptInt64Vals: map[string]int64{"id": 42},
			},
			store: &collectiontest.MockStore{
				GetOwnedCharacterFunc: func(ctx context.Context, userID collection.UserID, charID int64) (collection.OwnedCharacter, error) {
					if userID == 1 {
						return collection.OwnedCharacter{Character: collection.Character{ID: 42, Name: "Sakura"}}, nil
					}
					return collection.OwnedCharacter{}, collection.ErrNotFound
				},
				GiveCharacterFunc: func(ctx context.Context, from, to collection.UserID, charID int64) (collection.OwnedCharacter, error) {
					return collection.OwnedCharacter{Character: collection.Character{ID: 42, Name: "Sakura"}}, nil
				},
				RemoveFromWishlistFunc: func(ctx context.Context, userID collection.UserID, charID int64) error {
					return nil
				},
			},
			wantContent: "Sakura",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &cordetest.MockResponseWriter{}
			h := &GiveHandler{store: tt.store}

			h.Give(t.Context(), w, tt.cmd)

			assert.True(t, w.RespondCalled)
			w.AssertContains(t, tt.wantContent)
		})
	}
}

func TestGiveHandler_Give_RecipientUsername(t *testing.T) {
	w := &cordetest.MockResponseWriter{}
	store := &collectiontest.MockStore{
		GetOwnedCharacterFunc: func(ctx context.Context, userID collection.UserID, charID int64) (collection.OwnedCharacter, error) {
			if userID == 1 {
				return collection.OwnedCharacter{Character: collection.Character{ID: 42, Name: "Sakura"}}, nil
			}
			return collection.OwnedCharacter{}, collection.ErrNotFound
		},
		GiveCharacterFunc: func(ctx context.Context, from, to collection.UserID, charID int64) (collection.OwnedCharacter, error) {
			return collection.OwnedCharacter{Character: collection.Character{ID: 42, Name: "Sakura"}}, nil
		},
		RemoveFromWishlistFunc: func(ctx context.Context, userID collection.UserID, charID int64) error {
			return nil
		},
	}
	cmd := &MockCommandContext{
		UserIDVal: 1,
		OptUserVals: map[string]corde.User{
			"user": {ID: 99, Username: "recipient"},
		},
		OptInt64Vals: map[string]int64{"id": 42},
	}

	h := &GiveHandler{store: store}
	h.Give(t.Context(), w, cmd)

	assert.True(t, w.RespondCalled)
	w.AssertContains(t, "recipient")
}

func TestGiveHandler_Autocomplete(t *testing.T) {
	tests := []struct {
		name            string
		interaction     *corde.Interaction[corde.AutocompleteInteractionData]
		store           *collectiontest.MockStore
		wantCalled      bool
		wantChoiceCount int
	}{
		{
			name: "results found",
			interaction: cordetest.AutocompleteInteraction(
				1, 2, 3, "testuser",
				corde.OptionsInteractions{"id": corde.JsonRaw(`"sak"`)},
			),
			store: &collectiontest.MockStore{
				SearchCharactersFunc: func(ctx context.Context, userID uint64, term string) ([]catalog.Character, error) {
					return []catalog.Character{
						{ID: 10, Name: "Sakura"},
						{ID: 20, Name: "Sakura Haruno"},
					}, nil
				},
			},
			wantCalled:      true,
			wantChoiceCount: 2,
		},
		{
			name: "no results",
			interaction: cordetest.AutocompleteInteraction(
				1, 2, 3, "testuser",
				corde.OptionsInteractions{"id": corde.JsonRaw(`"nonexistent"`)},
			),
			store: &collectiontest.MockStore{
				SearchCharactersFunc: func(ctx context.Context, userID uint64, term string) ([]catalog.Character, error) {
					return nil, nil
				},
			},
			wantCalled:      true,
			wantChoiceCount: 0,
		},
		{
			name: "store error",
			interaction: cordetest.AutocompleteInteraction(
				1, 2, 3, "testuser",
				corde.OptionsInteractions{"id": corde.JsonRaw(`"sak"`)},
			),
			store: &collectiontest.MockStore{
				SearchCharactersFunc: func(ctx context.Context, userID uint64, term string) ([]catalog.Character, error) {
					return nil, errors.New("db error")
				},
			},
			wantCalled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &cordetest.MockResponseWriter{}
			h := &GiveHandler{store: tt.store}

			h.Autocomplete(t.Context(), w, tt.interaction)

			assert.Equal(t, tt.wantCalled, w.AutocompleteCalled)
			if tt.wantCalled {
				choices := w.Choices()
				assert.Len(t, choices, tt.wantChoiceCount)
			}
		})
	}
}
