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
