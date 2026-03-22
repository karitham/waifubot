package collection_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/karitham/waifubot/collection"
	"github.com/karitham/waifubot/collection/mocks"
)

func TestRoll_FreeRollSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mocks.NewMockStore(ctrl)
	anime := &mockAnimeService{
		char: collection.MediaCharacter{ID: 3, Name: "Char3", ImageURL: "img3"},
	}
	config := collection.Config{RollCooldown: time.Hour, TokensNeeded: 10}

	now := time.Now().Add(-2 * time.Hour)

	store.EXPECT().WithTx(gomock.Any()).Return(store, nil)
	store.EXPECT().GetUser(gomock.Any(), uint64(123)).Return(collection.User{
		UserID: 123, Date: now, Tokens: 5,
	}, nil)
	store.EXPECT().GetCollectionIDs(gomock.Any(), uint64(123)).Return([]int64{}, nil)
	store.EXPECT().UpsertCharacter(gomock.Any(), collection.Character{ID: 3, Name: "Char3", Image: "img3"}).Return(nil)
	store.EXPECT().AddToCollection(gomock.Any(), uint64(123), collection.Character{ID: 3, Name: "Char3", Image: "img3"}, "ROLL", gomock.Any()).Return(nil)
	store.EXPECT().RemoveFromWishlist(gomock.Any(), uint64(123), int64(3)).Return(nil)
	store.EXPECT().UpdateLastRoll(gomock.Any(), uint64(123), gomock.Any()).Return(nil)
	store.EXPECT().Commit(gomock.Any()).Return(nil)

	got, err := collection.Roll(t.Context(), store, anime, config, 123)
	require.NoError(t, err)
	assert.Equal(t, int64(3), got.ID)
	assert.Equal(t, "Char3", got.Name)
}

func TestRoll_CooldownAndNoTokens(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mocks.NewMockStore(ctrl)
	anime := &mockAnimeService{}
	config := collection.Config{RollCooldown: time.Hour, TokensNeeded: 10}

	store.EXPECT().WithTx(gomock.Any()).Return(store, nil)
	store.EXPECT().GetUser(gomock.Any(), uint64(123)).Return(collection.User{
		UserID: 123, Date: time.Now().Add(-30 * time.Minute), Tokens: 5,
	}, nil)
	store.EXPECT().Rollback(gomock.Any()).Return(nil)

	_, err := collection.Roll(t.Context(), store, anime, config, 123)
	require.Error(t, err)
	var cd collection.ErrRollCooldown
	assert.True(t, errors.As(err, &cd))
	assert.Equal(t, 5, cd.MissingTokens)
}

func TestRoll_NewUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mocks.NewMockStore(ctrl)
	anime := &mockAnimeService{
		char: collection.MediaCharacter{ID: 4, Name: "Char4", ImageURL: "img4"},
	}
	config := collection.Config{RollCooldown: time.Hour, TokensNeeded: 10}

	store.EXPECT().WithTx(gomock.Any()).Return(store, nil)
	store.EXPECT().GetUser(gomock.Any(), uint64(456)).Return(collection.User{}, collection.ErrNotFound)
	store.EXPECT().CreateUser(gomock.Any(), uint64(456)).Return(nil)
	store.EXPECT().GetUser(gomock.Any(), uint64(456)).Return(collection.User{
		UserID: 456, Date: time.Time{}, Tokens: 0,
	}, nil)
	store.EXPECT().GetCollectionIDs(gomock.Any(), uint64(456)).Return([]int64{}, nil)
	store.EXPECT().UpsertCharacter(gomock.Any(), collection.Character{ID: 4, Name: "Char4", Image: "img4"}).Return(nil)
	store.EXPECT().AddToCollection(gomock.Any(), uint64(456), collection.Character{ID: 4, Name: "Char4", Image: "img4"}, "ROLL", gomock.Any()).Return(nil)
	store.EXPECT().RemoveFromWishlist(gomock.Any(), uint64(456), int64(4)).Return(nil)
	store.EXPECT().UpdateLastRoll(gomock.Any(), uint64(456), gomock.Any()).Return(nil)
	store.EXPECT().Commit(gomock.Any()).Return(nil)

	got, err := collection.Roll(t.Context(), store, anime, config, 456)
	require.NoError(t, err)
	assert.Equal(t, int64(4), got.ID)
}

func TestRoll_TokenRoll(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mocks.NewMockStore(ctrl)
	anime := &mockAnimeService{
		char: collection.MediaCharacter{ID: 3, Name: "Char3"},
	}
	config := collection.Config{RollCooldown: 5 * time.Hour, TokensNeeded: 3}

	store.EXPECT().WithTx(gomock.Any()).Return(store, nil)
	store.EXPECT().GetUser(gomock.Any(), uint64(123)).Return(collection.User{
		UserID: 123, Date: time.Now().Add(-1 * time.Hour), Tokens: 5,
	}, nil)
	store.EXPECT().GetCollectionIDs(gomock.Any(), uint64(123)).Return([]int64{}, nil)
	store.EXPECT().UpsertCharacter(gomock.Any(), gomock.Any()).Return(nil)
	store.EXPECT().AddToCollection(gomock.Any(), uint64(123), gomock.Any(), "ROLL", gomock.Any()).Return(nil)
	store.EXPECT().RemoveFromWishlist(gomock.Any(), uint64(123), int64(3)).Return(nil)
	store.EXPECT().SpendTokens(gomock.Any(), uint64(123), int32(3)).Return(collection.User{UserID: 123, Tokens: 2}, nil)
	store.EXPECT().Commit(gomock.Any()).Return(nil)

	got, err := collection.Roll(t.Context(), store, anime, config, 123)
	require.NoError(t, err)
	assert.Equal(t, int64(3), got.ID)
}

func TestRoll_AnimeServiceError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mocks.NewMockStore(ctrl)
	anime := &mockAnimeService{err: errors.New("api error")}
	config := collection.Config{RollCooldown: time.Hour, TokensNeeded: 10}

	store.EXPECT().WithTx(gomock.Any()).Return(store, nil)
	store.EXPECT().GetUser(gomock.Any(), uint64(123)).Return(collection.User{
		UserID: 123, Date: time.Now().Add(-2 * time.Hour), Tokens: 5,
	}, nil)
	store.EXPECT().GetCollectionIDs(gomock.Any(), uint64(123)).Return([]int64{}, nil)
	store.EXPECT().Rollback(gomock.Any()).Return(nil)

	_, err := collection.Roll(t.Context(), store, anime, config, 123)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "api error")
}

func TestTransferTokens_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mocks.NewMockStore(ctrl)

	store.EXPECT().WithTx(gomock.Any()).Return(store, nil)
	store.EXPECT().SpendTokens(gomock.Any(), uint64(1), int32(50)).Return(collection.User{UserID: 1, Tokens: 50}, nil)
	store.EXPECT().AddTokens(gomock.Any(), uint64(2), int32(50)).Return(collection.User{UserID: 2, Tokens: 50}, nil)
	store.EXPECT().Commit(gomock.Any()).Return(nil)

	err := collection.TransferTokens(t.Context(), store, 1, 2, 50)
	require.NoError(t, err)
}

func TestTransferTokens_InsufficientFunds(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mocks.NewMockStore(ctrl)

	store.EXPECT().WithTx(gomock.Any()).Return(store, nil)
	store.EXPECT().SpendTokens(gomock.Any(), uint64(1), int32(100)).Return(collection.User{UserID: 1, Tokens: -1}, nil)
	store.EXPECT().Rollback(gomock.Any()).Return(nil)

	err := collection.TransferTokens(t.Context(), store, 1, 2, 100)
	require.ErrorIs(t, err, collection.ErrInsufficientTokens)
}

func TestTransferTokens_InvalidAmount(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mocks.NewMockStore(ctrl)

	err := collection.TransferTokens(t.Context(), store, 1, 2, 0)
	require.ErrorIs(t, err, collection.ErrInvalidAmount)

	err = collection.TransferTokens(t.Context(), store, 1, 2, -5)
	require.ErrorIs(t, err, collection.ErrInvalidAmount)
}

func TestTransferTokens_SameUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mocks.NewMockStore(ctrl)

	err := collection.TransferTokens(t.Context(), store, 1, 1, 50)
	require.ErrorIs(t, err, collection.ErrSameUserTransfer)
}

// mockAnimeService is a simple test double.
type mockAnimeService struct {
	char collection.MediaCharacter
	err  error
}

func (m *mockAnimeService) RandomChar(ctx context.Context, notIn ...int64) (collection.MediaCharacter, error) {
	return m.char, m.err
}
func (m *mockAnimeService) Anime(ctx context.Context, name string) ([]collection.Media, error) {
	return nil, nil
}
func (m *mockAnimeService) Manga(ctx context.Context, name string) ([]collection.Media, error) {
	return nil, nil
}
func (m *mockAnimeService) User(ctx context.Context, name string) ([]collection.TrackerUser, error) {
	return nil, nil
}
func (m *mockAnimeService) Character(ctx context.Context, name string) ([]collection.MediaCharacter, error) {
	return nil, nil
}
