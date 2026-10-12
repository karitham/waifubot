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
		userID     collection.UserID
		wantErr    bool
		wantErrIs  error // sentinel error to match with errors.Is
		wantErrAs  any   // error type to match with errors.As
		wantCharID catalog.CharacterID
	}{
		{
			name: "free_roll_success",
			setup: func(m *collectiontest.MockStore) {
				m.RandomCharNotOwnedFunc = func(_ context.Context, _ collection.UserID, _ float64) (catalog.Character, error) {
					return catalog.Character{ID: 3, Name: "Char3", Image: "img3"}, nil
				}
				m.GetUserFunc = func(_ context.Context, userID collection.UserID) (collection.User, error) {
					return collection.User{UserID: userID, LastRoll: time.Now().Add(-2 * time.Hour), Tokens: 5}, nil
				}
				m.AddToCollectionFunc = func(_ context.Context, _ collection.UserID, _ catalog.CharacterID, _ collection.Source, _ time.Time) error {
					return nil
				}
				m.RemoveFromWishlistFunc = func(_ context.Context, _ collection.UserID, _ catalog.CharacterID) error { return nil }
				m.RecordRollFunc = func(_ context.Context, _ collection.UserID, _ time.Time, _ time.Duration) error { return nil }
			},
			userID:     123,
			wantCharID: 3,
		},
		{
			name: "cooldown",
			setup: func(m *collectiontest.MockStore) {
				m.GetUserFunc = func(_ context.Context, userID collection.UserID) (collection.User, error) {
					return collection.User{UserID: userID, LastRoll: time.Now().Add(-30 * time.Minute), Tokens: 5}, nil
				}
			},
			userID:    123,
			wantErr:   true,
			wantErrAs: &collection.ErrRollCooldown{},
		},
		{
			name: "cooldown_lost_to_concurrent_roll",
			setup: func(m *collectiontest.MockStore) {
				m.RandomCharNotOwnedFunc = func(_ context.Context, _ collection.UserID, _ float64) (catalog.Character, error) {
					return catalog.Character{ID: 6, Name: "Char6", Image: "img6"}, nil
				}
				m.GetUserFunc = func(_ context.Context, userID collection.UserID) (collection.User, error) {
					return collection.User{UserID: userID, LastRoll: time.Now().Add(-2 * time.Hour)}, nil
				}
				// The checks pass on the stale read; the atomic charge is
				// what rejects the loser of the race.
				m.RecordRollFunc = func(_ context.Context, _ collection.UserID, _ time.Time, _ time.Duration) error {
					return collection.ErrRollCooldown{Until: time.Now().Add(time.Hour)}
				}
			},
			userID:    123,
			wantErr:   true,
			wantErrAs: &collection.ErrRollCooldown{},
		},
		{
			name: "user_created_concurrently",
			setup: func(m *collectiontest.MockStore) {
				m.RandomCharNotOwnedFunc = func(_ context.Context, _ collection.UserID, _ float64) (catalog.Character, error) {
					return catalog.Character{ID: 7, Name: "Char7", Image: "img7"}, nil
				}
				created := false
				m.GetUserFunc = func(_ context.Context, userID collection.UserID) (collection.User, error) {
					if !created {
						return collection.User{}, collection.ErrNotFound
					}
					return collection.User{UserID: userID}, nil
				}
				// The create loses a race; the follow-up read finds the
				// winner's row.
				m.CreateUserFunc = func(_ context.Context, _ collection.UserID) error {
					created = true
					return errors.New("duplicate key")
				}
			},
			userID:     789,
			wantCharID: 7,
		},
		{
			name: "new_user",
			setup: func(m *collectiontest.MockStore) {
				m.RandomCharNotOwnedFunc = func(_ context.Context, _ collection.UserID, _ float64) (catalog.Character, error) {
					return catalog.Character{ID: 4, Name: "Char4", Image: "img4"}, nil
				}
				created := false
				m.GetUserFunc = func(_ context.Context, userID collection.UserID) (collection.User, error) {
					if !created {
						return collection.User{}, collection.ErrNotFound
					}
					return collection.User{UserID: userID}, nil
				}
				m.CreateUserFunc = func(_ context.Context, _ collection.UserID) error {
					created = true
					return nil
				}
				m.AddToCollectionFunc = func(_ context.Context, _ collection.UserID, _ catalog.CharacterID, _ collection.Source, _ time.Time) error {
					return nil
				}
				m.RemoveFromWishlistFunc = func(_ context.Context, _ collection.UserID, _ catalog.CharacterID) error { return nil }
				m.RecordRollFunc = func(_ context.Context, _ collection.UserID, _ time.Time, _ time.Duration) error { return nil }
			},
			userID:     456,
			wantCharID: 4,
		},
		{
			name: "no_unowned_characters",
			setup: func(m *collectiontest.MockStore) {
				m.RandomCharNotOwnedFunc = func(_ context.Context, _ collection.UserID, _ float64) (catalog.Character, error) {
					return catalog.Character{}, collection.ErrNotFound
				}
				m.GetUserFunc = func(_ context.Context, userID collection.UserID) (collection.User, error) {
					return collection.User{UserID: userID, LastRoll: time.Now().Add(-2 * time.Hour), Tokens: 5}, nil
				}
			},
			userID:    123,
			wantErr:   true,
			wantErrIs: collection.ErrNoUnownedCharacters,
		},
		{
			name: "remove_from_wishlist_fails_roll",
			setup: func(m *collectiontest.MockStore) {
				m.RandomCharNotOwnedFunc = func(_ context.Context, _ collection.UserID, _ float64) (catalog.Character, error) {
					return catalog.Character{ID: 5, Name: "Char5", Image: "img5"}, nil
				}
				m.GetUserFunc = func(_ context.Context, userID collection.UserID) (collection.User, error) {
					return collection.User{UserID: userID, LastRoll: time.Now().Add(-2 * time.Hour), Tokens: 5}, nil
				}
				m.AddToCollectionFunc = func(_ context.Context, _ collection.UserID, _ catalog.CharacterID, _ collection.Source, _ time.Time) error {
					return nil
				}
				m.RemoveFromWishlistFunc = func(_ context.Context, _ collection.UserID, _ catalog.CharacterID) error {
					return errors.New("wishlist error")
				}
				m.RecordRollFunc = func(_ context.Context, _ collection.UserID, _ time.Time, _ time.Duration) error { return nil }
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
