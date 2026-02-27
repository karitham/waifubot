package collection

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Karitham/corde"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"go.uber.org/mock/gomock"

	"github.com/karitham/waifubot/storage/collectionstore"
	"github.com/karitham/waifubot/storage/mocks"
	"github.com/karitham/waifubot/storage/userstore"
	"github.com/karitham/waifubot/storage/wishliststore"
)

func TestRoll(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name        string
		userID      corde.Snowflake
		config      Config
		setupMocks  func(*MockProfileStore, *mocks.MockStorageStore, *MockAnimeService, *MockCollectionQuerier, *MockUserQuerier)
		wantErr     bool
		wantChar    MediaCharacter
		errContains string
	}{
		{
			name:   "free roll success",
			userID: 123,
			config: Config{RollCooldown: time.Hour, TokensNeeded: 10},
			setupMocks: func(store *MockProfileStore, tx *mocks.MockStorageStore, anime *MockAnimeService, coll *MockCollectionQuerier, user *MockUserQuerier) {
				wishlist := mocks.NewMockWishlistQuerier(ctrl)
				store.EXPECT().Tx(gomock.Any()).Return(tx, nil)
				tx.EXPECT().UserStore().Return(user).AnyTimes()
				tx.EXPECT().CollectionStore().Return(coll).AnyTimes()
				tx.EXPECT().WishlistStore().Return(wishlist).AnyTimes()
				user.EXPECT().Get(gomock.Any(), uint64(123)).Return(userstore.User{
					UserID: 123,
					Date:   pgtype.Timestamp{Time: time.Now().Add(-2 * time.Hour), Valid: true},
					Tokens: 5,
				}, nil)
				coll.EXPECT().ListIDs(gomock.Any(), uint64(123)).Return([]int64{}, nil)
				anime.EXPECT().RandomChar(gomock.Any()).Return(MediaCharacter{ID: 3, Name: "Char3", ImageURL: "img3"}, nil)
				coll.EXPECT().UpsertCharacter(gomock.Any(), gomock.Any()).Return(collectionstore.Character{}, nil)
				coll.EXPECT().Insert(gomock.Any(), gomock.Any()).Return(collectionstore.Collection{}, nil)
				wishlist.EXPECT().RemoveCharacterFromWishlist(gomock.Any(), wishliststore.RemoveCharacterFromWishlistParams{UserID: 123, CharacterID: 3}).Return(nil)
				user.EXPECT().UpdateDate(gomock.Any(), gomock.Any()).Return(nil)
				tx.EXPECT().Commit(gomock.Any()).Return(nil)
			},
			wantErr:  false,
			wantChar: MediaCharacter{ID: 3, Name: "Char3", ImageURL: "img3"},
		},
		{
			name:   "cooldown and insufficient tokens",
			userID: 123,
			config: Config{RollCooldown: time.Hour, TokensNeeded: 10},
			setupMocks: func(store *MockProfileStore, tx *mocks.MockStorageStore, anime *MockAnimeService, coll *MockCollectionQuerier, user *MockUserQuerier) {
				store.EXPECT().Tx(gomock.Any()).Return(tx, nil)
				tx.EXPECT().UserStore().Return(user)
				user.EXPECT().Get(gomock.Any(), uint64(123)).Return(userstore.User{
					UserID: 123,
					Date:   pgtype.Timestamp{Time: time.Now().Add(-30 * time.Minute), Valid: true},
					Tokens: 5,
				}, nil)
				tx.EXPECT().Rollback(gomock.Any()).Return(nil)
			},
			wantErr:     true,
			errContains: "You need another 5 tokens",
		},
		{
			name:   "get user error",
			userID: 123,
			config: Config{RollCooldown: time.Hour, TokensNeeded: 10},
			setupMocks: func(store *MockProfileStore, tx *mocks.MockStorageStore, anime *MockAnimeService, coll *MockCollectionQuerier, user *MockUserQuerier) {
				store.EXPECT().Tx(gomock.Any()).Return(tx, nil)
				tx.EXPECT().UserStore().Return(user)
				user.EXPECT().Get(gomock.Any(), uint64(123)).Return(userstore.User{}, errors.New("user not found"))
				tx.EXPECT().Rollback(gomock.Any()).Return(nil)
			},
			wantErr:     true,
			errContains: "user not found",
		},
		{
			name:   "new user roll success",
			userID: 456,
			config: Config{RollCooldown: time.Hour, TokensNeeded: 10},
			setupMocks: func(store *MockProfileStore, tx *mocks.MockStorageStore, anime *MockAnimeService, coll *MockCollectionQuerier, user *MockUserQuerier) {
				wishlist := mocks.NewMockWishlistQuerier(ctrl)
				store.EXPECT().Tx(gomock.Any()).Return(tx, nil)
				tx.EXPECT().UserStore().Return(user).AnyTimes()
				tx.EXPECT().CollectionStore().Return(coll).AnyTimes()
				tx.EXPECT().WishlistStore().Return(wishlist).AnyTimes()
				user.EXPECT().Get(gomock.Any(), uint64(456)).Return(userstore.User{}, pgx.ErrNoRows)
				user.EXPECT().Create(gomock.Any(), uint64(456)).Return(nil)
				user.EXPECT().Get(gomock.Any(), uint64(456)).Return(userstore.User{
					UserID: 456,
					Date:   pgtype.Timestamp{Valid: false}, // new user, no date
					Tokens: 0,
				}, nil)
				coll.EXPECT().ListIDs(gomock.Any(), uint64(456)).Return([]int64{}, nil)
				anime.EXPECT().RandomChar(gomock.Any()).Return(MediaCharacter{ID: 4, Name: "Char4", ImageURL: "img4"}, nil)
				coll.EXPECT().UpsertCharacter(gomock.Any(), gomock.Any()).Return(collectionstore.Character{}, nil)
				coll.EXPECT().Insert(gomock.Any(), gomock.Any()).Return(collectionstore.Collection{}, nil)
				wishlist.EXPECT().RemoveCharacterFromWishlist(gomock.Any(), wishliststore.RemoveCharacterFromWishlistParams{UserID: 456, CharacterID: 4}).Return(nil)
				user.EXPECT().UpdateDate(gomock.Any(), gomock.Any()).Return(nil)
				tx.EXPECT().Commit(gomock.Any()).Return(nil)
			},
			wantErr:  false,
			wantChar: MediaCharacter{ID: 4, Name: "Char4", ImageURL: "img4"},
		},
		{
			name:   "list ids error",
			userID: 123,
			config: Config{RollCooldown: time.Hour, TokensNeeded: 10},
			setupMocks: func(store *MockProfileStore, tx *mocks.MockStorageStore, anime *MockAnimeService, coll *MockCollectionQuerier, user *MockUserQuerier) {
				store.EXPECT().Tx(gomock.Any()).Return(tx, nil)
				tx.EXPECT().UserStore().Return(user).AnyTimes()
				tx.EXPECT().CollectionStore().Return(coll)
				user.EXPECT().Get(gomock.Any(), uint64(123)).Return(userstore.User{
					UserID: 123,
					Date:   pgtype.Timestamp{Time: time.Now().Add(-2 * time.Hour), Valid: true},
					Tokens: 5,
				}, nil)
				coll.EXPECT().ListIDs(gomock.Any(), uint64(123)).Return(nil, errors.New("list ids error"))
				tx.EXPECT().Rollback(gomock.Any()).Return(nil)
			},
			wantErr:     true,
			errContains: "list ids error",
		},
		{
			name:   "random char error",
			userID: 123,
			config: Config{RollCooldown: time.Hour, TokensNeeded: 10},
			setupMocks: func(store *MockProfileStore, tx *mocks.MockStorageStore, anime *MockAnimeService, coll *MockCollectionQuerier, user *MockUserQuerier) {
				store.EXPECT().Tx(gomock.Any()).Return(tx, nil)
				tx.EXPECT().UserStore().Return(user).AnyTimes()
				tx.EXPECT().CollectionStore().Return(coll).AnyTimes()
				user.EXPECT().Get(gomock.Any(), uint64(123)).Return(userstore.User{
					UserID: 123,
					Date:   pgtype.Timestamp{Time: time.Now().Add(-2 * time.Hour), Valid: true},
					Tokens: 5,
				}, nil)
				coll.EXPECT().ListIDs(gomock.Any(), uint64(123)).Return([]int64{}, nil)
				anime.EXPECT().RandomChar(gomock.Any()).Return(MediaCharacter{}, errors.New("random char error"))
				tx.EXPECT().Rollback(gomock.Any()).Return(nil)
			},
			wantErr:     true,
			errContains: "random char error",
		},
		{
			name:   "upsert character error",
			userID: 123,
			config: Config{RollCooldown: time.Hour, TokensNeeded: 10},
			setupMocks: func(store *MockProfileStore, tx *mocks.MockStorageStore, anime *MockAnimeService, coll *MockCollectionQuerier, user *MockUserQuerier) {
				store.EXPECT().Tx(gomock.Any()).Return(tx, nil)
				tx.EXPECT().UserStore().Return(user).AnyTimes()
				tx.EXPECT().CollectionStore().Return(coll).AnyTimes()
				user.EXPECT().Get(gomock.Any(), uint64(123)).Return(userstore.User{
					UserID: 123,
					Date:   pgtype.Timestamp{Time: time.Now().Add(-2 * time.Hour), Valid: true},
					Tokens: 5,
				}, nil)
				coll.EXPECT().ListIDs(gomock.Any(), uint64(123)).Return([]int64{}, nil)
				anime.EXPECT().RandomChar(gomock.Any()).Return(MediaCharacter{ID: 3, Name: "Char3", ImageURL: "img3"}, nil)
				coll.EXPECT().UpsertCharacter(gomock.Any(), gomock.Any()).Return(collectionstore.Character{}, errors.New("upsert error"))
				tx.EXPECT().Rollback(gomock.Any()).Return(nil)
			},
			wantErr:     true,
			errContains: "upsert error",
		},
		{
			name:   "insert character error",
			userID: 123,
			config: Config{RollCooldown: time.Hour, TokensNeeded: 10},
			setupMocks: func(store *MockProfileStore, tx *mocks.MockStorageStore, anime *MockAnimeService, coll *MockCollectionQuerier, user *MockUserQuerier) {
				store.EXPECT().Tx(gomock.Any()).Return(tx, nil)
				tx.EXPECT().UserStore().Return(user).AnyTimes()
				tx.EXPECT().CollectionStore().Return(coll).AnyTimes()
				user.EXPECT().Get(gomock.Any(), uint64(123)).Return(userstore.User{
					UserID: 123,
					Date:   pgtype.Timestamp{Time: time.Now().Add(-2 * time.Hour), Valid: true},
					Tokens: 5,
				}, nil)
				coll.EXPECT().ListIDs(gomock.Any(), uint64(123)).Return([]int64{}, nil)
				anime.EXPECT().RandomChar(gomock.Any()).Return(MediaCharacter{ID: 3, Name: "Char3", ImageURL: "img3"}, nil)
				coll.EXPECT().UpsertCharacter(gomock.Any(), gomock.Any()).Return(collectionstore.Character{}, nil)
				coll.EXPECT().Insert(gomock.Any(), gomock.Any()).Return(collectionstore.Collection{}, errors.New("insert error"))
				tx.EXPECT().Rollback(gomock.Any()).Return(nil)
			},
			wantErr:     true,
			errContains: "insert error",
		},
		{
			name:   "update date error",
			userID: 123,
			config: Config{RollCooldown: time.Hour, TokensNeeded: 10},
			setupMocks: func(store *MockProfileStore, tx *mocks.MockStorageStore, anime *MockAnimeService, coll *MockCollectionQuerier, user *MockUserQuerier) {
				wishlist := mocks.NewMockWishlistQuerier(ctrl)
				store.EXPECT().Tx(gomock.Any()).Return(tx, nil)
				tx.EXPECT().UserStore().Return(user).AnyTimes()
				tx.EXPECT().CollectionStore().Return(coll).AnyTimes()
				tx.EXPECT().WishlistStore().Return(wishlist).AnyTimes()
				user.EXPECT().Get(gomock.Any(), uint64(123)).Return(userstore.User{
					UserID: 123,
					Date:   pgtype.Timestamp{Time: time.Now().Add(-2 * time.Hour), Valid: true},
					Tokens: 5,
				}, nil)
				coll.EXPECT().ListIDs(gomock.Any(), uint64(123)).Return([]int64{}, nil)
				anime.EXPECT().RandomChar(gomock.Any()).Return(MediaCharacter{ID: 3, Name: "Char3", ImageURL: "img3"}, nil)
				coll.EXPECT().UpsertCharacter(gomock.Any(), gomock.Any()).Return(collectionstore.Character{}, nil)
				coll.EXPECT().Insert(gomock.Any(), gomock.Any()).Return(collectionstore.Collection{}, nil)
				wishlist.EXPECT().RemoveCharacterFromWishlist(gomock.Any(), gomock.Any()).Return(nil)
				user.EXPECT().UpdateDate(gomock.Any(), gomock.Any()).Return(errors.New("update date error"))
				tx.EXPECT().Rollback(gomock.Any()).Return(nil)
			},
			wantErr:     true,
			errContains: "update date error",
		},
		{
			name:   "token roll success",
			userID: 123,
			config: Config{RollCooldown: 5 * time.Hour, TokensNeeded: 3},
			setupMocks: func(store *MockProfileStore, tx *mocks.MockStorageStore, anime *MockAnimeService, coll *MockCollectionQuerier, user *MockUserQuerier) {
				wishlist := mocks.NewMockWishlistQuerier(ctrl)
				store.EXPECT().Tx(gomock.Any()).Return(tx, nil)
				tx.EXPECT().UserStore().Return(user).AnyTimes()
				tx.EXPECT().CollectionStore().Return(coll).AnyTimes()
				tx.EXPECT().WishlistStore().Return(wishlist).AnyTimes()
				user.EXPECT().Get(gomock.Any(), uint64(123)).Return(userstore.User{
					UserID: 123,
					Date:   pgtype.Timestamp{Time: time.Now().Add(-1 * time.Hour), Valid: true},
					Tokens: 5,
				}, nil)
				coll.EXPECT().ListIDs(gomock.Any(), uint64(123)).Return([]int64{}, nil)
				anime.EXPECT().RandomChar(gomock.Any()).Return(MediaCharacter{ID: 3, Name: "Char3", ImageURL: "img3"}, nil)
				coll.EXPECT().UpsertCharacter(gomock.Any(), gomock.Any()).Return(collectionstore.Character{}, nil)
				coll.EXPECT().Insert(gomock.Any(), gomock.Any()).Return(collectionstore.Collection{}, nil)
				wishlist.EXPECT().RemoveCharacterFromWishlist(gomock.Any(), gomock.Any()).Return(nil)
				user.EXPECT().UpdateTokens(gomock.Any(), userstore.UpdateTokensParams{
					UserID: 123,
					Tokens: -3,
				}).Return(userstore.User{}, nil)
				tx.EXPECT().Commit(gomock.Any()).Return(nil)
			},
			wantErr:  false,
			wantChar: MediaCharacter{ID: 3, Name: "Char3", ImageURL: "img3"},
		},
		{
			name:   "update tokens error",
			userID: 123,
			config: Config{RollCooldown: 5 * time.Hour, TokensNeeded: 3},
			setupMocks: func(store *MockProfileStore, tx *mocks.MockStorageStore, anime *MockAnimeService, coll *MockCollectionQuerier, user *MockUserQuerier) {
				wishlist := mocks.NewMockWishlistQuerier(ctrl)
				store.EXPECT().Tx(gomock.Any()).Return(tx, nil)
				tx.EXPECT().UserStore().Return(user).AnyTimes()
				tx.EXPECT().CollectionStore().Return(coll).AnyTimes()
				tx.EXPECT().WishlistStore().Return(wishlist).AnyTimes()
				user.EXPECT().Get(gomock.Any(), uint64(123)).Return(userstore.User{
					UserID: 123,
					Date:   pgtype.Timestamp{Time: time.Now().Add(-1 * time.Hour), Valid: true},
					Tokens: 5,
				}, nil)
				coll.EXPECT().ListIDs(gomock.Any(), uint64(123)).Return([]int64{}, nil)
				anime.EXPECT().RandomChar(gomock.Any()).Return(MediaCharacter{ID: 3, Name: "Char3", ImageURL: "img3"}, nil)
				coll.EXPECT().UpsertCharacter(gomock.Any(), gomock.Any()).Return(collectionstore.Character{}, nil)
				coll.EXPECT().Insert(gomock.Any(), gomock.Any()).Return(collectionstore.Collection{}, nil)
				wishlist.EXPECT().RemoveCharacterFromWishlist(gomock.Any(), gomock.Any()).Return(nil)
				user.EXPECT().UpdateTokens(gomock.Any(), gomock.Any()).Return(userstore.User{}, errors.New("update tokens error"))
				tx.EXPECT().Rollback(gomock.Any()).Return(nil)
			},
			wantErr:     true,
			errContains: "update tokens error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewMockProfileStore(ctrl)
			tx := mocks.NewMockStorageStore(ctrl)
			anime := NewMockAnimeService(ctrl)
			coll := NewMockCollectionQuerier(ctrl)
			user := NewMockUserQuerier(ctrl)
			tt.setupMocks(store, tx, anime, coll, user)

			got, err := Roll(context.Background(), store, anime, tt.config, tt.userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Roll() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
				t.Errorf("Roll() error = %v, want contains %v", err, tt.errContains)
			}
			if !tt.wantErr && got != tt.wantChar {
				t.Errorf("Roll() = %v, want %v", got, tt.wantChar)
			}
		})
	}
}
