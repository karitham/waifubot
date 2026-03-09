package inmemory

import (
	"context"
	"testing"

	"github.com/karitham/waifubot/storage/interfaces"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWishlistStore_AddAndGetCharacter(t *testing.T) {
	store := NewWishlistStore()
	ctx := context.Background()
	userID := uint64(123)
	charID := int64(456)

	err := store.AddCharacter(ctx, userID, charID)
	require.NoError(t, err)

	wishlist, err := store.GetUserWishlist(ctx, userID)
	require.NoError(t, err)
	assert.Contains(t, wishlist, charID)
}

func TestWishlistStore_RemoveCharacter(t *testing.T) {
	store := NewWishlistStore()
	ctx := context.Background()
	userID := uint64(123)
	charID := int64(456)

	store.AddCharacter(ctx, userID, charID)
	err := store.RemoveCharacter(ctx, userID, charID)
	require.NoError(t, err)

	wishlist, err := store.GetUserWishlist(ctx, userID)
	require.NoError(t, err)
	assert.NotContains(t, wishlist, charID)
}

func TestWishlistStore_CompareWithUser(t *testing.T) {
	store := NewWishlistStore()
	ctx := context.Background()
	userID := uint64(1)
	otherUserID := uint64(2)

	store.AddMultipleCharacters(ctx, userID, []int64{1, 2, 3})
	store.AddMultipleCharacters(ctx, otherUserID, []int64{2, 3, 4})

	comparison, err := store.CompareWithUser(ctx, userID, otherUserID)
	require.NoError(t, err)

	assert.ElementsMatch(t, []int64{2, 3, 4}, comparison.UserHas)
	assert.ElementsMatch(t, []int64{1, 2, 3}, comparison.UserWants)
	assert.ElementsMatch(t, []int64{2, 3}, comparison.MutualMatches)
}

func TestUserStore_CreateAndGet(t *testing.T) {
	store := NewUserStore()
	ctx := context.Background()
	userID := uint64(123)

	err := store.Create(ctx, userID)
	require.NoError(t, err)

	user, err := store.Get(ctx, userID)
	require.NoError(t, err)
	assert.Equal(t, userID, user.UserID)
	assert.Equal(t, 0, user.Tokens)
}

func TestUserStore_UpdateTokens(t *testing.T) {
	store := NewUserStore()
	ctx := context.Background()
	userID := uint64(123)

	store.Create(ctx, userID)

	user, err := store.UpdateTokens(ctx, userID, 100)
	require.NoError(t, err)
	assert.Equal(t, 100, user.Tokens)
}

func TestCharacterStore_InsertAndList(t *testing.T) {
	store := NewCharacterStore()
	ctx := context.Background()
	userID := uint64(123)
	charID := int64(456)

	store.UpsertCharacter(ctx, interfaces.Character{
		ID:    charID,
		Name:  "Test Character",
		Image: "test.jpg",
	})

	_, err := store.Insert(ctx, userID, charID, "anilist")
	require.NoError(t, err)

	cols, err := store.List(ctx, userID)
	require.NoError(t, err)
	assert.Len(t, cols, 1)
	assert.Equal(t, charID, cols[0].CharacterID)
}

func TestStore_Facade(t *testing.T) {
	store := NewStore()
	ctx := context.Background()

	err := store.UserStore().Create(ctx, 123)
	require.NoError(t, err)

	store.WishlistStore().AddCharacter(ctx, 123, 456)

	wishlist, err := store.WishlistStore().GetUserWishlist(ctx, 123)
	require.NoError(t, err)
	assert.Contains(t, wishlist, int64(456))
}
