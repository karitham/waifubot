package wishlist

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddCharacter_InMemory(t *testing.T) {
	store := NewInMemoryStore()
	ctx := context.Background()
	userID := uint64(123)
	charID := int64(456)

	err := AddCharacter(ctx, store, userID, charID)
	require.NoError(t, err)

	wishlist, err := GetUserWishlist(ctx, store, userID)
	require.NoError(t, err)
	assert.Len(t, wishlist, 1)
	assert.Equal(t, charID, wishlist[0].ID)
}

func TestRemoveCharacter_InMemory(t *testing.T) {
	store := NewInMemoryStore()
	ctx := context.Background()
	userID := uint64(123)
	charID := int64(456)

	AddCharacter(ctx, store, userID, charID)
	err := RemoveCharacter(ctx, store, userID, charID)
	require.NoError(t, err)

	wishlist, err := GetUserWishlist(ctx, store, userID)
	require.NoError(t, err)
	assert.Len(t, wishlist, 0)
}

func TestRemoveAll_InMemory(t *testing.T) {
	store := NewInMemoryStore()
	ctx := context.Background()
	userID := uint64(123)

	AddCharacter(ctx, store, userID, 1)
	AddCharacter(ctx, store, userID, 2)
	AddCharacter(ctx, store, userID, 3)

	err := RemoveAll(ctx, store, userID)
	require.NoError(t, err)

	wishlist, err := GetUserWishlist(ctx, store, userID)
	require.NoError(t, err)
	assert.Len(t, wishlist, 0)
}

func TestCompareWithUser_InMemory(t *testing.T) {
	store := NewInMemoryStore()
	ctx := context.Background()
	user1 := uint64(1)
	user2 := uint64(2)

	AddMultipleCharacters(ctx, store, user1, []int64{1, 2, 3})
	AddMultipleCharacters(ctx, store, user2, []int64{2, 3, 4})

	comp, err := CompareWithUser(ctx, store, user1, user2)
	require.NoError(t, err)

	// user1 has {1, 2, 3}, user2 has {2, 3, 4}, mutual = {2, 3}
	assert.Equal(t, 2, comp.MutualMatches)
}
