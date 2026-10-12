package discord

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Karitham/corde"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/karitham/waifubot/catalog"
	"github.com/karitham/waifubot/collection"
	"github.com/karitham/waifubot/collection/collectiontest"
	"github.com/karitham/waifubot/discord/cordetest"
	"github.com/karitham/waifubot/tracker"
)

func tokenAmount(t *testing.T, n int) collection.TokenAmount {
	t.Helper()

	amount, err := collection.NewTokenAmount(n)
	require.NoError(t, err)
	return amount
}

func tokenConfig(t *testing.T) collection.Config {
	t.Helper()

	return collection.Config{
		RollCooldown:   time.Hour,
		SeriesRollCost: tokenAmount(t, 5),
	}
}

func TestTokenHandler_Balance(t *testing.T) {
	tests := []struct {
		name        string
		store       *collectiontest.MockStore
		wantContent string
	}{
		{
			name: "success",
			store: &collectiontest.MockStore{
				GetUserFunc: func(ctx context.Context, userID collection.UserID) (collection.User, error) {
					return collection.User{UserID: userID, Tokens: 42}, nil
				},
			},
			wantContent: "You have 42 tokens",
		},
		{
			name: "fresh user",
			store: &collectiontest.MockStore{
				GetUserFunc: func(ctx context.Context, userID collection.UserID) (collection.User, error) {
					return collection.User{}, collection.ErrNotFound
				},
			},
			wantContent: "You have 0 tokens",
		},
		{
			name: "store error",
			store: &collectiontest.MockStore{
				GetUserFunc: func(ctx context.Context, userID collection.UserID) (collection.User, error) {
					return collection.User{}, errors.New("database on fire")
				},
			},
			wantContent: "Failed to get your balance",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &cordetest.MockResponseWriter{}
			cmd := &MockCommandContext{UserIDVal: 1}
			h := &TokenHandler{store: tt.store}

			h.Balance(t.Context(), w, cmd)

			assert.True(t, w.RespondCalled)
			w.AssertContains(t, tt.wantContent)
		})
	}
}

func TestTokenHandler_Give(t *testing.T) {
	tests := []struct {
		name        string
		cmd         CommandContext
		transferErr error
		wantContent string
	}{
		{
			name: "missing user",
			cmd: &MockCommandContext{
				UserIDVal: 1,
				ErrVal:    errors.New("option not found"),
			},
			wantContent: "select a user",
		},
		{
			name: "missing amount",
			cmd: &MockCommandContext{
				UserIDVal: 1,
				OptUserVals: map[string]corde.User{
					"user": {ID: 99, Username: "recipient"},
				},
				ErrVal: errors.New("option not found"),
			},
			wantContent: "specify a valid amount",
		},
		{
			name: "insufficient tokens",
			cmd: &MockCommandContext{
				UserIDVal: 1,
				OptUserVals: map[string]corde.User{
					"user": {ID: 99, Username: "recipient"},
				},
				OptIntVals: map[string]int{"amount": 10},
			},
			transferErr: collection.ErrInsufficientTokens,
			wantContent: "don't have enough tokens",
		},
		{
			name: "invalid amount",
			cmd: &MockCommandContext{
				UserIDVal: 1,
				OptUserVals: map[string]corde.User{
					"user": {ID: 99, Username: "recipient"},
				},
				OptIntVals: map[string]int{"amount": 10},
			},
			transferErr: collection.ErrInvalidAmount,
			wantContent: "must be positive",
		},
		{
			name: "same user",
			cmd: &MockCommandContext{
				UserIDVal: 1,
				OptUserVals: map[string]corde.User{
					"user": {ID: 1, Username: "self"},
				},
				OptIntVals: map[string]int{"amount": 10},
			},
			transferErr: collection.ErrSameUserTransfer,
			wantContent: "cannot transfer tokens to yourself",
		},
		{
			name: "generic error",
			cmd: &MockCommandContext{
				UserIDVal: 1,
				OptUserVals: map[string]corde.User{
					"user": {ID: 99, Username: "recipient"},
				},
				OptIntVals: map[string]int{"amount": 10},
			},
			transferErr: errors.New("database on fire"),
			wantContent: "Failed to transfer",
		},
		{
			name: "success",
			cmd: &MockCommandContext{
				UserIDVal: 1,
				OptUserVals: map[string]corde.User{
					"user": {ID: 99, Username: "recipient"},
				},
				OptIntVals: map[string]int{"amount": 10},
			},
			wantContent: "Gave 10 tokens to recipient",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &cordetest.MockResponseWriter{}
			store := &collectiontest.MockStore{}
			if tt.transferErr != nil {
				store.AddTokensFunc = func(ctx context.Context, userID collection.UserID, amount collection.TokenAmount) (collection.User, error) {
					return collection.User{}, tt.transferErr
				}
			}
			h := &TokenHandler{store: store}

			h.Give(t.Context(), w, tt.cmd)

			assert.True(t, w.RespondCalled)
			w.AssertContains(t, tt.wantContent)
		})
	}
}

func TestTokenHandler_Sell(t *testing.T) {
	tests := []struct {
		name        string
		cmd         CommandContext
		store       *collectiontest.MockStore
		wantContent string
	}{
		{
			name: "missing character",
			cmd: &MockCommandContext{
				UserIDVal: 1,
				ErrVal:    errors.New("option not found"),
			},
			wantContent: "select a character",
		},
		{
			name: "don't own character",
			cmd: &MockCommandContext{
				UserIDVal:  1,
				OptIntVals: map[string]int{"id": 42},
			},
			store: &collectiontest.MockStore{
				GetOwnedCharacterFunc: func(ctx context.Context, userID collection.UserID, charID catalog.CharacterID) (collection.OwnedCharacter, error) {
					return collection.OwnedCharacter{}, collection.ErrNotFound
				},
			},
			wantContent: "don't own that character",
		},
		{
			name: "generic error",
			cmd: &MockCommandContext{
				UserIDVal:  1,
				OptIntVals: map[string]int{"id": 42},
			},
			store: &collectiontest.MockStore{
				GetOwnedCharacterFunc: func(ctx context.Context, userID collection.UserID, charID catalog.CharacterID) (collection.OwnedCharacter, error) {
					return collection.OwnedCharacter{}, nil
				},
				RemoveFromCollectionFunc: func(ctx context.Context, userID collection.UserID, charID catalog.CharacterID) error {
					return errors.New("database on fire")
				},
			},
			wantContent: "Failed to sell",
		},
		{
			name: "success",
			cmd: &MockCommandContext{
				UserIDVal:  1,
				OptIntVals: map[string]int{"id": 42},
			},
			store: &collectiontest.MockStore{
				GetCharacterByIDFunc: func(ctx context.Context, charID catalog.CharacterID) (catalog.Character, error) {
					return catalog.Character{ID: 42, Name: "Sakura"}, nil
				},
				GetOwnedCharacterFunc: func(ctx context.Context, userID collection.UserID, charID catalog.CharacterID) (collection.OwnedCharacter, error) {
					return collection.OwnedCharacter{ID: 42, Name: "Sakura"}, nil
				},
				RemoveFromCollectionFunc: func(ctx context.Context, userID collection.UserID, charID catalog.CharacterID) error {
					return nil
				},
				AddTokensFunc: func(ctx context.Context, userID collection.UserID, amount collection.TokenAmount) (collection.User, error) {
					return collection.User{UserID: userID}, nil
				},
			},
			wantContent: "Sold Sakura",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &cordetest.MockResponseWriter{}
			h := &TokenHandler{store: tt.store}

			h.Sell(t.Context(), w, tt.cmd)

			assert.True(t, w.RespondCalled)
			w.AssertContains(t, tt.wantContent)
		})
	}
}

func TestTokenHandler_Roll(t *testing.T) {
	tests := []struct {
		name         string
		cmd          CommandContext
		store        *collectiontest.MockStore
		animeService *collectiontest.MockAnimeService
		wantContent  string
		wantEmbed    bool
	}{
		{
			name: "missing series",
			cmd: &MockCommandContext{
				UserIDVal: 1,
				ErrVal:    errors.New("option not found"),
			},
			store:        &collectiontest.MockStore{},
			animeService: &collectiontest.MockAnimeService{},
			wantContent:  "select a series",
		},
		{
			name: "insufficient tokens",
			cmd: &MockCommandContext{
				UserIDVal:    1,
				OptInt64Vals: map[string]int64{"series": 100},
			},
			store: &collectiontest.MockStore{
				GetUserFunc: func(ctx context.Context, userID collection.UserID) (collection.User, error) {
					return collection.User{UserID: userID, Tokens: 2}, nil
				},
			},
			animeService: &collectiontest.MockAnimeService{},
			wantContent:  "need 5 tokens",
		},
		{
			name: "already owned all",
			cmd: &MockCommandContext{
				UserIDVal:    1,
				OptInt64Vals: map[string]int64{"series": 100},
			},
			store: &collectiontest.MockStore{
				GetUserFunc: func(ctx context.Context, userID collection.UserID) (collection.User, error) {
					return collection.User{UserID: userID, Tokens: 10}, nil
				},
				GetCollectionIDsFunc: func(ctx context.Context, userID collection.UserID) ([]catalog.CharacterID, error) {
					return []catalog.CharacterID{1, 2, 3}, nil
				},
			},
			animeService: &collectiontest.MockAnimeService{
				GetMediaCharactersFunc: func(ctx context.Context, mediaID tracker.MediaID) ([]tracker.Character, error) {
					return []tracker.Character{
						{ID: 1, Name: "Char1"},
						{ID: 2, Name: "Char2"},
						{ID: 3, Name: "Char3"},
					}, nil
				},
			},
			wantContent: "already own all characters",
		},
		{
			name: "media not found",
			cmd: &MockCommandContext{
				UserIDVal:    1,
				OptInt64Vals: map[string]int64{"series": 100},
			},
			store: &collectiontest.MockStore{
				GetUserFunc: func(ctx context.Context, userID collection.UserID) (collection.User, error) {
					return collection.User{UserID: userID, Tokens: 10}, nil
				},
			},
			animeService: &collectiontest.MockAnimeService{
				GetMediaCharactersFunc: func(ctx context.Context, mediaID tracker.MediaID) ([]tracker.Character, error) {
					return nil, nil
				},
			},
			wantContent: "No characters found",
		},
		{
			name: "generic error",
			cmd: &MockCommandContext{
				UserIDVal:    1,
				OptInt64Vals: map[string]int64{"series": 100},
			},
			store: &collectiontest.MockStore{
				GetUserFunc: func(ctx context.Context, userID collection.UserID) (collection.User, error) {
					return collection.User{UserID: userID, Tokens: 10}, nil
				},
			},
			animeService: &collectiontest.MockAnimeService{
				GetMediaCharactersFunc: func(ctx context.Context, mediaID tracker.MediaID) ([]tracker.Character, error) {
					return nil, errors.New("database on fire")
				},
			},
			wantContent: "error occurred",
		},
		{
			name: "success",
			cmd: &MockCommandContext{
				UserIDVal:    1,
				OptInt64Vals: map[string]int64{"series": 100},
			},
			store: &collectiontest.MockStore{
				GetUserFunc: func(ctx context.Context, userID collection.UserID) (collection.User, error) {
					return collection.User{UserID: userID, Tokens: 10}, nil
				},
				GetCollectionIDsFunc: func(ctx context.Context, userID collection.UserID) ([]catalog.CharacterID, error) {
					return nil, nil
				},
				UpsertCharacterFunc: func(ctx context.Context, char catalog.Character) error {
					return nil
				},
				AddToCollectionFunc: func(ctx context.Context, userID collection.UserID, charID catalog.CharacterID, source collection.Source, acquiredAt time.Time) error {
					return nil
				},
				RecordRollFunc: func(ctx context.Context, userID collection.UserID, now time.Time, cooldown time.Duration) error {
					return nil
				},
				SpendTokensFunc: func(ctx context.Context, userID collection.UserID, amount collection.TokenAmount) (collection.User, error) {
					return collection.User{UserID: userID}, nil
				},
			},
			animeService: &collectiontest.MockAnimeService{
				GetMediaCharactersFunc: func(ctx context.Context, mediaID tracker.MediaID) ([]tracker.Character, error) {
					return []tracker.Character{
						{ID: 10, Name: "Sakura", MediaTitle: "Naruto", Popularity: 5000},
					}, nil
				},
			},
			wantEmbed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &cordetest.MockResponseWriter{}
			config := tokenConfig(t)
			h := &TokenHandler{store: tt.store, animeService: tt.animeService, config: config}

			h.Roll(t.Context(), w, tt.cmd)

			assert.True(t, w.RespondCalled)
			if tt.wantEmbed {
				data := w.LastRespond.InteractionRespData()
				assert.Len(t, data.Embeds, 1)
			} else if tt.wantContent != "" {
				w.AssertContains(t, tt.wantContent)
			}
		})
	}
}

func TestTokenHandler_SeriesAutocomplete(t *testing.T) {
	tests := []struct {
		name            string
		interaction     *corde.Interaction[corde.AutocompleteInteractionData]
		animeService    *collectiontest.MockAnimeService
		wantCalled      bool
		wantChoiceCount int
	}{
		{
			name: "results found",
			interaction: cordetest.AutocompleteInteraction(
				1, 2, 3, "testuser",
				corde.OptionsInteractions{"series": corde.JsonRaw(`"nar"`)},
			),
			animeService: &collectiontest.MockAnimeService{
				SearchMediaFunc: func(ctx context.Context, search string) ([]tracker.Media, error) {
					return []tracker.Media{
						{ID: 10, Title: "Naruto", Type: "ANIME"},
						{ID: 20, Title: "Naruto Shippuden", Type: "ANIME"},
					}, nil
				},
			},
			wantCalled:      true,
			wantChoiceCount: 2,
		},
		{
			name: "no input",
			interaction: cordetest.AutocompleteInteraction(
				1, 2, 3, "testuser",
				corde.OptionsInteractions{},
			),
			animeService:    &collectiontest.MockAnimeService{},
			wantCalled:      true,
			wantChoiceCount: 0,
		},
		{
			name: "store error",
			interaction: cordetest.AutocompleteInteraction(
				1, 2, 3, "testuser",
				corde.OptionsInteractions{"series": corde.JsonRaw(`"nar"`)},
			),
			animeService: &collectiontest.MockAnimeService{
				SearchMediaFunc: func(ctx context.Context, search string) ([]tracker.Media, error) {
					return nil, errors.New("api error")
				},
			},
			wantCalled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &cordetest.MockResponseWriter{}
			h := &TokenHandler{animeService: tt.animeService}

			h.SeriesAutocomplete(t.Context(), w, tt.interaction)

			assert.Equal(t, tt.wantCalled, w.AutocompleteCalled)
			if tt.wantCalled {
				choices := w.Choices()
				assert.Len(t, choices, tt.wantChoiceCount)
			}
		})
	}
}

func TestTokenHandler_UserCollectionAutocomplete(t *testing.T) {
	tests := []struct {
		name        string
		interaction *corde.Interaction[corde.AutocompleteInteractionData]
		store       *collectiontest.MockStore
		wantCalled  bool
	}{
		{
			name: "results found",
			interaction: cordetest.AutocompleteInteraction(
				1, 2, 3, "testuser",
				corde.OptionsInteractions{"id": corde.JsonRaw(`"sak"`)},
			),
			store: &collectiontest.MockStore{
				SearchCharactersFunc: func(ctx context.Context, userID collection.UserID, term string) ([]catalog.Character, error) {
					return []catalog.Character{
						{ID: 10, Name: "Sakura"},
						{ID: 20, Name: "Sakura Haruno"},
					}, nil
				},
			},
			wantCalled: true,
		},
		{
			name: "store error",
			interaction: cordetest.AutocompleteInteraction(
				1, 2, 3, "testuser",
				corde.OptionsInteractions{"id": corde.JsonRaw(`"sak"`)},
			),
			store: &collectiontest.MockStore{
				SearchCharactersFunc: func(ctx context.Context, userID collection.UserID, term string) ([]catalog.Character, error) {
					return nil, errors.New("db error")
				},
			},
			wantCalled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &cordetest.MockResponseWriter{}
			h := &TokenHandler{store: tt.store}

			h.userCollectionAutocomplete(t.Context(), w, tt.interaction)

			assert.Equal(t, tt.wantCalled, w.AutocompleteCalled)
			if tt.wantCalled {
				choices := w.Choices()
				assert.NotEmpty(t, choices)
			}
		})
	}
}
