package collection

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Karitham/corde"
	"github.com/jackc/pgx/v5/pgtype"
	"go.uber.org/mock/gomock"

	"github.com/karitham/waifubot/storage/collectionstore"
	"github.com/karitham/waifubot/storage/mocks"
	"github.com/karitham/waifubot/storage/wishliststore"
)

func TestGive(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name       string
		from       corde.Snowflake
		to         corde.Snowflake
		charID     int64
		setupMocks func(*MockProfileStore, *MockCollectionQuerier)
		want       Character
		assertErr  func(error) bool
	}{
		{
			name:   "success",
			from:   123,
			to:     456,
			charID: 1,
			setupMocks: func(store *MockProfileStore, coll *MockCollectionQuerier) {
				wishlist := mocks.NewMockWishlistQuerier(ctrl)
				store.EXPECT().CollectionStore().Return(coll).AnyTimes()
				store.EXPECT().WishlistStore().Return(wishlist).AnyTimes()
				coll.EXPECT().Get(gomock.Any(), collectionstore.GetParams{ID: 1, UserID: uint64(123)}).Return(collectionstore.GetRow{
					ID:     1,
					Date:   pgtype.Timestamp{Time: time.Now(), Valid: true},
					Image:  "img1",
					Name:   "Char1",
					Source: "ROLL",
				}, nil)
				coll.EXPECT().Get(gomock.Any(), collectionstore.GetParams{ID: 1, UserID: uint64(456)}).Return(collectionstore.GetRow{}, errors.New("not found"))
				coll.EXPECT().Give(gomock.Any(), gomock.Any()).Return(collectionstore.Collection{}, nil)
				wishlist.EXPECT().RemoveCharacterFromWishlist(gomock.Any(), wishliststore.RemoveCharacterFromWishlistParams{UserID: 456, CharacterID: 1}).Return(nil)
			},
			want:      Character{ID: 1, Name: "Char1", Image: "img1", Type: "ROLL", UserID: 456},
			assertErr: nil,
		},
		{
			name:   "to already owns",
			from:   123,
			to:     456,
			charID: 1,
			setupMocks: func(store *MockProfileStore, coll *MockCollectionQuerier) {
				store.EXPECT().CollectionStore().Return(coll).AnyTimes()
				coll.EXPECT().Get(gomock.Any(), collectionstore.GetParams{ID: 1, UserID: uint64(123)}).Return(collectionstore.GetRow{
					ID:     1,
					Date:   pgtype.Timestamp{Time: time.Now(), Valid: true},
					Image:  "img1",
					Name:   "Char1",
					Source: "ROLL",
				}, nil)
				coll.EXPECT().Get(gomock.Any(), collectionstore.GetParams{ID: 1, UserID: uint64(456)}).Return(collectionstore.GetRow{
					ID:     1,
					Date:   pgtype.Timestamp{Time: time.Now(), Valid: true},
					Image:  "img1",
					Name:   "Char1",
					Source: "ROLL",
				}, nil)
			},
			want:      Character{},
			assertErr: func(err error) bool { return strings.Contains(err.Error(), "to user already owns char 1") },
		},
		{
			name:   "from_does_not_own",
			from:   123,
			to:     456,
			charID: 1,
			setupMocks: func(store *MockProfileStore, coll *MockCollectionQuerier) {
				store.EXPECT().CollectionStore().Return(coll)
				coll.EXPECT().Get(gomock.Any(), collectionstore.GetParams{ID: 1, UserID: uint64(123)}).Return(collectionstore.GetRow{}, sql.ErrNoRows)
			},
			want:      Character{},
			assertErr: func(err error) bool { return errors.Is(err, ErrUserDoesNotOwnCharacter) },
		},
		{
			name:   "give_error",
			from:   123,
			to:     456,
			charID: 1,
			setupMocks: func(store *MockProfileStore, coll *MockCollectionQuerier) {
				store.EXPECT().CollectionStore().Return(coll).AnyTimes()
				coll.EXPECT().Get(gomock.Any(), collectionstore.GetParams{ID: 1, UserID: uint64(123)}).Return(collectionstore.GetRow{
					ID:     1,
					Date:   pgtype.Timestamp{Time: time.Now(), Valid: true},
					Image:  "img1",
					Name:   "Char1",
					Source: "ROLL",
				}, nil)
				coll.EXPECT().Get(gomock.Any(), collectionstore.GetParams{ID: 1, UserID: uint64(456)}).Return(collectionstore.GetRow{}, errors.New("not found"))
				coll.EXPECT().Give(gomock.Any(), gomock.Any()).Return(collectionstore.Collection{}, errors.New("give error"))
			},
			want:      Character{},
			assertErr: func(err error) bool { return strings.Contains(err.Error(), "error giving char") },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewMockProfileStore(ctrl)
			coll := NewMockCollectionQuerier(ctrl)
			tt.setupMocks(store, coll)

			got, err := Give(context.Background(), store, tt.from, tt.to, tt.charID)
			if tt.assertErr == nil {
				if err != nil {
					t.Errorf("Give() expected no error, got %v", err)
					return
				}
				if got.ID != tt.want.ID || got.Name != tt.want.Name || got.Image != tt.want.Image || got.Type != tt.want.Type || got.UserID != tt.want.UserID {
					t.Errorf("Give() = %v, want %v", got, tt.want)
				}
			} else {
				if err == nil {
					t.Errorf("Give() expected error, got nil")
					return
				}
				if !tt.assertErr(err) {
					t.Errorf("Give() error assertion failed: %v", err)
				}
			}
		})
	}
}
