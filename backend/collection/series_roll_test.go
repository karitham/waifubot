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
)

func TestSeriesRoll(t *testing.T) {
	config := collection.Config{SeriesRollCost: 20}
	rollConfig := collection.RollConfig{}

	tests := []struct {
		name       string
		setup      func(m *collectiontest.MockStore, anime *collectiontest.MockAnimeService)
		userID     uint64
		mediaID    int64
		wantErr    bool
		errAs      error
		wantCharID int64
	}{
		{
			name: "success",
			setup: func(m *collectiontest.MockStore, anime *collectiontest.MockAnimeService) {
				m.GetUserFunc = func(_ context.Context, userID uint64) (collection.User, error) {
					return collection.User{UserID: userID, Tokens: 50}, nil
				}
				m.GetCollectionIDsFunc = func(_ context.Context, _ uint64) ([]int64, error) { return nil, nil }
				anime.GetMediaCharactersFunc = func(_ context.Context, _ int64) ([]collection.MediaCharacter, error) {
					return []collection.MediaCharacter{
						{ID: 10, Name: "Char10", ImageURL: "img10"},
						{ID: 20, Name: "Char20", ImageURL: "img20"},
					}, nil
				}
				m.UpsertCharacterFunc = func(_ context.Context, _ catalog.Character) error { return nil }
				m.AddToCollectionFunc = func(_ context.Context, _ uint64, _ collection.Character, _ string, _ time.Time) error {
					return nil
				}
				m.RemoveFromWishlistFunc = func(_ context.Context, _ uint64, _ int64) error { return nil }
				m.SpendTokensFunc = func(_ context.Context, _ uint64, _ int32) (collection.User, error) {
					return collection.User{}, nil
				}
			},
			userID:  123,
			mediaID: 1,
		},
		{
			name: "insufficient_tokens",
			setup: func(m *collectiontest.MockStore, _ *collectiontest.MockAnimeService) {
				m.GetUserFunc = func(_ context.Context, userID uint64) (collection.User, error) {
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
			setup: func(m *collectiontest.MockStore, anime *collectiontest.MockAnimeService) {
				m.GetUserFunc = func(_ context.Context, userID uint64) (collection.User, error) {
					return collection.User{UserID: userID, Tokens: 50}, nil
				}
				m.GetCollectionIDsFunc = func(_ context.Context, _ uint64) ([]int64, error) { return nil, nil }
				anime.GetMediaCharactersFunc = func(_ context.Context, _ int64) ([]collection.MediaCharacter, error) {
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
			setup: func(m *collectiontest.MockStore, anime *collectiontest.MockAnimeService) {
				m.GetUserFunc = func(_ context.Context, userID uint64) (collection.User, error) {
					return collection.User{UserID: userID, Tokens: 50}, nil
				}
				m.GetCollectionIDsFunc = func(_ context.Context, _ uint64) ([]int64, error) {
					return []int64{10, 20}, nil
				}
				anime.GetMediaCharactersFunc = func(_ context.Context, _ int64) ([]collection.MediaCharacter, error) {
					return []collection.MediaCharacter{
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
			setup: func(m *collectiontest.MockStore, anime *collectiontest.MockAnimeService) {
				m.GetUserFunc = func(_ context.Context, userID uint64) (collection.User, error) {
					return collection.User{UserID: userID, Tokens: 50}, nil
				}
				m.GetCollectionIDsFunc = func(_ context.Context, _ uint64) ([]int64, error) {
					return []int64{10}, nil
				}
				anime.GetMediaCharactersFunc = func(_ context.Context, _ int64) ([]collection.MediaCharacter, error) {
					return []collection.MediaCharacter{
						{ID: 10, Name: "Char10", ImageURL: "img10"},
						{ID: 20, Name: "Char20", ImageURL: "img20"},
						{ID: 30, Name: "Char30", ImageURL: "img30"},
					}, nil
				}
				m.UpsertCharacterFunc = func(_ context.Context, _ catalog.Character) error { return nil }
				m.AddToCollectionFunc = func(_ context.Context, _ uint64, _ collection.Character, _ string, _ time.Time) error {
					return nil
				}
				m.RemoveFromWishlistFunc = func(_ context.Context, _ uint64, _ int64) error { return nil }
				m.SpendTokensFunc = func(_ context.Context, _ uint64, _ int32) (collection.User, error) {
					return collection.User{}, nil
				}
			},
			userID:  123,
			mediaID: 1,
		},
		{
			name: "new_user_creates_account",
			setup: func(m *collectiontest.MockStore, anime *collectiontest.MockAnimeService) {
				callCount := 0
				m.GetUserFunc = func(_ context.Context, userID uint64) (collection.User, error) {
					callCount++
					if callCount == 1 {
						return collection.User{}, collection.ErrNotFound
					}
					return collection.User{UserID: userID, Tokens: 50}, nil
				}
				m.CreateUserFunc = func(_ context.Context, _ uint64) error { return nil }
				m.GetCollectionIDsFunc = func(_ context.Context, _ uint64) ([]int64, error) { return nil, nil }
				anime.GetMediaCharactersFunc = func(_ context.Context, _ int64) ([]collection.MediaCharacter, error) {
					return []collection.MediaCharacter{
						{ID: 10, Name: "Char10", ImageURL: "img10"},
					}, nil
				}
				m.UpsertCharacterFunc = func(_ context.Context, _ catalog.Character) error { return nil }
				m.AddToCollectionFunc = func(_ context.Context, _ uint64, _ collection.Character, _ string, _ time.Time) error {
					return nil
				}
				m.RemoveFromWishlistFunc = func(_ context.Context, _ uint64, _ int64) error { return nil }
				m.SpendTokensFunc = func(_ context.Context, _ uint64, _ int32) (collection.User, error) {
					return collection.User{}, nil
				}
			},
			userID:  456,
			mediaID: 1,
		},
		{
			name: "remove_from_wishlist_fails_roll",
			setup: func(m *collectiontest.MockStore, anime *collectiontest.MockAnimeService) {
				m.GetUserFunc = func(_ context.Context, userID uint64) (collection.User, error) {
					return collection.User{UserID: userID, Tokens: 50}, nil
				}
				m.GetCollectionIDsFunc = func(_ context.Context, _ uint64) ([]int64, error) { return nil, nil }
				anime.GetMediaCharactersFunc = func(_ context.Context, _ int64) ([]collection.MediaCharacter, error) {
					return []collection.MediaCharacter{
						{ID: 10, Name: "Char10", ImageURL: "img10"},
					}, nil
				}
				m.UpsertCharacterFunc = func(_ context.Context, _ catalog.Character) error { return nil }
				m.AddToCollectionFunc = func(_ context.Context, _ uint64, _ collection.Character, _ string, _ time.Time) error {
					return nil
				}
				m.RemoveFromWishlistFunc = func(_ context.Context, _ uint64, _ int64) error {
					return assert.AnError
				}
				m.SpendTokensFunc = func(_ context.Context, _ uint64, _ int32) (collection.User, error) {
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
			anime := &collectiontest.MockAnimeService{}
			if tt.setup != nil {
				tt.setup(store, anime)
			}

			svc := collection.NewRollService(store, rollConfig)
			got, err := svc.SeriesRoll(t.Context(), tt.userID, tt.mediaID, config.SeriesRollCost, anime)

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
