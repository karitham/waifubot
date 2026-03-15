package collection

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"go.uber.org/mock/gomock"

	"github.com/karitham/waifubot/storage/collectionstore"
	"github.com/karitham/waifubot/storage/dropstore"
	"github.com/karitham/waifubot/storage/mocks"
	"github.com/karitham/waifubot/storage/userstore"
	"github.com/karitham/waifubot/storage/wishliststore"
)

func TestClaim(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name        string
		userID      uint64
		channelID   uint64
		charName    string
		setupMocks  func(*MockProfileStore, *mocks.MockStorageStore, *mocks.MockDropstoreQuerier, *MockCollectionQuerier, *MockUserQuerier, *mocks.MockWishlistQuerier)
		wantErr     bool
		errContains string
		errIs       error
		wantChar    Character
	}{
		{
			name:      "successful claim",
			userID:    123,
			channelID: 456,
			charName:  "Test Character",
			setupMocks: func(store *MockProfileStore, tx *mocks.MockStorageStore, drop *mocks.MockDropstoreQuerier, coll *MockCollectionQuerier, user *MockUserQuerier, wishlist *mocks.MockWishlistQuerier) {
				store.EXPECT().Tx(gomock.Any()).Return(tx, nil)
				tx.EXPECT().DropStore().Return(drop).AnyTimes()
				tx.EXPECT().UserStore().Return(user).AnyTimes()
				tx.EXPECT().CollectionStore().Return(coll).AnyTimes()
				tx.EXPECT().WishlistStore().Return(wishlist).AnyTimes()

				drop.EXPECT().GetDropForUpdate(gomock.Any(), uint64(456)).Return(dropstore.Character{
					ID:         1,
					Name:       "Test Character",
					Image:      "test.jpg",
					MediaTitle: "Test Anime",
				}, nil)

				user.EXPECT().Get(gomock.Any(), uint64(123)).Return(userstore.User{UserID: 123}, nil)

				coll.EXPECT().UpsertCharacter(gomock.Any(), collectionstore.UpsertCharacterParams{
					ID: 1, Name: "Test Character", Image: "test.jpg",
				}).Return(collectionstore.Character{}, nil)

				coll.EXPECT().Insert(gomock.Any(), gomock.Any()).Return(collectionstore.Collection{}, nil)

				wishlist.EXPECT().RemoveCharacterFromWishlist(gomock.Any(), wishliststore.RemoveCharacterFromWishlistParams{
					UserID: 123, CharacterID: 1,
				}).Return(nil)

				drop.EXPECT().DeleteDrop(gomock.Any(), uint64(456)).Return(nil)

				tx.EXPECT().Commit(gomock.Any()).Return(nil)
			},
			wantErr: false,
			wantChar: Character{
				ID:    1,
				Name:  "Test Character",
				Image: "test.jpg",
				Type:  "CLAIM",
			},
		},
		{
			name:      "no drop in channel",
			userID:    123,
			channelID: 456,
			charName:  "Test Character",
			setupMocks: func(store *MockProfileStore, tx *mocks.MockStorageStore, drop *mocks.MockDropstoreQuerier, coll *MockCollectionQuerier, user *MockUserQuerier, wishlist *mocks.MockWishlistQuerier) {
				store.EXPECT().Tx(gomock.Any()).Return(tx, nil)
				tx.EXPECT().DropStore().Return(drop).AnyTimes()

				drop.EXPECT().GetDropForUpdate(gomock.Any(), uint64(456)).Return(dropstore.Character{}, sql.ErrNoRows)

				tx.EXPECT().Rollback(gomock.Any()).Return(nil)
			},
			wantErr:     true,
			errIs:       ErrNoDropInChannel,
			errContains: "no drop",
		},
		{
			name:      "wrong character name",
			userID:    123,
			channelID: 456,
			charName:  "Wrong Name",
			setupMocks: func(store *MockProfileStore, tx *mocks.MockStorageStore, drop *mocks.MockDropstoreQuerier, coll *MockCollectionQuerier, user *MockUserQuerier, wishlist *mocks.MockWishlistQuerier) {
				store.EXPECT().Tx(gomock.Any()).Return(tx, nil)
				tx.EXPECT().DropStore().Return(drop).AnyTimes()

				drop.EXPECT().GetDropForUpdate(gomock.Any(), uint64(456)).Return(dropstore.Character{
					ID:         1,
					Name:       "Test Character",
					Image:      "test.jpg",
					MediaTitle: "Test Anime",
				}, nil)

				tx.EXPECT().Rollback(gomock.Any()).Return(nil)
			},
			wantErr:     true,
			errIs:       ErrWrongCharacterName,
			errContains: "wrong character name",
		},
		{
			name:      "new user created",
			userID:    123,
			channelID: 456,
			charName:  "Test Character",
			setupMocks: func(store *MockProfileStore, tx *mocks.MockStorageStore, drop *mocks.MockDropstoreQuerier, coll *MockCollectionQuerier, user *MockUserQuerier, wishlist *mocks.MockWishlistQuerier) {
				store.EXPECT().Tx(gomock.Any()).Return(tx, nil)
				tx.EXPECT().DropStore().Return(drop).AnyTimes()
				tx.EXPECT().UserStore().Return(user).AnyTimes()
				tx.EXPECT().CollectionStore().Return(coll).AnyTimes()
				tx.EXPECT().WishlistStore().Return(wishlist).AnyTimes()

				drop.EXPECT().GetDropForUpdate(gomock.Any(), uint64(456)).Return(dropstore.Character{
					ID:         1,
					Name:       "Test Character",
					Image:      "test.jpg",
					MediaTitle: "Test Anime",
				}, nil)

				user.EXPECT().Get(gomock.Any(), uint64(123)).Return(userstore.User{}, pgx.ErrNoRows)
				user.EXPECT().Create(gomock.Any(), uint64(123)).Return(nil)

				coll.EXPECT().UpsertCharacter(gomock.Any(), gomock.Any()).Return(collectionstore.Character{}, nil)
				coll.EXPECT().Insert(gomock.Any(), gomock.Any()).Return(collectionstore.Collection{}, nil)
				wishlist.EXPECT().RemoveCharacterFromWishlist(gomock.Any(), gomock.Any()).Return(nil)
				drop.EXPECT().DeleteDrop(gomock.Any(), uint64(456)).Return(nil)
				tx.EXPECT().Commit(gomock.Any()).Return(nil)
			},
			wantErr: false,
			wantChar: Character{
				ID:    1,
				Name:  "Test Character",
				Image: "test.jpg",
				Type:  "CLAIM",
			},
		},
		{
			name:      "user get error",
			userID:    123,
			channelID: 456,
			charName:  "Test Character",
			setupMocks: func(store *MockProfileStore, tx *mocks.MockStorageStore, drop *mocks.MockDropstoreQuerier, coll *MockCollectionQuerier, user *MockUserQuerier, wishlist *mocks.MockWishlistQuerier) {
				store.EXPECT().Tx(gomock.Any()).Return(tx, nil)
				tx.EXPECT().DropStore().Return(drop).AnyTimes()
				tx.EXPECT().UserStore().Return(user).AnyTimes()

				drop.EXPECT().GetDropForUpdate(gomock.Any(), uint64(456)).Return(dropstore.Character{
					ID:         1,
					Name:       "Test Character",
					Image:      "test.jpg",
					MediaTitle: "Test Anime",
				}, nil)

				user.EXPECT().Get(gomock.Any(), uint64(123)).Return(userstore.User{}, errors.New("db error"))

				tx.EXPECT().Rollback(gomock.Any()).Return(nil)
			},
			wantErr:     true,
			errContains: "failed to get user",
		},
		{
			name:      "commit error",
			userID:    123,
			channelID: 456,
			charName:  "Test Character",
			setupMocks: func(store *MockProfileStore, tx *mocks.MockStorageStore, drop *mocks.MockDropstoreQuerier, coll *MockCollectionQuerier, user *MockUserQuerier, wishlist *mocks.MockWishlistQuerier) {
				store.EXPECT().Tx(gomock.Any()).Return(tx, nil)
				tx.EXPECT().DropStore().Return(drop).AnyTimes()
				tx.EXPECT().UserStore().Return(user).AnyTimes()
				tx.EXPECT().CollectionStore().Return(coll).AnyTimes()
				tx.EXPECT().WishlistStore().Return(wishlist).AnyTimes()

				drop.EXPECT().GetDropForUpdate(gomock.Any(), uint64(456)).Return(dropstore.Character{
					ID:         1,
					Name:       "Test Character",
					Image:      "test.jpg",
					MediaTitle: "Test Anime",
				}, nil)

				user.EXPECT().Get(gomock.Any(), uint64(123)).Return(userstore.User{UserID: 123}, nil)
				coll.EXPECT().UpsertCharacter(gomock.Any(), gomock.Any()).Return(collectionstore.Character{}, nil)
				coll.EXPECT().Insert(gomock.Any(), gomock.Any()).Return(collectionstore.Collection{}, nil)
				wishlist.EXPECT().RemoveCharacterFromWishlist(gomock.Any(), gomock.Any()).Return(nil)
				drop.EXPECT().DeleteDrop(gomock.Any(), uint64(456)).Return(nil)

				tx.EXPECT().Commit(gomock.Any()).Return(errors.New("commit failed"))
				tx.EXPECT().Rollback(gomock.Any()).Return(nil)
			},
			wantErr:     true,
			errContains: "failed to commit transaction",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewMockProfileStore(ctrl)
			tx := mocks.NewMockStorageStore(ctrl)
			drop := mocks.NewMockDropstoreQuerier(ctrl)
			coll := NewMockCollectionQuerier(ctrl)
			user := NewMockUserQuerier(ctrl)
			wishlist := mocks.NewMockWishlistQuerier(ctrl)

			tt.setupMocks(store, tx, drop, coll, user, wishlist)

			got, err := Claim(context.Background(), store, tt.userID, tt.channelID, tt.charName)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error but got nil")
				}
				if tt.errIs != nil && !errors.Is(err, tt.errIs) {
					t.Errorf("expected error %v, got %v", tt.errIs, err)
				}
				if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("expected error containing %q, got %q", tt.errContains, err.Error())
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if got.ID != tt.wantChar.ID {
					t.Errorf("got ID %d, want %d", got.ID, tt.wantChar.ID)
				}
				if got.Name != tt.wantChar.Name {
					t.Errorf("got Name %q, want %q", got.Name, tt.wantChar.Name)
				}
				if got.Image != tt.wantChar.Image {
					t.Errorf("got Image %q, want %q", got.Image, tt.wantChar.Image)
				}
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
