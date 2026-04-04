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
)

func TestProfileHandler_View(t *testing.T) {
	tests := []struct {
		name        string
		cmd         CommandContext
		store       *collectiontest.MockStore
		wantContent string
		wantTitle   string
	}{
		{
			name: "own profile",
			cmd: &MockCommandContext{
				UserIDVal:   1,
				UsernameVal: "testuser",
			},
			store: &collectiontest.MockStore{
				GetUserFunc: func(ctx context.Context, userID collection.UserID) (collection.User, error) {
					return collection.User{
						UserID:     1,
						Tokens:     42,
						Quote:      "hello world",
						Favorite:   10,
						AnilistURL: "https://anilist.co/user/testuser",
						Date:       time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
					}, nil
				},
				GetCharacterByIDFunc: func(ctx context.Context, charID int64) (catalog.Character, error) {
					return catalog.Character{ID: 10, Name: "Sakura", Image: "https://example.com/sakura.png"}, nil
				},
				CountCollectionFunc: func(ctx context.Context, userID collection.UserID) (int64, error) {
					return 5, nil
				},
			},
			wantTitle: "testuser",
		},
		{
			name: "other user profile",
			cmd: &MockCommandContext{
				UserIDVal:   1,
				UsernameVal: "testuser",
				FirstResolvedUserVal: corde.User{
					ID:       99,
					Username: "otheruser",
				},
				HasResolvedUser: true,
			},
			store: &collectiontest.MockStore{
				GetUserFunc: func(ctx context.Context, userID collection.UserID) (collection.User, error) {
					return collection.User{
						UserID: 99,
						Tokens: 10,
						Date:   time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
					}, nil
				},
				CountCollectionFunc: func(ctx context.Context, userID collection.UserID) (int64, error) {
					return 3, nil
				},
			},
			wantTitle: "otheruser",
		},
		{
			name: "store error",
			cmd: &MockCommandContext{
				UserIDVal:   1,
				UsernameVal: "testuser",
			},
			store: &collectiontest.MockStore{
				GetUserFunc: func(ctx context.Context, userID collection.UserID) (collection.User, error) {
					return collection.User{}, errors.New("db error")
				},
			},
			wantContent: "An error occurred dialing the database, please try again later",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &cordetest.MockResponseWriter{}
			h := &ProfileHandler{store: tt.store}

			h.View(t.Context(), w, tt.cmd)

			assert.True(t, w.RespondCalled)
			if tt.wantContent != "" {
				data := w.LastRespond.InteractionRespData()
				assert.Contains(t, data.Content, tt.wantContent)
			}
			if tt.wantTitle != "" {
				data := w.LastRespond.InteractionRespData()
				if assert.Len(t, data.Embeds, 1) {
					assert.Equal(t, tt.wantTitle, data.Embeds[0].Title)
				}
			}
		})
	}
}

func TestProfileHandler_EditFavorite(t *testing.T) {
	tests := []struct {
		name        string
		cmd         CommandContext
		store       *collectiontest.MockStore
		wantContent string
	}{
		{
			name: "success",
			cmd: &MockCommandContext{
				UserIDVal:    1,
				UsernameVal:  "testuser",
				OptInt64Vals: map[string]int64{"id": 42},
			},
			store: &collectiontest.MockStore{
				UpdateFavoriteFunc: func(ctx context.Context, userID collection.UserID, charID int64) error {
					return nil
				},
			},
			wantContent: "Favorite character set as char id 42",
		},
		{
			name: "store error",
			cmd: &MockCommandContext{
				UserIDVal:    1,
				UsernameVal:  "testuser",
				OptInt64Vals: map[string]int64{"id": 42},
			},
			store: &collectiontest.MockStore{
				UpdateFavoriteFunc: func(ctx context.Context, userID collection.UserID, charID int64) error {
					return errors.New("db error")
				},
			},
			wantContent: "An error occurred setting this character",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &cordetest.MockResponseWriter{}
			h := &ProfileHandler{store: tt.store}

			h.EditFavorite(t.Context(), w, tt.cmd)

			assert.True(t, w.RespondCalled)
			data := w.LastRespond.InteractionRespData()
			assert.Contains(t, data.Content, tt.wantContent)
		})
	}
}

func TestProfileHandler_EditQuote(t *testing.T) {
	tests := []struct {
		name        string
		cmd         CommandContext
		store       *collectiontest.MockStore
		wantContent string
	}{
		{
			name: "success",
			cmd: &MockCommandContext{
				UserIDVal:     1,
				UsernameVal:   "testuser",
				OptStringVals: map[string]string{"value": "hello world"},
			},
			store: &collectiontest.MockStore{
				UpdateQuoteFunc: func(ctx context.Context, userID collection.UserID, quote string) error {
					return nil
				},
			},
			wantContent: "Quote set",
		},
		{
			name: "too long",
			cmd: &MockCommandContext{
				UserIDVal:     1,
				UsernameVal:   "testuser",
				OptStringVals: map[string]string{"value": string(make([]byte, 1025))},
			},
			store:       &collectiontest.MockStore{},
			wantContent: "quote is too long",
		},
		{
			name: "store error",
			cmd: &MockCommandContext{
				UserIDVal:     1,
				UsernameVal:   "testuser",
				OptStringVals: map[string]string{"value": "hello"},
			},
			store: &collectiontest.MockStore{
				UpdateQuoteFunc: func(ctx context.Context, userID collection.UserID, quote string) error {
					return errors.New("db error")
				},
			},
			wantContent: "db error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &cordetest.MockResponseWriter{}
			h := &ProfileHandler{store: tt.store}

			h.EditQuote(t.Context(), w, tt.cmd)

			assert.True(t, w.RespondCalled)
			data := w.LastRespond.InteractionRespData()
			assert.Contains(t, data.Content, tt.wantContent)
		})
	}
}

func TestProfileHandler_EditAnilistURL(t *testing.T) {
	tests := []struct {
		name        string
		cmd         CommandContext
		store       *collectiontest.MockStore
		wantContent string
	}{
		{
			name: "success",
			cmd: &MockCommandContext{
				UserIDVal:     1,
				UsernameVal:   "testuser",
				OptStringVals: map[string]string{"url": "https://anilist.co/user/testuser"},
			},
			store: &collectiontest.MockStore{
				UpdateAnilistURLFunc: func(ctx context.Context, userID collection.UserID, url string) error {
					return nil
				},
			},
			wantContent: "Anilist URL set as https://anilist.co/user/testuser",
		},
		{
			name: "invalid url bad host",
			cmd: &MockCommandContext{
				UserIDVal:     1,
				UsernameVal:   "testuser",
				OptStringVals: map[string]string{"url": "https://myanimelist.net/user/testuser"},
			},
			store:       &collectiontest.MockStore{},
			wantContent: "invalid Anilist URL",
		},
		{
			name: "invalid url bad path",
			cmd: &MockCommandContext{
				UserIDVal:     1,
				UsernameVal:   "testuser",
				OptStringVals: map[string]string{"url": "https://anilist.co/settings"},
			},
			store:       &collectiontest.MockStore{},
			wantContent: "invalid Anilist URL",
		},
		{
			name: "store error",
			cmd: &MockCommandContext{
				UserIDVal:     1,
				UsernameVal:   "testuser",
				OptStringVals: map[string]string{"url": "https://anilist.co/user/testuser"},
			},
			store: &collectiontest.MockStore{
				UpdateAnilistURLFunc: func(ctx context.Context, userID collection.UserID, url string) error {
					return errors.New("db error")
				},
			},
			wantContent: "db error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &cordetest.MockResponseWriter{}
			h := &ProfileHandler{store: tt.store}

			h.EditAnilistURL(t.Context(), w, tt.cmd)

			assert.True(t, w.RespondCalled)
			data := w.LastRespond.InteractionRespData()
			assert.Contains(t, data.Content, tt.wantContent)
		})
	}
}

func TestProfileHandler_Autocomplete(t *testing.T) {
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
			h := &ProfileHandler{store: tt.store}

			h.Autocomplete(t.Context(), w, tt.interaction)

			assert.Equal(t, tt.wantCalled, w.AutocompleteCalled)
			if tt.wantCalled {
				choices := w.Choices()
				assert.Len(t, choices, tt.wantChoiceCount)
			}
		})
	}
}
