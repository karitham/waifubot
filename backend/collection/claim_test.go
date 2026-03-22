package collection_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/karitham/waifubot/catalog"
	"github.com/karitham/waifubot/collection"
	"github.com/karitham/waifubot/collection/mocks"
)

func TestClaim_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mocks.NewMockStore(ctrl)

	store.EXPECT().WithTx(gomock.Any()).Return(store, nil)
	store.EXPECT().GetDropForUpdate(gomock.Any(), uint64(456)).Return(collection.Drop{
		ID: 1, Name: "Test Character", Image: "test.jpg", MediaTitle: "Test Anime",
	}, nil)
	store.EXPECT().GetUser(gomock.Any(), uint64(123)).Return(collection.User{UserID: 123}, nil)
	store.EXPECT().UpsertCharacter(gomock.Any(), collection.Character{
		ID: 1, Name: "Test Character", Image: "test.jpg",
	}).Return(nil)
	store.EXPECT().AddToCollection(gomock.Any(), uint64(123), collection.Character{
		ID: 1, Name: "Test Character", Image: "test.jpg", MediaTitle: "Test Anime",
	}, "CLAIM", gomock.Any()).Return(nil)
	store.EXPECT().RemoveFromWishlist(gomock.Any(), uint64(123), int64(1)).Return(nil)
	store.EXPECT().DeleteDrop(gomock.Any(), uint64(456)).Return(nil)
	store.EXPECT().Commit(gomock.Any()).Return(nil)

	char, err := collection.Claim(t.Context(), store, 123, 456, "Test Character")
	require.NoError(t, err)
	assert.Equal(t, int64(1), char.ID)
	assert.Equal(t, "Test Character", char.Name)
	assert.Equal(t, "Test Anime", char.MediaTitle)
}

func TestClaim_NoDropInChannel(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mocks.NewMockStore(ctrl)

	store.EXPECT().WithTx(gomock.Any()).Return(store, nil)
	store.EXPECT().GetDropForUpdate(gomock.Any(), uint64(456)).Return(collection.Drop{}, catalog.ErrNotFound)
	store.EXPECT().Rollback(gomock.Any()).Return(nil)

	_, err := collection.Claim(t.Context(), store, 123, 456, "Test")
	require.ErrorIs(t, err, collection.ErrNoDropInChannel)
}

func TestClaim_WrongName(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mocks.NewMockStore(ctrl)

	store.EXPECT().WithTx(gomock.Any()).Return(store, nil)
	store.EXPECT().GetDropForUpdate(gomock.Any(), uint64(456)).Return(collection.Drop{
		ID: 1, Name: "Test Character",
	}, nil)
	store.EXPECT().Rollback(gomock.Any()).Return(nil)

	_, err := collection.Claim(t.Context(), store, 123, 456, "Wrong Name")
	require.ErrorIs(t, err, collection.ErrWrongCharacterName)
}

func TestClaim_AlreadyOwned(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mocks.NewMockStore(ctrl)

	store.EXPECT().WithTx(gomock.Any()).Return(store, nil)
	store.EXPECT().GetDropForUpdate(gomock.Any(), uint64(456)).Return(collection.Drop{
		ID: 1, Name: "Test Character",
	}, nil)
	store.EXPECT().GetUser(gomock.Any(), uint64(123)).Return(collection.User{UserID: 123}, nil)
	store.EXPECT().UpsertCharacter(gomock.Any(), gomock.Any()).Return(nil)
	store.EXPECT().AddToCollection(gomock.Any(), uint64(123), gomock.Any(), "CLAIM", gomock.Any()).Return(collection.ErrAlreadyOwned)
	store.EXPECT().Rollback(gomock.Any()).Return(nil)

	_, err := collection.Claim(t.Context(), store, 123, 456, "Test Character")
	require.ErrorIs(t, err, collection.ErrAlreadyOwned)
}

func TestClaim_NewUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mocks.NewMockStore(ctrl)

	store.EXPECT().WithTx(gomock.Any()).Return(store, nil)
	store.EXPECT().GetDropForUpdate(gomock.Any(), uint64(456)).Return(collection.Drop{
		ID: 1, Name: "Test Character", Image: "test.jpg",
	}, nil)
	store.EXPECT().GetUser(gomock.Any(), uint64(123)).Return(collection.User{}, collection.ErrNotFound)
	store.EXPECT().CreateUser(gomock.Any(), uint64(123)).Return(nil)
	store.EXPECT().UpsertCharacter(gomock.Any(), gomock.Any()).Return(nil)
	store.EXPECT().AddToCollection(gomock.Any(), uint64(123), gomock.Any(), "CLAIM", gomock.Any()).Return(nil)
	store.EXPECT().RemoveFromWishlist(gomock.Any(), uint64(123), int64(1)).Return(nil)
	store.EXPECT().DeleteDrop(gomock.Any(), uint64(456)).Return(nil)
	store.EXPECT().Commit(gomock.Any()).Return(nil)

	char, err := collection.Claim(t.Context(), store, 123, 456, "Test Character")
	require.NoError(t, err)
	assert.Equal(t, int64(1), char.ID)
}

func TestExchange_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mocks.NewMockStore(ctrl)

	store.EXPECT().WithTx(gomock.Any()).Return(store, nil)
	store.EXPECT().GetCharacterByID(gomock.Any(), int64(1)).Return(collection.Character{ID: 1, Name: "Char1", Image: "img1"}, nil)
	store.EXPECT().GetOwnedCharacter(gomock.Any(), uint64(123), int64(1)).Return(collection.OwnedCharacter{
		Character: collection.Character{ID: 1, Name: "Char1"},
		Date:      time.Now(),
		Source:    "ROLL",
		UserID:    123,
	}, nil)
	store.EXPECT().RemoveFromCollection(gomock.Any(), uint64(123), int64(1)).Return(nil)
	store.EXPECT().AddTokens(gomock.Any(), uint64(123), int32(1)).Return(collection.User{UserID: 123, Tokens: 6}, nil)
	store.EXPECT().Commit(gomock.Any()).Return(nil)

	char, err := collection.Exchange(t.Context(), store, 123, 1)
	require.NoError(t, err)
	assert.Equal(t, int64(1), char.ID)
	assert.Equal(t, "Char1", char.Name)
}

func TestExchange_NotOwned(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mocks.NewMockStore(ctrl)

	store.EXPECT().WithTx(gomock.Any()).Return(store, nil)
	store.EXPECT().GetCharacterByID(gomock.Any(), int64(1)).Return(collection.Character{ID: 1}, nil)
	store.EXPECT().GetOwnedCharacter(gomock.Any(), uint64(123), int64(1)).Return(collection.OwnedCharacter{}, collection.ErrNotFound)
	store.EXPECT().Rollback(gomock.Any()).Return(nil)

	_, err := collection.Exchange(t.Context(), store, 123, 1)
	require.ErrorIs(t, err, collection.ErrUserDoesNotOwnCharacter)
}

func TestGive_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mocks.NewMockStore(ctrl)

	store.EXPECT().WithTx(gomock.Any()).Return(store, nil)
	store.EXPECT().GetOwnedCharacter(gomock.Any(), uint64(123), int64(1)).Return(collection.OwnedCharacter{
		Character: collection.Character{ID: 1, Name: "Char1", Image: "img1"},
	}, nil)
	store.EXPECT().GetOwnedCharacter(gomock.Any(), uint64(456), int64(1)).Return(collection.OwnedCharacter{}, collection.ErrNotFound)
	store.EXPECT().GiveCharacter(gomock.Any(), uint64(123), uint64(456), int64(1)).Return(collection.OwnedCharacter{
		Character: collection.Character{ID: 1, Name: "Char1", Image: "img1"},
		Source:    "TRADE",
		UserID:    456,
	}, nil)
	store.EXPECT().RemoveFromWishlist(gomock.Any(), uint64(456), int64(1)).Return(nil)
	store.EXPECT().Commit(gomock.Any()).Return(nil)

	char, err := collection.Give(t.Context(), store, 123, 456, 1)
	require.NoError(t, err)
	assert.Equal(t, int64(1), char.ID)
	assert.Equal(t, "TRADE", char.Source)
	assert.Equal(t, uint64(456), char.UserID)
}

func TestGive_NotOwned(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mocks.NewMockStore(ctrl)

	store.EXPECT().WithTx(gomock.Any()).Return(store, nil)
	store.EXPECT().GetOwnedCharacter(gomock.Any(), uint64(123), int64(1)).Return(collection.OwnedCharacter{}, collection.ErrNotFound)
	store.EXPECT().Rollback(gomock.Any()).Return(nil)

	_, err := collection.Give(t.Context(), store, 123, 456, 1)
	require.ErrorIs(t, err, collection.ErrUserDoesNotOwnCharacter)
}

func TestGive_TargetAlreadyOwns(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mocks.NewMockStore(ctrl)

	store.EXPECT().WithTx(gomock.Any()).Return(store, nil)
	store.EXPECT().GetOwnedCharacter(gomock.Any(), uint64(123), int64(1)).Return(collection.OwnedCharacter{
		Character: collection.Character{ID: 1, Name: "Char1"},
	}, nil)
	store.EXPECT().GetOwnedCharacter(gomock.Any(), uint64(456), int64(1)).Return(collection.OwnedCharacter{
		Character: collection.Character{ID: 1, Name: "Char1"},
	}, nil)
	store.EXPECT().Rollback(gomock.Any()).Return(nil)

	_, err := collection.Give(t.Context(), store, 123, 456, 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already owns")
}
