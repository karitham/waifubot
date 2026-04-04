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
	"github.com/karitham/waifubot/wishlist"
	"github.com/karitham/waifubot/wishlist/wishlisttest"
)

func TestWishlistHandler_CharacterAdd(t *testing.T) {
	tests := []struct {
		name        string
		cmd         CommandContext
		collStore   *collectiontest.MockStore
		wlStore     *wishlisttest.MockStore
		wantContent string
	}{
		{
			name: "happy path adds character",
			cmd: &MockCommandContext{
				UserIDVal:    1,
				UsernameVal:  "testuser",
				GuildIDVal:   1,
				OptInt64Vals: map[string]int64{"character": 42},
			},
			collStore: &collectiontest.MockStore{
				GetOwnedCharacterFunc: func(ctx context.Context, userID collection.UserID, charID int64) (collection.OwnedCharacter, error) {
					return collection.OwnedCharacter{}, collection.ErrNotFound
				},
			},
			wlStore:     &wishlisttest.MockStore{},
			wantContent: "Added  (0) to your wishlist.",
		},
		{
			name: "already owns character",
			cmd: &MockCommandContext{
				UserIDVal:    1,
				UsernameVal:  "testuser",
				GuildIDVal:   1,
				OptInt64Vals: map[string]int64{"character": 42},
			},
			collStore: &collectiontest.MockStore{
				GetOwnedCharacterFunc: func(ctx context.Context, userID collection.UserID, charID int64) (collection.OwnedCharacter, error) {
					return collection.OwnedCharacter{Character: collection.Character{ID: 42, Name: "TestChar"}}, nil
				},
			},
			wantContent: "You already own this character.",
		},
		{
			name: "ownership check error",
			cmd: &MockCommandContext{
				UserIDVal:    1,
				UsernameVal:  "testuser",
				GuildIDVal:   1,
				OptInt64Vals: map[string]int64{"character": 42},
			},
			collStore: &collectiontest.MockStore{
				GetOwnedCharacterFunc: func(ctx context.Context, userID collection.UserID, charID int64) (collection.OwnedCharacter, error) {
					return collection.OwnedCharacter{}, errors.New("db error")
				},
			},
			wantContent: "Unable to verify character ownership.",
		},
		{
			name: "wishlist add error",
			cmd: &MockCommandContext{
				UserIDVal:    1,
				UsernameVal:  "testuser",
				GuildIDVal:   1,
				OptInt64Vals: map[string]int64{"character": 42},
			},
			collStore: &collectiontest.MockStore{
				GetOwnedCharacterFunc: func(ctx context.Context, userID collection.UserID, charID int64) (collection.OwnedCharacter, error) {
					return collection.OwnedCharacter{}, collection.ErrNotFound
				},
			},
			wlStore: &wishlisttest.MockStore{
				AddCharactersToWishlistFunc: func(ctx context.Context, userID uint64, characterIDs []int64) error {
					return errors.New("db error")
				},
			},
			wantContent: "Unable to add character to wishlist.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &cordetest.MockResponseWriter{}
			h := &WishlistHandler{
				wishlist: tt.wlStore,
				store:    tt.collStore,
				catalog:  &cordetest.MockCatalogStore{},
			}

			h.CharacterAdd(t.Context(), w, tt.cmd)

			assert.True(t, w.RespondCalled)
			if tt.wantContent != "" {
				data := w.LastRespond.InteractionRespData()
				assert.Contains(t, data.Content, tt.wantContent)
			}
		})
	}
}

func TestWishlistHandler_CharacterRemove(t *testing.T) {
	tests := []struct {
		name        string
		cmd         CommandContext
		wlStore     *wishlisttest.MockStore
		wantContent string
	}{
		{
			name: "happy path removes character",
			cmd: &MockCommandContext{
				UserIDVal:    1,
				UsernameVal:  "testuser",
				GuildIDVal:   1,
				OptInt64Vals: map[string]int64{"character": 42},
			},
			wlStore:     &wishlisttest.MockStore{},
			wantContent: "Removed character 42 from your wishlist.",
		},
		{
			name: "remove error",
			cmd: &MockCommandContext{
				UserIDVal:    1,
				UsernameVal:  "testuser",
				GuildIDVal:   1,
				OptInt64Vals: map[string]int64{"character": 42},
			},
			wlStore: &wishlisttest.MockStore{
				RemoveCharactersFromWishlistFunc: func(ctx context.Context, userID uint64, characterIDs []int64) error {
					return errors.New("db error")
				},
			},
			wantContent: "Unable to remove character from wishlist.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &cordetest.MockResponseWriter{}
			h := &WishlistHandler{
				wishlist: tt.wlStore,
				store:    &collectiontest.MockStore{},
			}

			h.CharacterRemove(t.Context(), w, tt.cmd)

			assert.True(t, w.RespondCalled)
			if tt.wantContent != "" {
				data := w.LastRespond.InteractionRespData()
				assert.Contains(t, data.Content, tt.wantContent)
			}
		})
	}
}

func TestWishlistHandler_CharacterList(t *testing.T) {
	tests := []struct {
		name        string
		cmd         CommandContext
		wlStore     *wishlisttest.MockStore
		wantContent string
	}{
		{
			name: "happy path shows wishlist",
			cmd: &MockCommandContext{
				UserIDVal:   1,
				UsernameVal: "testuser",
				GuildIDVal:  1,
			},
			wlStore: &wishlisttest.MockStore{
				GetUserCharacterWishlistFunc: func(ctx context.Context, userID uint64) ([]wishlist.Character, error) {
					return []wishlist.Character{
						{ID: 1, Name: "Sakura"},
						{ID: 2, Name: "Naruto"},
					}, nil
				},
			},
			wantContent: "testuser's Wishlist",
		},
		{
			name: "empty wishlist",
			cmd: &MockCommandContext{
				UserIDVal:   1,
				UsernameVal: "testuser",
				GuildIDVal:  1,
			},
			wlStore: &wishlisttest.MockStore{
				GetUserCharacterWishlistFunc: func(ctx context.Context, userID uint64) ([]wishlist.Character, error) {
					return []wishlist.Character{}, nil
				},
			},
			wantContent: "Your wishlist is empty.",
		},
		{
			name: "store error",
			cmd: &MockCommandContext{
				UserIDVal:   1,
				UsernameVal: "testuser",
				GuildIDVal:  1,
			},
			wlStore: &wishlisttest.MockStore{
				GetUserCharacterWishlistFunc: func(ctx context.Context, userID uint64) ([]wishlist.Character, error) {
					return nil, errors.New("db error")
				},
			},
			wantContent: "Unable to retrieve your wishlist.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &cordetest.MockResponseWriter{}
			h := &WishlistHandler{
				wishlist: tt.wlStore,
				store:    &collectiontest.MockStore{},
			}

			h.CharacterList(t.Context(), w, tt.cmd)

			assert.True(t, w.RespondCalled)
			if tt.wantContent != "" {
				w.AssertContains(t, tt.wantContent)
			}
		})
	}
}

func TestWishlistHandler_CharacterRemoveAll(t *testing.T) {
	tests := []struct {
		name        string
		cmd         CommandContext
		wlStore     *wishlisttest.MockStore
		wantContent string
	}{
		{
			name: "happy path clears wishlist",
			cmd: &MockCommandContext{
				UserIDVal:   1,
				UsernameVal: "testuser",
				GuildIDVal:  1,
			},
			wlStore:     &wishlisttest.MockStore{},
			wantContent: "Cleared your wishlist.",
		},
		{
			name: "remove all error",
			cmd: &MockCommandContext{
				UserIDVal:   1,
				UsernameVal: "testuser",
				GuildIDVal:  1,
			},
			wlStore: &wishlisttest.MockStore{
				RemoveAllFromWishlistFunc: func(ctx context.Context, userID uint64) error {
					return errors.New("db error")
				},
			},
			wantContent: "Unable to clear your wishlist.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &cordetest.MockResponseWriter{}
			h := &WishlistHandler{
				wishlist: tt.wlStore,
				store:    &collectiontest.MockStore{},
			}

			h.CharacterRemoveAll(t.Context(), w, tt.cmd)

			assert.True(t, w.RespondCalled)
			if tt.wantContent != "" {
				data := w.LastRespond.InteractionRespData()
				assert.Contains(t, data.Content, tt.wantContent)
			}
		})
	}
}

func TestWishlistHandler_MediaAdd(t *testing.T) {
	tests := []struct {
		name         string
		cmd          CommandContext
		wlStore      *wishlisttest.MockStore
		animeService *collectiontest.MockAnimeService
		collStore    *collectiontest.MockStore
		wantContent  string
	}{
		{
			name: "happy path adds media characters",
			cmd: &MockCommandContext{
				UserIDVal:    1,
				UsernameVal:  "testuser",
				GuildIDVal:   1,
				OptInt64Vals: map[string]int64{"media": 10},
			},
			collStore: &collectiontest.MockStore{
				GetCollectionIDsFunc: func(ctx context.Context, userID collection.UserID) ([]int64, error) {
					return nil, nil
				},
			},
			animeService: &collectiontest.MockAnimeService{
				GetMediaCharactersFunc: func(ctx context.Context, mediaId int64) ([]collection.MediaCharacter, error) {
					return []collection.MediaCharacter{
						{ID: 1, Name: "Char1", ImageURL: "http://img1", MediaTitle: "Anime1"},
						{ID: 2, Name: "Char2", ImageURL: "http://img2", MediaTitle: "Anime1"},
					}, nil
				},
			},
			wlStore:     &wishlisttest.MockStore{},
			wantContent: "Added 2 characters from this media to your wishlist.",
		},
		{
			name: "no characters found",
			cmd: &MockCommandContext{
				UserIDVal:    1,
				UsernameVal:  "testuser",
				GuildIDVal:   1,
				OptInt64Vals: map[string]int64{"media": 10},
			},
			collStore: &collectiontest.MockStore{
				GetCollectionIDsFunc: func(ctx context.Context, userID collection.UserID) ([]int64, error) {
					return nil, nil
				},
			},
			animeService: &collectiontest.MockAnimeService{
				GetMediaCharactersFunc: func(ctx context.Context, mediaId int64) ([]collection.MediaCharacter, error) {
					return nil, nil
				},
			},
			wlStore:     &wishlisttest.MockStore{},
			wantContent: "No characters found for this media",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &cordetest.MockResponseWriter{}
			h := &WishlistHandler{
				wishlist:     tt.wlStore,
				store:        tt.collStore,
				animeService: tt.animeService,
			}

			h.MediaAdd(t.Context(), w, tt.cmd)

			assert.True(t, w.RespondCalled)
			if tt.wantContent != "" {
				w.AssertContains(t, tt.wantContent)
			}
		})
	}
}

func TestWishlistHandler_Holders(t *testing.T) {
	tests := []struct {
		name        string
		cmd         CommandContext
		wlStore     *wishlisttest.MockStore
		wantContent string
	}{
		{
			name: "happy path shows holders",
			cmd: &MockCommandContext{
				UserIDVal:   1,
				UsernameVal: "testuser",
				GuildIDVal:  1,
			},
			wlStore: &wishlisttest.MockStore{
				GetUserCharacterWishlistFunc: func(ctx context.Context, userID uint64) ([]wishlist.Character, error) {
					return []wishlist.Character{{ID: 1, Name: "Sakura"}}, nil
				},
				GetWishlistHoldersFunc: func(ctx context.Context, characterIDs []int64, userID, guildID uint64) ([]wishlist.UserCharacterSet, error) {
					return []wishlist.UserCharacterSet{
						{UserID: 100, Characters: []wishlist.Character{{ID: 1, Name: "Sakura"}}},
					}, nil
				},
			},
			wantContent: "Characters from Your Wishlist",
		},
		{
			name: "empty wishlist",
			cmd: &MockCommandContext{
				UserIDVal:   1,
				UsernameVal: "testuser",
				GuildIDVal:  1,
			},
			wlStore: &wishlisttest.MockStore{
				GetUserCharacterWishlistFunc: func(ctx context.Context, userID uint64) ([]wishlist.Character, error) {
					return []wishlist.Character{}, nil
				},
			},
			wantContent: "Your wishlist is empty.",
		},
		{
			name: "no holders found",
			cmd: &MockCommandContext{
				UserIDVal:   1,
				UsernameVal: "testuser",
				GuildIDVal:  1,
			},
			wlStore: &wishlisttest.MockStore{
				GetUserCharacterWishlistFunc: func(ctx context.Context, userID uint64) ([]wishlist.Character, error) {
					return []wishlist.Character{{ID: 1, Name: "Sakura"}}, nil
				},
				GetWishlistHoldersFunc: func(ctx context.Context, characterIDs []int64, userID, guildID uint64) ([]wishlist.UserCharacterSet, error) {
					return []wishlist.UserCharacterSet{}, nil
				},
			},
			wantContent: "No one has characters from your wishlist.",
		},
		{
			name: "store error",
			cmd: &MockCommandContext{
				UserIDVal:   1,
				UsernameVal: "testuser",
				GuildIDVal:  1,
			},
			wlStore: &wishlisttest.MockStore{
				GetUserCharacterWishlistFunc: func(ctx context.Context, userID uint64) ([]wishlist.Character, error) {
					return nil, errors.New("db error")
				},
			},
			wantContent: "Unable to retrieve your wishlist.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &cordetest.MockResponseWriter{}
			h := &WishlistHandler{
				wishlist: tt.wlStore,
				store:    &collectiontest.MockStore{},
			}

			h.Holders(t.Context(), w, tt.cmd)

			assert.True(t, w.RespondCalled)
			if tt.wantContent != "" {
				w.AssertContains(t, tt.wantContent)
			}
		})
	}
}

func TestWishlistHandler_Wanted(t *testing.T) {
	tests := []struct {
		name        string
		cmd         CommandContext
		wlStore     *wishlisttest.MockStore
		wantContent string
	}{
		{
			name: "happy path shows wanted",
			cmd: &MockCommandContext{
				UserIDVal:   1,
				UsernameVal: "testuser",
				GuildIDVal:  1,
			},
			wlStore: &wishlisttest.MockStore{
				GetWantedCharactersFunc: func(ctx context.Context, userID, guildID uint64) ([]wishlist.UserCharacterSet, error) {
					return []wishlist.UserCharacterSet{
						{UserID: 100, Characters: []wishlist.Character{{ID: 1, Name: "Sakura"}}},
					}, nil
				},
			},
			wantContent: "People Who Want Your Characters",
		},
		{
			name: "no wanted characters",
			cmd: &MockCommandContext{
				UserIDVal:   1,
				UsernameVal: "testuser",
				GuildIDVal:  1,
			},
			wlStore: &wishlisttest.MockStore{
				GetWantedCharactersFunc: func(ctx context.Context, userID, guildID uint64) ([]wishlist.UserCharacterSet, error) {
					return []wishlist.UserCharacterSet{}, nil
				},
			},
			wantContent: "No one wants characters from your collection.",
		},
		{
			name: "store error",
			cmd: &MockCommandContext{
				UserIDVal:   1,
				UsernameVal: "testuser",
				GuildIDVal:  1,
			},
			wlStore: &wishlisttest.MockStore{
				GetWantedCharactersFunc: func(ctx context.Context, userID, guildID uint64) ([]wishlist.UserCharacterSet, error) {
					return nil, errors.New("db error")
				},
			},
			wantContent: "Unable to retrieve wanted characters.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &cordetest.MockResponseWriter{}
			h := &WishlistHandler{
				wishlist: tt.wlStore,
				store:    &collectiontest.MockStore{},
			}

			h.Wanted(t.Context(), w, tt.cmd)

			assert.True(t, w.RespondCalled)
			if tt.wantContent != "" {
				w.AssertContains(t, tt.wantContent)
			}
		})
	}
}

func TestWishlistHandler_Compare(t *testing.T) {
	tests := []struct {
		name        string
		cmd         CommandContext
		wlStore     *wishlisttest.MockStore
		wantContent string
	}{
		{
			name: "happy path compares wishlists",
			cmd: &MockCommandContext{
				UserIDVal:   1,
				UsernameVal: "testuser",
				GuildIDVal:  1,
				FirstResolvedUserVal: corde.User{
					ID:       2,
					Username: "otheruser",
				},
				HasResolvedUser: true,
			},
			wlStore: &wishlisttest.MockStore{
				CompareWithUserFunc: func(ctx context.Context, userID1, userID2 uint64) (wishlist.WishlistComparison, error) {
					return wishlist.WishlistComparison{
						UserHasCharacters: []wishlist.Character{{ID: 1, Name: "Sakura"}},
					}, nil
				},
			},
			wantContent: "Wishlist Comparison with otheruser",
		},
		{
			name: "missing target user",
			cmd: &MockCommandContext{
				UserIDVal:   1,
				UsernameVal: "testuser",
				GuildIDVal:  1,
			},
			wlStore:     &wishlisttest.MockStore{},
			wantContent: "you must specify a user to compare with",
		},
		{
			name: "compare error",
			cmd: &MockCommandContext{
				UserIDVal:   1,
				UsernameVal: "testuser",
				GuildIDVal:  1,
				FirstResolvedUserVal: corde.User{
					ID:       2,
					Username: "otheruser",
				},
				HasResolvedUser: true,
			},
			wlStore: &wishlisttest.MockStore{
				CompareWithUserFunc: func(ctx context.Context, userID1, userID2 uint64) (wishlist.WishlistComparison, error) {
					return wishlist.WishlistComparison{}, errors.New("db error")
				},
			},
			wantContent: "Unable to compare wishlists.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &cordetest.MockResponseWriter{}
			h := &WishlistHandler{
				wishlist: tt.wlStore,
				store:    &collectiontest.MockStore{},
			}

			h.Compare(t.Context(), w, tt.cmd)

			assert.True(t, w.RespondCalled)
			if tt.wantContent != "" {
				w.AssertContains(t, tt.wantContent)
			}
		})
	}
}

func TestWishlistHandler_CharacterAutocomplete(t *testing.T) {
	tests := []struct {
		name             string
		input            corde.JsonRaw
		searchResult     []catalog.Character
		searchErr        error
		wantAutocomplete bool
		wantSearchCalled bool
		wantChoices      int
	}{
		{
			name:             "results found",
			input:            corde.JsonRaw(`"sakura"`),
			searchResult:     []catalog.Character{{ID: 1, Name: "Sakura"}, {ID: 2, Name: "Sakura Kinomoto"}},
			wantAutocomplete: true,
			wantSearchCalled: true,
			wantChoices:      2,
		},
		{
			name:             "no results",
			input:            corde.JsonRaw(`"xyz"`),
			searchResult:     nil,
			wantAutocomplete: true,
			wantSearchCalled: true,
			wantChoices:      0,
		},
		{
			name:             "search error",
			input:            corde.JsonRaw(`"test"`),
			searchErr:        errors.New("search failed"),
			wantAutocomplete: false,
			wantSearchCalled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var searchCalled bool
			mockCatalog := &cordetest.MockCatalogStore{
				SearchGlobalCharactersFunc: func(ctx context.Context, term string) ([]catalog.Character, error) {
					searchCalled = true
					return tt.searchResult, tt.searchErr
				},
			}

			w := &cordetest.MockResponseWriter{}
			i := cordetest.AutocompleteInteraction(1, 1, 1, "test", corde.OptionsInteractions{"character": tt.input})
			h := &WishlistHandler{catalog: mockCatalog}

			h.CharacterAutocomplete(t.Context(), w, i)

			assert.Equal(t, tt.wantSearchCalled, searchCalled)
			assert.Equal(t, tt.wantAutocomplete, w.AutocompleteCalled)

			if tt.wantAutocomplete {
				choices := w.Choices()
				assert.Len(t, choices, tt.wantChoices)
			}
		})
	}
}

func TestWishlistHandler_MediaAutocomplete(t *testing.T) {
	tests := []struct {
		name             string
		input            corde.JsonRaw
		searchResult     []collection.Media
		searchErr        error
		wantAutocomplete bool
		wantSearchCalled bool
		wantChoices      int
	}{
		{
			name:             "results found",
			input:            corde.JsonRaw(`"naruto"`),
			searchResult:     []collection.Media{{ID: 1, Title: "Naruto", Type: "ANIME"}},
			wantAutocomplete: true,
			wantSearchCalled: true,
			wantChoices:      1,
		},
		{
			name:             "no results",
			input:            corde.JsonRaw(`"xyz"`),
			searchResult:     nil,
			wantAutocomplete: true,
			wantSearchCalled: true,
			wantChoices:      0,
		},
		{
			name:             "search error",
			input:            corde.JsonRaw(`"test"`),
			searchErr:        errors.New("search failed"),
			wantAutocomplete: false,
			wantSearchCalled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var searchCalled bool
			mockAnime := &collectiontest.MockAnimeService{
				SearchMediaFunc: func(ctx context.Context, search string) ([]collection.Media, error) {
					searchCalled = true
					return tt.searchResult, tt.searchErr
				},
			}

			w := &cordetest.MockResponseWriter{}
			i := cordetest.AutocompleteInteraction(1, 1, 1, "test", corde.OptionsInteractions{"media": tt.input})
			h := &WishlistHandler{animeService: mockAnime}

			h.MediaAutocomplete(t.Context(), w, i)

			assert.Equal(t, tt.wantSearchCalled, searchCalled)
			assert.Equal(t, tt.wantAutocomplete, w.AutocompleteCalled)

			if tt.wantAutocomplete {
				choices := w.Choices()
				assert.Len(t, choices, tt.wantChoices)
			}
		})
	}
}

func TestWishlistHandler_WishlistAutocomplete(t *testing.T) {
	tests := []struct {
		name             string
		input            corde.JsonRaw
		wishlistChars    []wishlist.Character
		wishlistErr      error
		wantAutocomplete bool
		wantChoices      int
	}{
		{
			name:  "results found with matching input",
			input: corde.JsonRaw(`"sakura"`),
			wishlistChars: []wishlist.Character{
				{ID: 1, Name: "Sakura"},
				{ID: 2, Name: "Sakura Kinomoto"},
			},
			wantAutocomplete: true,
			wantChoices:      2,
		},
		{
			name:  "no matching input returns empty",
			input: corde.JsonRaw(`"xyz"`),
			wishlistChars: []wishlist.Character{
				{ID: 1, Name: "Sakura"},
			},
			wantAutocomplete: true,
			wantChoices:      0,
		},
		{
			name:             "empty input returns all",
			input:            corde.JsonRaw(`""`),
			wishlistChars:    []wishlist.Character{{ID: 1, Name: "Sakura"}, {ID: 2, Name: "Naruto"}},
			wantAutocomplete: true,
			wantChoices:      2,
		},
		{
			name:             "wishlist error returns empty",
			input:            corde.JsonRaw(`"test"`),
			wishlistErr:      errors.New("db error"),
			wantAutocomplete: false,
			wantChoices:      0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wlStore := &wishlisttest.MockStore{
				GetUserCharacterWishlistFunc: func(ctx context.Context, userID uint64) ([]wishlist.Character, error) {
					return tt.wishlistChars, tt.wishlistErr
				},
			}

			w := &cordetest.MockResponseWriter{}
			i := cordetest.AutocompleteInteraction(1, 1, 1, "test", corde.OptionsInteractions{"character": tt.input})
			h := &WishlistHandler{wishlist: wlStore}

			h.WishlistAutocomplete(t.Context(), w, i)

			assert.Equal(t, tt.wantAutocomplete, w.AutocompleteCalled)

			if tt.wantAutocomplete {
				choices := w.Choices()
				assert.Len(t, choices, tt.wantChoices)
			}
		})
	}
}
