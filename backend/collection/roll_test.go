package collection

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Karitham/corde"
	"github.com/jackc/pgx/v5/pgtype"
	"go.uber.org/mock/gomock"

	"github.com/karitham/waifubot/storage/collectionstore"
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
		setupMocks  func(*MockProfileStore, *MockAnimeService, *MockCollectionQuerier, *MockUserQuerier)
		wantErr     bool
		wantChar    MediaCharacter
		errContains string
	}{
		{
			name:   "free roll success",
			userID: 123,
			config: Config{RollCooldown: time.Hour, TokensNeeded: 10},
			setupMocks: func(store *MockProfileStore, anime *MockAnimeService, coll *MockCollectionQuerier, user *MockUserQuerier) {
				wishlist := NewMockWishlistQuerier(ctrl)
				store.EXPECT().Tx(gomock.Any()).Return(store, nil)
				store.EXPECT().UserStore().Return(user).AnyTimes()
				store.EXPECT().CollectionStore().Return(coll).AnyTimes()
				store.EXPECT().WishlistStore().Return(wishlist).AnyTimes()
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
				store.EXPECT().Commit(gomock.Any()).Return(nil)
			},
			wantErr:  false,
			wantChar: MediaCharacter{ID: 3, Name: "Char3", ImageURL: "img3"},
		},
		{
			name:   "cooldown and insufficient tokens",
			userID: 123,
			config: Config{RollCooldown: time.Hour, TokensNeeded: 10},
			setupMocks: func(store *MockProfileStore, anime *MockAnimeService, coll *MockCollectionQuerier, user *MockUserQuerier) {
				store.EXPECT().Tx(gomock.Any()).Return(store, nil)
				store.EXPECT().UserStore().Return(user)
				user.EXPECT().Get(gomock.Any(), uint64(123)).Return(userstore.User{
					UserID: 123,
					Date:   pgtype.Timestamp{Time: time.Now().Add(-30 * time.Minute), Valid: true},
					Tokens: 5,
				}, nil)
				store.EXPECT().Rollback(gomock.Any()).Return(nil)
			},
			wantErr:     true,
			errContains: "You need another 5 tokens",
		},
		{
			name:   "get user error",
			userID: 123,
			config: Config{RollCooldown: time.Hour, TokensNeeded: 10},
			setupMocks: func(store *MockProfileStore, anime *MockAnimeService, coll *MockCollectionQuerier, user *MockUserQuerier) {
				store.EXPECT().Tx(gomock.Any()).Return(store, nil)
				store.EXPECT().UserStore().Return(user)
				user.EXPECT().Get(gomock.Any(), uint64(123)).Return(userstore.User{}, errors.New("user not found"))
				store.EXPECT().Rollback(gomock.Any()).Return(nil)
			},
			wantErr:     true,
			errContains: "user not found",
		},
		{
			name:   "list ids error",
			userID: 123,
			config: Config{RollCooldown: time.Hour, TokensNeeded: 10},
			setupMocks: func(store *MockProfileStore, anime *MockAnimeService, coll *MockCollectionQuerier, user *MockUserQuerier) {
				store.EXPECT().Tx(gomock.Any()).Return(store, nil)
				store.EXPECT().UserStore().Return(user).AnyTimes()
				store.EXPECT().CollectionStore().Return(coll)
				user.EXPECT().Get(gomock.Any(), uint64(123)).Return(userstore.User{
					UserID: 123,
					Date:   pgtype.Timestamp{Time: time.Now().Add(-2 * time.Hour), Valid: true},
					Tokens: 5,
				}, nil)
				coll.EXPECT().ListIDs(gomock.Any(), uint64(123)).Return(nil, errors.New("list error"))
				store.EXPECT().Rollback(gomock.Any()).Return(nil)
			},
			wantErr:     true,
			errContains: "list error",
		},
		{
			name:   "random char error",
			userID: 123,
			config: Config{RollCooldown: time.Hour, TokensNeeded: 10},
			setupMocks: func(store *MockProfileStore, anime *MockAnimeService, coll *MockCollectionQuerier, user *MockUserQuerier) {
				store.EXPECT().Tx(gomock.Any()).Return(store, nil)
				store.EXPECT().UserStore().Return(user).AnyTimes()
				store.EXPECT().CollectionStore().Return(coll)
				user.EXPECT().Get(gomock.Any(), uint64(123)).Return(userstore.User{
					UserID: 123,
					Date:   pgtype.Timestamp{Time: time.Now().Add(-2 * time.Hour), Valid: true},
					Tokens: 5,
				}, nil)
				coll.EXPECT().ListIDs(gomock.Any(), uint64(123)).Return([]int64{}, nil)
				anime.EXPECT().RandomChar(gomock.Any()).Return(MediaCharacter{}, errors.New("random error"))
				store.EXPECT().Rollback(gomock.Any()).Return(nil)
			},
			wantErr:     true,
			errContains: "random error",
		},
		{
			name:   "upsert error",
			userID: 123,
			config: Config{RollCooldown: time.Hour, TokensNeeded: 10},
			setupMocks: func(store *MockProfileStore, anime *MockAnimeService, coll *MockCollectionQuerier, user *MockUserQuerier) {
				store.EXPECT().Tx(gomock.Any()).Return(store, nil)
				store.EXPECT().UserStore().Return(user).AnyTimes()
				store.EXPECT().CollectionStore().Return(coll).AnyTimes()
				user.EXPECT().Get(gomock.Any(), uint64(123)).Return(userstore.User{
					UserID: 123,
					Date:   pgtype.Timestamp{Time: time.Now().Add(-2 * time.Hour), Valid: true},
					Tokens: 5,
				}, nil)
				coll.EXPECT().ListIDs(gomock.Any(), uint64(123)).Return([]int64{}, nil)
				anime.EXPECT().RandomChar(gomock.Any()).Return(MediaCharacter{ID: 3, Name: "Char3", ImageURL: "img3"}, nil)
				coll.EXPECT().UpsertCharacter(gomock.Any(), gomock.Any()).Return(collectionstore.Character{}, errors.New("upsert error"))
				store.EXPECT().Rollback(gomock.Any()).Return(nil)
			},
			wantErr:     true,
			errContains: "upsert error",
		},
		{
			name:   "insert error",
			userID: 123,
			config: Config{RollCooldown: time.Hour, TokensNeeded: 10},
			setupMocks: func(store *MockProfileStore, anime *MockAnimeService, coll *MockCollectionQuerier, user *MockUserQuerier) {
				store.EXPECT().Tx(gomock.Any()).Return(store, nil)
				store.EXPECT().UserStore().Return(user).AnyTimes()
				store.EXPECT().CollectionStore().Return(coll).AnyTimes()
				user.EXPECT().Get(gomock.Any(), uint64(123)).Return(userstore.User{
					UserID: 123,
					Date:   pgtype.Timestamp{Time: time.Now().Add(-2 * time.Hour), Valid: true},
					Tokens: 5,
				}, nil)
				coll.EXPECT().ListIDs(gomock.Any(), uint64(123)).Return([]int64{}, nil)
				anime.EXPECT().RandomChar(gomock.Any()).Return(MediaCharacter{ID: 3, Name: "Char3", ImageURL: "img3"}, nil)
				coll.EXPECT().UpsertCharacter(gomock.Any(), gomock.Any()).Return(collectionstore.Character{}, nil)
				coll.EXPECT().Insert(gomock.Any(), gomock.Any()).Return(collectionstore.Collection{}, errors.New("insert error"))
				store.EXPECT().Rollback(gomock.Any()).Return(nil)
			},
			wantErr:     true,
			errContains: "insert error",
		},
		{
			name:   "update date error",
			userID: 123,
			config: Config{RollCooldown: time.Hour, TokensNeeded: 10},
			setupMocks: func(store *MockProfileStore, anime *MockAnimeService, coll *MockCollectionQuerier, user *MockUserQuerier) {
				wishlist := NewMockWishlistQuerier(ctrl)
				store.EXPECT().Tx(gomock.Any()).Return(store, nil)
				store.EXPECT().UserStore().Return(user).AnyTimes()
				store.EXPECT().CollectionStore().Return(coll).AnyTimes()
				store.EXPECT().WishlistStore().Return(wishlist).AnyTimes()
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
				user.EXPECT().UpdateDate(gomock.Any(), gomock.Any()).Return(errors.New("update error"))
				store.EXPECT().Rollback(gomock.Any()).Return(nil)
			},
			wantErr:     true,
			errContains: "update error",
		},
		{
			name:   "consume tokens error",
			userID: 123,
			config: Config{RollCooldown: time.Hour, TokensNeeded: 10},
			setupMocks: func(store *MockProfileStore, anime *MockAnimeService, coll *MockCollectionQuerier, user *MockUserQuerier) {
				wishlist := NewMockWishlistQuerier(ctrl)
				store.EXPECT().Tx(gomock.Any()).Return(store, nil)
				store.EXPECT().UserStore().Return(user).AnyTimes()
				store.EXPECT().CollectionStore().Return(coll).AnyTimes()
				store.EXPECT().WishlistStore().Return(wishlist).AnyTimes()
				user.EXPECT().Get(gomock.Any(), uint64(123)).Return(userstore.User{
					UserID: 123,
					Date:   pgtype.Timestamp{Time: time.Now().Add(-30 * time.Minute), Valid: true},
					Tokens: 15,
				}, nil)
				coll.EXPECT().ListIDs(gomock.Any(), uint64(123)).Return([]int64{}, nil)
				anime.EXPECT().RandomChar(gomock.Any()).Return(MediaCharacter{ID: 3, Name: "Char3", ImageURL: "img3"}, nil)
				coll.EXPECT().UpsertCharacter(gomock.Any(), gomock.Any()).Return(collectionstore.Character{}, nil)
				coll.EXPECT().Insert(gomock.Any(), gomock.Any()).Return(collectionstore.Collection{}, nil)
				wishlist.EXPECT().RemoveCharacterFromWishlist(gomock.Any(), wishliststore.RemoveCharacterFromWishlistParams{UserID: 123, CharacterID: 3}).Return(nil)
				user.EXPECT().ConsumeTokens(gomock.Any(), gomock.Any()).Return(userstore.User{}, errors.New("consume error"))
				store.EXPECT().Rollback(gomock.Any()).Return(nil)
			},
			wantErr:     true,
			errContains: "consume error",
		},
		{
			name:   "commit error",
			userID: 123,
			config: Config{RollCooldown: time.Hour, TokensNeeded: 10},
			setupMocks: func(store *MockProfileStore, anime *MockAnimeService, coll *MockCollectionQuerier, user *MockUserQuerier) {
				wishlist := NewMockWishlistQuerier(ctrl)
				store.EXPECT().Tx(gomock.Any()).Return(store, nil)
				store.EXPECT().UserStore().Return(user).AnyTimes()
				store.EXPECT().CollectionStore().Return(coll).AnyTimes()
				store.EXPECT().WishlistStore().Return(wishlist).AnyTimes()
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
				store.EXPECT().Commit(gomock.Any()).Return(errors.New("commit error"))
				store.EXPECT().Rollback(gomock.Any()).Return(nil)
			},
			wantErr:     true,
			errContains: "commit error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewMockProfileStore(ctrl)
			anime := NewMockAnimeService(ctrl)
			coll := NewMockCollectionQuerier(ctrl)
			user := NewMockUserQuerier(ctrl)
			tt.setupMocks(store, anime, coll, user)

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
