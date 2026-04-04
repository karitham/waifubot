package collection_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/karitham/waifubot/catalog"
	"github.com/karitham/waifubot/collection"
	"github.com/karitham/waifubot/collection/collectiontest"
)

func TestRoll(t *testing.T) {
	config := collection.RollConfig{RollCooldown: time.Hour}

	tests := []struct {
		name       string
		setup      func(m *collectiontest.MockStore, anime *collectiontest.MockAnimeService)
		userID     uint64
		wantErr    bool
		errAs      any
		wantCharID int64
	}{
		{
			name: "free_roll_success",
			setup: func(m *collectiontest.MockStore, anime *collectiontest.MockAnimeService) {
				anime.RandomCharFunc = func(_ context.Context, _ ...int64) (collection.MediaCharacter, error) {
					return collection.MediaCharacter{ID: 3, Name: "Char3", ImageURL: "img3"}, nil
				}
				m.GetUserFunc = func(_ context.Context, userID uint64) (collection.User, error) {
					return collection.User{UserID: userID, Date: time.Now().Add(-2 * time.Hour), Tokens: 5}, nil
				}
				m.GetCollectionIDsFunc = func(_ context.Context, _ uint64) ([]int64, error) { return nil, nil }
				m.UpsertCharacterFunc = func(_ context.Context, _ catalog.Character) error { return nil }
				m.AddToCollectionFunc = func(_ context.Context, _ uint64, _ collection.Character, _ string, _ time.Time) error {
					return nil
				}
				m.RemoveFromWishlistFunc = func(_ context.Context, _ uint64, _ int64) error { return nil }
				m.UpdateLastRollFunc = func(_ context.Context, _ uint64, _ time.Time) error { return nil }
			},
			userID:     123,
			wantCharID: 3,
		},
		{
			name: "cooldown",
			setup: func(m *collectiontest.MockStore, _ *collectiontest.MockAnimeService) {
				m.GetUserFunc = func(_ context.Context, userID uint64) (collection.User, error) {
					return collection.User{UserID: userID, Date: time.Now().Add(-30 * time.Minute), Tokens: 5}, nil
				}
			},
			userID:  123,
			wantErr: true,
			errAs:   &collection.ErrRollCooldown{},
		},
		{
			name: "new_user",
			setup: func(m *collectiontest.MockStore, anime *collectiontest.MockAnimeService) {
				anime.RandomCharFunc = func(_ context.Context, _ ...int64) (collection.MediaCharacter, error) {
					return collection.MediaCharacter{ID: 4, Name: "Char4", ImageURL: "img4"}, nil
				}
				callCount := 0
				m.GetUserFunc = func(_ context.Context, userID uint64) (collection.User, error) {
					callCount++
					if callCount == 1 {
						return collection.User{}, collection.ErrNotFound
					}
					return collection.User{UserID: userID, Date: time.Time{}, Tokens: 0}, nil
				}
				m.CreateUserFunc = func(_ context.Context, _ uint64) error { return nil }
				m.GetCollectionIDsFunc = func(_ context.Context, _ uint64) ([]int64, error) { return nil, nil }
				m.UpsertCharacterFunc = func(_ context.Context, _ catalog.Character) error { return nil }
				m.AddToCollectionFunc = func(_ context.Context, _ uint64, _ collection.Character, _ string, _ time.Time) error {
					return nil
				}
				m.RemoveFromWishlistFunc = func(_ context.Context, _ uint64, _ int64) error { return nil }
				m.UpdateLastRollFunc = func(_ context.Context, _ uint64, _ time.Time) error { return nil }
			},
			userID:     456,
			wantCharID: 4,
		},
		{
			name: "anime_service_error",
			setup: func(m *collectiontest.MockStore, anime *collectiontest.MockAnimeService) {
				anime.RandomCharFunc = func(_ context.Context, _ ...int64) (collection.MediaCharacter, error) {
					return collection.MediaCharacter{}, errors.New("api error")
				}
				m.GetUserFunc = func(_ context.Context, userID uint64) (collection.User, error) {
					return collection.User{UserID: userID, Date: time.Now().Add(-2 * time.Hour), Tokens: 5}, nil
				}
				m.GetCollectionIDsFunc = func(_ context.Context, _ uint64) ([]int64, error) { return nil, nil }
			},
			userID:  123,
			wantErr: true,
		},
		{
			name: "remove_from_wishlist_fails_roll",
			setup: func(m *collectiontest.MockStore, anime *collectiontest.MockAnimeService) {
				anime.RandomCharFunc = func(_ context.Context, _ ...int64) (collection.MediaCharacter, error) {
					return collection.MediaCharacter{ID: 5, Name: "Char5", ImageURL: "img5"}, nil
				}
				m.GetUserFunc = func(_ context.Context, userID uint64) (collection.User, error) {
					return collection.User{UserID: userID, Date: time.Now().Add(-2 * time.Hour), Tokens: 5}, nil
				}
				m.GetCollectionIDsFunc = func(_ context.Context, _ uint64) ([]int64, error) { return nil, nil }
				m.UpsertCharacterFunc = func(_ context.Context, _ catalog.Character) error { return nil }
				m.AddToCollectionFunc = func(_ context.Context, _ uint64, _ collection.Character, _ string, _ time.Time) error {
					return nil
				}
				m.RemoveFromWishlistFunc = func(_ context.Context, _ uint64, _ int64) error {
					return errors.New("wishlist error")
				}
				m.UpdateLastRollFunc = func(_ context.Context, _ uint64, _ time.Time) error { return nil }
			},
			userID:  123,
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

			svc := collection.NewRollService(store, anime, config)
			got, err := svc.Roll(t.Context(), tt.userID)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errAs != nil {
					assert.True(t, errors.As(err, tt.errAs))
				}
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantCharID, got.ID)
		})
	}
}
