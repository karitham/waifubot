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
		setup      func(m *collectiontest.MockStore)
		userID     uint64
		wantErr    bool
		wantErrIs  error // sentinel error to match with errors.Is
		wantErrAs  any   // error type to match with errors.As
		wantCharID int64
	}{
		{
			name: "free_roll_success",
			setup: func(m *collectiontest.MockStore) {
				m.RandomCharNotOwnedFunc = func(_ context.Context, _ uint64, _ float64) (catalog.Character, error) {
					return catalog.Character{ID: 3, Name: "Char3", Image: "img3"}, nil
				}
				m.GetUserFunc = func(_ context.Context, userID uint64) (collection.User, error) {
					return collection.User{UserID: userID, Date: time.Now().Add(-2 * time.Hour), Tokens: 5}, nil
				}
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
			setup: func(m *collectiontest.MockStore) {
				m.GetUserFunc = func(_ context.Context, userID uint64) (collection.User, error) {
					return collection.User{UserID: userID, Date: time.Now().Add(-30 * time.Minute), Tokens: 5}, nil
				}
			},
			userID:    123,
			wantErr:   true,
			wantErrAs: &collection.ErrRollCooldown{},
		},
		{
			name: "new_user",
			setup: func(m *collectiontest.MockStore) {
				m.RandomCharNotOwnedFunc = func(_ context.Context, _ uint64, _ float64) (catalog.Character, error) {
					return catalog.Character{ID: 4, Name: "Char4", Image: "img4"}, nil
				}
				m.GetUserFunc = func(_ context.Context, userID uint64) (collection.User, error) {
					return collection.User{}, collection.ErrNotFound
				}
				m.CreateUserFunc = func(_ context.Context, _ uint64) error { return nil }
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
			name: "no_unowned_characters",
			setup: func(m *collectiontest.MockStore) {
				m.RandomCharNotOwnedFunc = func(_ context.Context, _ uint64, _ float64) (catalog.Character, error) {
					return catalog.Character{}, collection.ErrNotFound
				}
				m.GetUserFunc = func(_ context.Context, userID uint64) (collection.User, error) {
					return collection.User{UserID: userID, Date: time.Now().Add(-2 * time.Hour), Tokens: 5}, nil
				}
			},
			userID:    123,
			wantErr:   true,
			wantErrIs: collection.ErrNoUnownedCharacters,
		},
		{
			name: "remove_from_wishlist_fails_roll",
			setup: func(m *collectiontest.MockStore) {
				m.RandomCharNotOwnedFunc = func(_ context.Context, _ uint64, _ float64) (catalog.Character, error) {
					return catalog.Character{ID: 5, Name: "Char5", Image: "img5"}, nil
				}
				m.GetUserFunc = func(_ context.Context, userID uint64) (collection.User, error) {
					return collection.User{UserID: userID, Date: time.Now().Add(-2 * time.Hour), Tokens: 5}, nil
				}
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
			if tt.setup != nil {
				tt.setup(store)
			}

			svc := collection.NewRollService(store, config)
			got, err := svc.Roll(t.Context(), tt.userID)

			if tt.wantErr {
				require.Error(t, err)
				if tt.wantErrIs != nil {
					assert.ErrorIs(t, err, tt.wantErrIs)
				}
				if tt.wantErrAs != nil {
					assert.ErrorAs(t, err, tt.wantErrAs)
				}
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantCharID, got.ID)
		})
	}
}
