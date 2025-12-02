package wishlist

import (
	"context"
	"testing"
	"time"

	"github.com/Karitham/corde"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/karitham/waifubot/collection"
	"github.com/karitham/waifubot/storage/collectionstore"
	"github.com/karitham/waifubot/storage/mocks"
	"github.com/karitham/waifubot/storage/wishliststore"
)

// Mock implementations for testing
type mockMediaService struct {
	characters []collection.MediaCharacter
	charErr    error
}

func (m *mockMediaService) SearchMedia(ctx context.Context, search string) ([]collection.Media, error) {
	// Not used in current tests
	return nil, nil
}

func (m *mockMediaService) GetMediaCharacters(ctx context.Context, mediaId int64) ([]collection.MediaCharacter, error) {
	return m.characters, m.charErr
}

type mockCollectionService struct {
	ownership map[int64]bool
	ownedIDs  []int64
	checkErr  error
	listErr   error
	upsertErr error
}

func (m *mockCollectionService) CheckOwnership(ctx context.Context, userID corde.Snowflake, charID int64) (bool, collectionstore.Character, error) {
	if m.checkErr != nil {
		return false, collectionstore.Character{}, m.checkErr
	}
	owned, exists := m.ownership[charID]
	if !exists {
		return false, collectionstore.Character{}, nil
	}
	return owned, collectionstore.Character{ID: charID, Name: "Test Character", Image: "test.jpg"}, nil
}

func (m *mockCollectionService) GetUserCollectionIDs(ctx context.Context, userID corde.Snowflake) ([]int64, error) {
	return m.ownedIDs, m.listErr
}

func (m *mockCollectionService) UpsertCharacter(ctx context.Context, charID int64, name, image string) error {
	return m.upsertErr
}

func TestAddCharacter(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := mocks.NewMockWishlistQuerier(ctrl)
	store := New(mockStore)

	ctx := context.Background()
	userID := uint64(123)
	charID := int64(456)

	mockStore.EXPECT().AddMultipleCharactersToWishlist(ctx, wishliststore.AddMultipleCharactersToWishlistParams{
		UserID:  userID,
		Column2: []int64{charID},
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

	mockStore.EXPECT().RemoveMultipleCharactersFromWishlist(ctx, wishliststore.RemoveMultipleCharactersFromWishlistParams{
		UserID:  userID,
		Column2: []int64{charID},
	}).Return(nil)

	err := RemoveCharacter(ctx, store, userID, charID)
	require.NoError(t, err)
}

func TestRemoveAll(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := mocks.NewMockWishlistQuerier(ctrl)
	store := New(mockStore)

	ctx := context.Background()
	userID := uint64(123)

	mockStore.EXPECT().RemoveAllFromWishlist(ctx, userID).Return(nil)

	err := RemoveAll(ctx, store, userID)
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

func TestGetWantedCharacters(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := mocks.NewMockWishlistQuerier(ctrl)
	store := New(mockStore)

	ctx := context.Background()
	userID := uint64(123)

	mockStore.EXPECT().GetWantedCharacters(ctx, wishliststore.GetWantedCharactersParams{
		UserID:  userID,
		GuildID: 123,
	}).Return([]wishliststore.GetWantedCharactersRow{
		{
			UserID:         789,
			CharacterID:    456,
			CharacterName:  "Test Character",
			CharacterImage: "http://example.com/image.jpg",
		},
	}, nil)

	wanted, err := GetWantedCharacters(ctx, store, userID, 123)
	require.NoError(t, err)
	assert.Len(t, wanted, 1)
	assert.Equal(t, uint64(789), wanted[0].UserID)
	assert.Len(t, wanted[0].Characters, 1)
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
			Name:  "Character A",
			Image: "http://example.com/image.jpg",
			Date:  pgtype.Timestamp{Time: time.Now(), Valid: true},
		},
		{
			Type:  "wants",
			ID:    789, // Same ID - mutual match
			Name:  "Character A",
			Image: "http://example.com/image.jpg",
			Date:  pgtype.Timestamp{Time: time.Now(), Valid: true},
		},
		{
			Type:  "wants",
			ID:    999, // Different ID - no mutual match
			Name:  "Character B",
			Image: "http://example.com/image.jpg",
			Date:  pgtype.Timestamp{Time: time.Now(), Valid: true},
		},
	}, nil)

	comparison, err := CompareWithUser(ctx, store, userID1, userID2)
	require.NoError(t, err)
	assert.Len(t, comparison.UserHasCharacters, 1)
	assert.Len(t, comparison.UserWantsCharacters, 2)
	assert.Equal(t, 1, comparison.MutualMatches) // One character appears in both lists
}

func TestAddMediaToWishlist(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWishlistStore := mocks.NewMockWishlistQuerier(ctrl)
	wishlistStore := New(mockWishlistStore)

	ctx := context.Background()
	userID := corde.Snowflake(123)
	mediaID := int64(456)

	tests := []struct {
		name          string
		characters    []collection.MediaCharacter
		ownedIDs      []int64
		expectedCount int
		expectAddCall bool
		charErr       error
		listErr       error
	}{
		{
			name: "success - add new characters",
			characters: []collection.MediaCharacter{
				{ID: 1, Name: "Char 1"},
				{ID: 2, Name: "Char 2"},
				{ID: 3, Name: "Char 3"},
			},
			ownedIDs:      []int64{1}, // User owns char 1
			expectedCount: 2,
			expectAddCall: true,
		},
		{
			name: "no characters to add - all owned",
			characters: []collection.MediaCharacter{
				{ID: 1, Name: "Char 1"},
			},
			ownedIDs:      []int64{1},
			expectedCount: 0,
			expectAddCall: false,
		},
		{
			name:          "no characters in media",
			characters:    []collection.MediaCharacter{},
			ownedIDs:      []int64{},
			expectedCount: 0,
			expectAddCall: false,
		},
		{
			name: "error getting characters",
			characters: []collection.MediaCharacter{
				{ID: 1, Name: "Char 1"},
			},
			charErr:       assert.AnError,
			expectedCount: 0,
			expectAddCall: false,
		},
		{
			name: "error getting owned IDs",
			characters: []collection.MediaCharacter{
				{ID: 1, Name: "Char 1"},
			},
			listErr:       assert.AnError,
			expectedCount: 0,
			expectAddCall: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mediaService := &mockMediaService{
				characters: tt.characters,
				charErr:    tt.charErr,
			}
			collectionService := &mockCollectionService{
				ownedIDs: tt.ownedIDs,
				listErr:  tt.listErr,
			}

			if tt.expectAddCall {
				// Characters that are not in ownedIDs should be added
				ownedSet := make(map[int64]bool)
				for _, id := range tt.ownedIDs {
					ownedSet[id] = true
				}
				var expectedIDs []int64
				for _, char := range tt.characters {
					if !ownedSet[char.ID] {
						expectedIDs = append(expectedIDs, char.ID)
					}
				}
				mockWishlistStore.EXPECT().AddMultipleCharactersToWishlist(ctx, wishliststore.AddMultipleCharactersToWishlistParams{
					UserID:  uint64(userID),
					Column2: expectedIDs,
				}).Return(nil)
			}

			count, err := AddMediaToWishlist(ctx, wishlistStore, mediaService, collectionService, userID, mediaID)

			if tt.charErr != nil || tt.listErr != nil {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedCount, count)
		})
	}
}
