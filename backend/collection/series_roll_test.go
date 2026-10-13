package collection_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/karitham/waifubot/catalog"
	"github.com/karitham/waifubot/collection"
	"github.com/karitham/waifubot/collection/collectiontest"
	"github.com/karitham/waifubot/tracker"
)

func TestSeriesRoll(t *testing.T) {
	config := collection.Config{SeriesRollCost: tokenAmount(t, 20)}

	tests := []struct {
		name       string
		setup      func(m *collectiontest.MockStore, anime *collectiontest.MockTrackerService)
		userID     collection.UserID
		mediaID    tracker.MediaID
		wantErr    bool
		errAs      error
		wantCharID catalog.CharacterID
	}{
		{
			name: "success",
			setup: func(m *collectiontest.MockStore, anime *collectiontest.MockTrackerService) {
				m.GetUserFunc = func(_ context.Context, userID collection.UserID) (collection.User, error) {
					return collection.User{UserID: userID, Tokens: 50}, nil
				}
				m.GetCollectionIDsFunc = func(_ context.Context, _ collection.UserID) ([]catalog.CharacterID, error) { return nil, nil }
				anime.GetMediaCharactersFunc = func(_ context.Context, _ tracker.MediaID) ([]tracker.Character, error) {
					return []tracker.Character{
						{ID: 10, Name: "Char10", ImageURL: "img10"},
						{ID: 20, Name: "Char20", ImageURL: "img20"},
					}, nil
				}
				m.UpsertCharacterFunc = func(_ context.Context, _ catalog.Character) error { return nil }
				m.AddToCollectionFunc = func(_ context.Context, _ collection.UserID, _ catalog.CharacterID, _ collection.Source, _ time.Time) error {
					return nil
				}
				m.RemoveFromWishlistFunc = func(_ context.Context, _ collection.UserID, _ catalog.CharacterID) error { return nil }
				m.SpendTokensFunc = func(_ context.Context, _ collection.UserID, _ collection.TokenAmount) (collection.User, error) {
					return collection.User{}, nil
				}
			},
			userID:  123,
			mediaID: 1,
		},
		{
			name: "insufficient_tokens",
			setup: func(m *collectiontest.MockStore, _ *collectiontest.MockTrackerService) {
				m.GetUserFunc = func(_ context.Context, userID collection.UserID) (collection.User, error) {
					return collection.User{UserID: userID, Tokens: 5}, nil
				}
			},
			userID:  123,
			mediaID: 1,
			wantErr: true,
			errAs:   collection.ErrInsufficientTokens,
		},
		{
			name: "media_not_found",
			setup: func(m *collectiontest.MockStore, anime *collectiontest.MockTrackerService) {
				m.GetUserFunc = func(_ context.Context, userID collection.UserID) (collection.User, error) {
					return collection.User{UserID: userID, Tokens: 50}, nil
				}
				m.GetCollectionIDsFunc = func(_ context.Context, _ collection.UserID) ([]catalog.CharacterID, error) { return nil, nil }
				anime.GetMediaCharactersFunc = func(_ context.Context, _ tracker.MediaID) ([]tracker.Character, error) {
					return nil, nil
				}
			},
			userID:  123,
			mediaID: 999,
			wantErr: true,
			errAs:   collection.ErrMediaNotFound,
		},
		{
			name: "all_owned",
			setup: func(m *collectiontest.MockStore, anime *collectiontest.MockTrackerService) {
				m.GetUserFunc = func(_ context.Context, userID collection.UserID) (collection.User, error) {
					return collection.User{UserID: userID, Tokens: 50}, nil
				}
				m.GetCollectionIDsFunc = func(_ context.Context, _ collection.UserID) ([]catalog.CharacterID, error) {
					return []catalog.CharacterID{10, 20}, nil
				}
				anime.GetMediaCharactersFunc = func(_ context.Context, _ tracker.MediaID) ([]tracker.Character, error) {
					return []tracker.Character{
						{ID: 10, Name: "Char10", ImageURL: "img10"},
						{ID: 20, Name: "Char20", ImageURL: "img20"},
					}, nil
				}
			},
			userID:  123,
			mediaID: 1,
			wantErr: true,
			errAs:   collection.ErrNoUnownedCharacters,
		},
		{
			name: "partial_owned",
			setup: func(m *collectiontest.MockStore, anime *collectiontest.MockTrackerService) {
				m.GetUserFunc = func(_ context.Context, userID collection.UserID) (collection.User, error) {
					return collection.User{UserID: userID, Tokens: 50}, nil
				}
				m.GetCollectionIDsFunc = func(_ context.Context, _ collection.UserID) ([]catalog.CharacterID, error) {
					return []catalog.CharacterID{10}, nil
				}
				anime.GetMediaCharactersFunc = func(_ context.Context, _ tracker.MediaID) ([]tracker.Character, error) {
					return []tracker.Character{
						{ID: 10, Name: "Char10", ImageURL: "img10"},
						{ID: 20, Name: "Char20", ImageURL: "img20"},
						{ID: 30, Name: "Char30", ImageURL: "img30"},
					}, nil
				}
				m.KnownCharacterIDsFunc = func(_ context.Context, ids []catalog.CharacterID) ([]catalog.CharacterID, error) {
					return ids, nil
				}
				m.AddToCollectionFunc = func(_ context.Context, _ collection.UserID, _ catalog.CharacterID, _ collection.Source, _ time.Time) error {
					return nil
				}
				m.RemoveFromWishlistFunc = func(_ context.Context, _ collection.UserID, _ catalog.CharacterID) error { return nil }
				m.SpendTokensFunc = func(_ context.Context, _ collection.UserID, _ collection.TokenAmount) (collection.User, error) {
					return collection.User{}, nil
				}
			},
			userID:  123,
			mediaID: 1,
		},
		{
			name: "none_synced",
			setup: func(m *collectiontest.MockStore, anime *collectiontest.MockTrackerService) {
				m.GetUserFunc = func(_ context.Context, userID collection.UserID) (collection.User, error) {
					return collection.User{UserID: userID, Tokens: 50}, nil
				}
				m.GetCollectionIDsFunc = func(_ context.Context, _ collection.UserID) ([]catalog.CharacterID, error) { return nil, nil }
				anime.GetMediaCharactersFunc = func(_ context.Context, _ tracker.MediaID) ([]tracker.Character, error) {
					return []tracker.Character{
						{ID: 10, Name: "Char10", ImageURL: "img10"},
						{ID: 20, Name: "Char20", ImageURL: "img20"},
					}, nil
				}
				m.KnownCharacterIDsFunc = func(_ context.Context, _ []catalog.CharacterID) ([]catalog.CharacterID, error) {
					return nil, nil
				}
			},
			userID:  123,
			mediaID: 1,
			wantErr: true,
			errAs:   collection.ErrCharactersNotSynced,
		},
		{
			name: "partially_synced",
			setup: func(m *collectiontest.MockStore, anime *collectiontest.MockTrackerService) {
				m.GetUserFunc = func(_ context.Context, userID collection.UserID) (collection.User, error) {
					return collection.User{UserID: userID, Tokens: 50}, nil
				}
				m.GetCollectionIDsFunc = func(_ context.Context, _ collection.UserID) ([]catalog.CharacterID, error) { return nil, nil }
				anime.GetMediaCharactersFunc = func(_ context.Context, _ tracker.MediaID) ([]tracker.Character, error) {
					return []tracker.Character{
						{ID: 10, Name: "Char10", ImageURL: "img10"},
						{ID: 20, Name: "Char20", ImageURL: "img20"},
					}, nil
				}
				m.KnownCharacterIDsFunc = func(_ context.Context, ids []catalog.CharacterID) ([]catalog.CharacterID, error) {
					assert.Contains(t, ids, catalog.CharacterID(10))
					assert.Contains(t, ids, catalog.CharacterID(20))
					return []catalog.CharacterID{20}, nil
				}
				m.AddToCollectionFunc = func(_ context.Context, _ collection.UserID, charID catalog.CharacterID, _ collection.Source, _ time.Time) error {
					assert.Equal(t, catalog.CharacterID(20), charID)
					return nil
				}
				m.RemoveFromWishlistFunc = func(_ context.Context, _ collection.UserID, _ catalog.CharacterID) error { return nil }
				m.SpendTokensFunc = func(_ context.Context, _ collection.UserID, _ collection.TokenAmount) (collection.User, error) {
					return collection.User{}, nil
				}
			},
			userID:  123,
			mediaID: 1,
		},
		{
			name: "new_user_insufficient_tokens",
			setup: func(m *collectiontest.MockStore, _ *collectiontest.MockTrackerService) {
				m.GetUserFunc = func(_ context.Context, _ collection.UserID) (collection.User, error) {
					return collection.User{}, collection.ErrNotFound
				}
				// A failed acquisition must not create the account: if it
				// did, this error would replace ErrInsufficientTokens.
				m.CreateUserFunc = func(_ context.Context, _ collection.UserID) error {
					return assert.AnError
				}
			},
			userID:  456,
			mediaID: 1,
			wantErr: true,
			errAs:   collection.ErrInsufficientTokens,
		},
		{
			name: "remove_from_wishlist_fails_roll",
			setup: func(m *collectiontest.MockStore, anime *collectiontest.MockTrackerService) {
				m.GetUserFunc = func(_ context.Context, userID collection.UserID) (collection.User, error) {
					return collection.User{UserID: userID, Tokens: 50}, nil
				}
				m.GetCollectionIDsFunc = func(_ context.Context, _ collection.UserID) ([]catalog.CharacterID, error) { return nil, nil }
				anime.GetMediaCharactersFunc = func(_ context.Context, _ tracker.MediaID) ([]tracker.Character, error) {
					return []tracker.Character{
						{ID: 10, Name: "Char10", ImageURL: "img10"},
					}, nil
				}
				m.KnownCharacterIDsFunc = func(_ context.Context, ids []catalog.CharacterID) ([]catalog.CharacterID, error) {
					return ids, nil
				}
				m.AddToCollectionFunc = func(_ context.Context, _ collection.UserID, _ catalog.CharacterID, _ collection.Source, _ time.Time) error {
					return nil
				}
				m.RemoveFromWishlistFunc = func(_ context.Context, _ collection.UserID, _ catalog.CharacterID) error {
					return assert.AnError
				}
				m.SpendTokensFunc = func(_ context.Context, _ collection.UserID, _ collection.TokenAmount) (collection.User, error) {
					return collection.User{}, nil
				}
			},
			userID:  123,
			mediaID: 1,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &collectiontest.MockStore{}
			anime := &collectiontest.MockTrackerService{}
			if tt.setup != nil {
				tt.setup(store, anime)
			}

			got, err := collection.SeriesRoll(t.Context(), store, anime, tt.userID, tt.mediaID, config)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errAs != nil {
					assert.ErrorIs(t, err, tt.errAs)
				}
				return
			}

			require.NoError(t, err)
			assert.NotZero(t, got.ID)
		})
	}
}
