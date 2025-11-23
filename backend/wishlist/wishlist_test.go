package wishlist

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/karitham/waifubot/storage/mocks"
	"github.com/karitham/waifubot/storage/wishliststore"
)

func TestAddCharacter(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := mocks.NewMockWishlistQuerier(ctrl)
	store := New(mockStore)

	ctx := context.Background()
	userID := uint64(123)
	charID := int64(456)

	mockStore.EXPECT().AddCharacterToWishlist(ctx, wishliststore.AddCharacterToWishlistParams{
		UserID:      userID,
		CharacterID: charID,
	}).Return(nil)

	err := AddCharacter(ctx, store, userID, charID)
	require.NoError(t, err)
}

func TestRemoveCharacter(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := mocks.NewMockWishlistQuerier(ctrl)
	store := New(mockStore)

	ctx := context.Background()
	userID := uint64(123)
	charID := int64(456)

	mockStore.EXPECT().RemoveCharacterFromWishlist(ctx, wishliststore.RemoveCharacterFromWishlistParams{
		UserID:      userID,
		CharacterID: charID,
	}).Return(nil)

	err := RemoveCharacter(ctx, store, userID, charID)
	require.NoError(t, err)
}

func TestGetUserWishlist(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := mocks.NewMockWishlistQuerier(ctrl)
	store := New(mockStore)

	ctx := context.Background()
	userID := uint64(123)

	mockStore.EXPECT().GetUserCharacterWishlist(ctx, userID).Return([]wishliststore.GetUserCharacterWishlistRow{
		{
			ID:    456,
			Name:  "Test Character",
			Image: "http://example.com/image.jpg",
			Date:  pgtype.Timestamp{Time: time.Now(), Valid: true},
		},
	}, nil)

	chars, err := GetUserWishlist(ctx, store, userID)
	require.NoError(t, err)
	assert.Len(t, chars, 1)
	assert.Equal(t, int64(456), chars[0].ID)
	assert.Equal(t, "Test Character", chars[0].Name)
}

func TestGetWishlistHolders(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := mocks.NewMockWishlistQuerier(ctrl)
	store := New(mockStore)

	ctx := context.Background()
	userID := uint64(123)

	mockStore.EXPECT().GetUserCharacterWishlist(ctx, userID).Return([]wishliststore.GetUserCharacterWishlistRow{
		{ID: 456},
	}, nil)

	mockStore.EXPECT().GetWishlistHolders(ctx, wishliststore.GetWishlistHoldersParams{
		Column1: []int64{456},
		UserID:  userID,
		GuildID: 123,
	}).Return([]wishliststore.GetWishlistHoldersRow{
		{
			UserID:         789,
			CharacterID:    456,
			CharacterName:  "Test Character",
			CharacterImage: "http://example.com/image.jpg",
		},
	}, nil)

	holders, err := GetWishlistHolders(ctx, store, userID, 123)
	require.NoError(t, err)
	assert.Len(t, holders, 1)
	assert.Equal(t, uint64(789), holders[0].UserID)
	assert.Len(t, holders[0].Characters, 1)
}

func TestCompareWithUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := mocks.NewMockWishlistQuerier(ctrl)
	store := New(mockStore)

	ctx := context.Background()
	userID1 := uint64(123)
	userID2 := uint64(456)

	mockStore.EXPECT().CompareWithUser(ctx, wishliststore.CompareWithUserParams{
		UserID:   userID1,
		UserID_2: userID2,
	}).Return([]wishliststore.CompareWithUserRow{
		{
			Type:  "has",
			ID:    789,
			Name:  "Test Character",
			Image: "http://example.com/image.jpg",
			Date:  pgtype.Timestamp{Time: time.Now(), Valid: true},
		},
	}, nil)

	comparison, err := CompareWithUser(ctx, store, userID1, userID2)
	require.NoError(t, err)
	assert.Len(t, comparison.UserHasCharacters, 1)
	assert.Len(t, comparison.UserWantsCharacters, 0)
	assert.Equal(t, 0, comparison.MutualMatches)
}
