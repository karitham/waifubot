package wishlist_test

import (
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/karitham/waifubot/storage/mocks"
	"github.com/karitham/waifubot/storage/wishliststore"
	"github.com/karitham/waifubot/wishlist"
)

func TestGetUserCharacterWishlist(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := mocks.NewMockWishlistQuerier(ctrl)
	store := wishlist.New(mockStore)

	ctx := t.Context()
	userID := uint64(123)

	mockStore.EXPECT().GetUserCharacterWishlist(ctx, userID).Return([]wishliststore.GetUserCharacterWishlistRow{
		{
			ID:    456,
			Name:  "Test Character",
			Image: "http://example.com/image.jpg",
			Date:  pgtype.Timestamp{Time: time.Now().UTC(), Valid: true},
		},
	}, nil)

	chars, err := store.GetUserCharacterWishlist(ctx, userID)
	require.NoError(t, err)
	assert.Len(t, chars, 1)
	assert.Equal(t, int64(456), chars[0].ID)
	assert.Equal(t, "Test Character", chars[0].Name)
}

func TestAddCharactersToWishlist(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := mocks.NewMockWishlistQuerier(ctrl)
	store := wishlist.New(mockStore)

	ctx := t.Context()
	userID := uint64(123)

	mockStore.EXPECT().AddCharactersToWishlist(ctx, wishliststore.AddCharactersToWishlistParams{
		UserID:  userID,
		Column2: []int64{456, 789},
	}).Return(nil)

	err := store.AddCharactersToWishlist(ctx, userID, []int64{456, 789})
	require.NoError(t, err)
}

func TestRemoveAllFromWishlist(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := mocks.NewMockWishlistQuerier(ctrl)
	store := wishlist.New(mockStore)

	ctx := t.Context()
	userID := uint64(123)

	mockStore.EXPECT().RemoveAllFromWishlist(ctx, userID).Return(nil)

	err := store.RemoveAllFromWishlist(ctx, userID)
	require.NoError(t, err)
}
