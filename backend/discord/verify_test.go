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

func TestVerifyHandler_Verify(t *testing.T) {
	tests := []struct {
		name              string
		cmd               CommandContext
		store             *collectiontest.MockStore
		wantContent       string
		wantRespondCalled bool
	}{
		{
			name: "missing character ID defaults to zero",
			cmd: &MockCommandContext{
				UserIDVal:   1,
				UsernameVal: "testuser",
				ErrVal:      errors.New("option not found"),
			},
			store: &collectiontest.MockStore{
				GetOwnedCharacterFunc: func(ctx context.Context, userID collection.UserID, charID int64) (collection.OwnedCharacter, error) {
					return collection.OwnedCharacter{}, collection.ErrNotFound
				},
			},
			wantContent:       "testuser doesn't have this character",
			wantRespondCalled: true,
		},
		{
			name: "user has character",
			cmd: &MockCommandContext{
				UserIDVal:    1,
				UsernameVal:  "testuser",
				OptInt64Vals: map[string]int64{"id": 42},
			},
			store: &collectiontest.MockStore{
				GetOwnedCharacterFunc: func(ctx context.Context, userID collection.UserID, charID int64) (collection.OwnedCharacter, error) {
					return collection.OwnedCharacter{
						Character: collection.Character{
							ID:   42,
							Name: "Sakura",
						},
					}, nil
				},
			},
			wantContent:       "testuser has Sakura in their collection",
			wantRespondCalled: true,
		},
		{
			name: "user doesn't have character",
			cmd: &MockCommandContext{
				UserIDVal:    1,
				UsernameVal:  "testuser",
				OptInt64Vals: map[string]int64{"id": 42},
			},
			store: &collectiontest.MockStore{
				GetOwnedCharacterFunc: func(ctx context.Context, userID collection.UserID, charID int64) (collection.OwnedCharacter, error) {
					return collection.OwnedCharacter{}, collection.ErrNotFound
				},
			},
			wantContent:       "testuser doesn't have this character",
			wantRespondCalled: true,
		},
		{
			name: "store error",
			cmd: &MockCommandContext{
				UserIDVal:    1,
				UsernameVal:  "testuser",
				OptInt64Vals: map[string]int64{"id": 42},
			},
			store: &collectiontest.MockStore{
				GetOwnedCharacterFunc: func(ctx context.Context, userID collection.UserID, charID int64) (collection.OwnedCharacter, error) {
					return collection.OwnedCharacter{}, errors.New("database on fire")
				},
			},
			wantContent:       "Error checking ownership",
			wantRespondCalled: true,
		},
		{
			name: "other user resolved",
			cmd: &MockCommandContext{
				UserIDVal:       1,
				UsernameVal:     "testuser",
				OptInt64Vals:    map[string]int64{"id": 42},
				HasResolvedUser: true,
				FirstResolvedUserVal: corde.User{
					ID:       99,
					Username: "alice",
				},
			},
			store: &collectiontest.MockStore{
				GetOwnedCharacterFunc: func(ctx context.Context, userID collection.UserID, charID int64) (collection.OwnedCharacter, error) {
					return collection.OwnedCharacter{
						Character: collection.Character{
							ID:   42,
							Name: "Naruto",
						},
					}, nil
				},
			},
			wantContent:       "alice has Naruto in their collection",
			wantRespondCalled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &cordetest.MockResponseWriter{}
			h := &VerifyHandler{store: tt.store}

			h.Verify(t.Context(), w, tt.cmd)

			assert.Equal(t, tt.wantRespondCalled, w.RespondCalled)
			if tt.wantContent != "" {
				w.AssertContains(t, tt.wantContent)
			}
		})
	}
}

func TestVerifyHandler_Autocomplete(t *testing.T) {
	tests := []struct {
		name                 string
		interaction          *corde.Interaction[corde.AutocompleteInteractionData]
		store                *collectiontest.MockStore
		wantAutocompleteCall bool
		wantChoiceCount      int
	}{
		{
			name: "results found",
			interaction: cordetest.AutocompleteInteraction(
				1, 2, 3, "testuser",
				corde.OptionsInteractions{"id": corde.JsonRaw(`"sak"`)},
			),
			store: &collectiontest.MockStore{
				SearchGlobalCharactersFunc: func(ctx context.Context, term string) ([]catalog.Character, error) {
					return []catalog.Character{
						{ID: 1, Name: "Sakura"},
						{ID: 2, Name: "Sakura Haruno"},
					}, nil
				},
			},
			wantAutocompleteCall: true,
			wantChoiceCount:      2,
		},
		{
			name: "no results",
			interaction: cordetest.AutocompleteInteraction(
				1, 2, 3, "testuser",
				corde.OptionsInteractions{"id": corde.JsonRaw(`"xyz"`)},
			),
			store: &collectiontest.MockStore{
				SearchGlobalCharactersFunc: func(ctx context.Context, term string) ([]catalog.Character, error) {
					return nil, nil
				},
			},
			wantAutocompleteCall: true,
			wantChoiceCount:      0,
		},
		{
			name: "store error",
			interaction: cordetest.AutocompleteInteraction(
				1, 2, 3, "testuser",
				corde.OptionsInteractions{"id": corde.JsonRaw(`"sak"`)},
			),
			store: &collectiontest.MockStore{
				SearchGlobalCharactersFunc: func(ctx context.Context, term string) ([]catalog.Character, error) {
					return nil, errors.New("db error")
				},
			},
			wantAutocompleteCall: false,
			wantChoiceCount:      0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &cordetest.MockResponseWriter{}
			h := &VerifyHandler{store: tt.store}

			h.Autocomplete(t.Context(), w, tt.interaction)

			assert.Equal(t, tt.wantAutocompleteCall, w.AutocompleteCalled)
			if tt.wantAutocompleteCall {
				choices := w.Choices()
				assert.Len(t, choices, tt.wantChoiceCount)
			}
		})
	}
}
