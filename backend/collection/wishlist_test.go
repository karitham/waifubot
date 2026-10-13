package collection_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/karitham/waifubot/catalog"
	"github.com/karitham/waifubot/collection"
	"github.com/karitham/waifubot/collection/collectiontest"
)

func TestRemoveFromWishlist(t *testing.T) {
	created := false
	store := &collectiontest.MockStore{
		CreateUserFunc: func(_ context.Context, _ collection.UserID) error {
			created = true
			return nil
		},
		RemoveFromWishlistFunc: func(_ context.Context, userID collection.UserID, charID catalog.CharacterID) error {
			assert.Equal(t, collection.UserID(1), userID)
			assert.Equal(t, catalog.CharacterID(42), charID)
			return nil
		},
	}

	require.NoError(t, collection.RemoveFromWishlist(t.Context(), store, 1, 42))
	assert.False(t, created, "removal must not materialize the user")
	assert.Equal(t, 1, store.CommitCalls)
}

func TestRemoveAllFromWishlist(t *testing.T) {
	created := false
	store := &collectiontest.MockStore{
		CreateUserFunc: func(_ context.Context, _ collection.UserID) error {
			created = true
			return nil
		},
		RemoveAllFromWishlistFunc: func(_ context.Context, userID collection.UserID) error {
			assert.Equal(t, collection.UserID(1), userID)
			return nil
		},
	}

	require.NoError(t, collection.RemoveAllFromWishlist(t.Context(), store, 1))
	assert.False(t, created, "removal must not materialize the user")
	assert.Equal(t, 1, store.CommitCalls)
}

func TestAddToWishlist(t *testing.T) {
	tests := []struct {
		name         string
		addErr       error
		wantCommit   int
		wantRollback int
	}{
		{
			name:       "success",
			wantCommit: 1,
		},
		{
			name:         "insert failure rolls back",
			addErr:       errors.New("database on fire"),
			wantRollback: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &collectiontest.MockStore{
				AddCharactersToWishlistFunc: func(_ context.Context, userID collection.UserID, characterIDs []catalog.CharacterID) (int64, error) {
					assert.Equal(t, collection.UserID(1), userID)
					assert.Equal(t, []catalog.CharacterID{42}, characterIDs)
					return 1, tt.addErr
				},
			}

			added, err := collection.AddToWishlist(t.Context(), store, 1, []catalog.CharacterID{42})

			if tt.addErr != nil {
				require.ErrorIs(t, err, tt.addErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, int64(1), added)
			}
			assert.Equal(t, tt.wantCommit, store.CommitCalls)
			assert.Equal(t, tt.wantRollback, store.RollbackCalls)
		})
	}
}
