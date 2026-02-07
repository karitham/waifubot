package collection

import (
	"context"
	"errors"
	"testing"

	"github.com/Karitham/corde"
	"go.uber.org/mock/gomock"

	"github.com/karitham/waifubot/storage/collectionstore"
	"github.com/karitham/waifubot/storage/userstore"
)

func TestExchange(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name       string
		userID     corde.Snowflake
		charID     int64
		setupMocks func(*MockProfileStore, *MockProfileStore, *MockCollectionQuerier, *MockUserQuerier)
		want       collectionstore.Character
		wantErr    bool
	}{
		{
			name:   "success",
			userID: 123,
			charID: 1,
			setupMocks: func(store, tx *MockProfileStore, coll *MockCollectionQuerier, user *MockUserQuerier) {
				store.EXPECT().Tx(gomock.Any()).Return(tx, nil)
				tx.EXPECT().CollectionStore().Return(coll).AnyTimes()
				tx.EXPECT().UserStore().Return(user).AnyTimes()
				coll.EXPECT().GetByID(gomock.Any(), int64(1)).Return(collectionstore.Character{
					ID:    1,
					Name:  "Char1",
					Image: "img1",
				}, nil)
				coll.EXPECT().Delete(gomock.Any(), collectionstore.DeleteParams{UserID: uint64(123), CharacterID: 1}).Return(collectionstore.Collection{}, nil)
				user.EXPECT().UpdateTokens(gomock.Any(), userstore.UpdateTokensParams{
					UserID: uint64(123),
					Tokens: 1,
				}).Return(userstore.User{}, nil)
				tx.EXPECT().Commit(gomock.Any()).Return(nil)
			},
			want: collectionstore.Character{
				ID:    1,
				Name:  "Char1",
				Image: "img1",
			},
			wantErr: false,
		},
		{
			name:   "GetByID error",
			userID: 123,
			charID: 1,
			setupMocks: func(store, tx *MockProfileStore, coll *MockCollectionQuerier, user *MockUserQuerier) {
				store.EXPECT().Tx(gomock.Any()).Return(tx, nil)
				tx.EXPECT().CollectionStore().Return(coll).AnyTimes()
				coll.EXPECT().GetByID(gomock.Any(), int64(1)).Return(collectionstore.Character{}, errors.New("get error"))
				tx.EXPECT().Rollback(gomock.Any()).Return(nil)
			},
			want:    collectionstore.Character{},
			wantErr: true,
		},
		{
			name:   "Delete error",
			userID: 123,
			charID: 1,
			setupMocks: func(store, tx *MockProfileStore, coll *MockCollectionQuerier, user *MockUserQuerier) {
				store.EXPECT().Tx(gomock.Any()).Return(tx, nil)
				tx.EXPECT().CollectionStore().Return(coll).AnyTimes()
				tx.EXPECT().UserStore().Return(user).AnyTimes()
				coll.EXPECT().GetByID(gomock.Any(), int64(1)).Return(collectionstore.Character{
					ID:    1,
					Name:  "Char1",
					Image: "img1",
				}, nil)
				coll.EXPECT().Delete(gomock.Any(), collectionstore.DeleteParams{UserID: uint64(123), CharacterID: 1}).Return(collectionstore.Collection{}, errors.New("delete error"))
				tx.EXPECT().Rollback(gomock.Any()).Return(nil)
			},
			want:    collectionstore.Character{},
			wantErr: true,
		},
		{
			name:   "IncTokens error",
			userID: 123,
			charID: 1,
			setupMocks: func(store, tx *MockProfileStore, coll *MockCollectionQuerier, user *MockUserQuerier) {
				store.EXPECT().Tx(gomock.Any()).Return(tx, nil)
				tx.EXPECT().CollectionStore().Return(coll).AnyTimes()
				tx.EXPECT().UserStore().Return(user).AnyTimes()
				coll.EXPECT().GetByID(gomock.Any(), int64(1)).Return(collectionstore.Character{
					ID:    1,
					Name:  "Char1",
					Image: "img1",
				}, nil)
				coll.EXPECT().Delete(gomock.Any(), collectionstore.DeleteParams{UserID: uint64(123), CharacterID: 1}).Return(collectionstore.Collection{}, nil)
				user.EXPECT().UpdateTokens(gomock.Any(), userstore.UpdateTokensParams{
					UserID: uint64(123),
					Tokens: 1,
				}).Return(userstore.User{}, errors.New("inc error"))
				tx.EXPECT().Rollback(gomock.Any()).Return(nil)
			},
			want:    collectionstore.Character{},
			wantErr: true,
		},
		{
			name:   "Commit error",
			userID: 123,
			charID: 1,
			setupMocks: func(store, tx *MockProfileStore, coll *MockCollectionQuerier, user *MockUserQuerier) {
				store.EXPECT().Tx(gomock.Any()).Return(tx, nil)
				tx.EXPECT().CollectionStore().Return(coll).AnyTimes()
				tx.EXPECT().UserStore().Return(user).AnyTimes()
				coll.EXPECT().GetByID(gomock.Any(), int64(1)).Return(collectionstore.Character{
					ID:    1,
					Name:  "Char1",
					Image: "img1",
				}, nil)
				coll.EXPECT().Delete(gomock.Any(), collectionstore.DeleteParams{UserID: uint64(123), CharacterID: 1}).Return(collectionstore.Collection{}, nil)
				user.EXPECT().UpdateTokens(gomock.Any(), userstore.UpdateTokensParams{
					UserID: uint64(123),
					Tokens: 1,
				}).Return(userstore.User{}, nil)
				tx.EXPECT().Commit(gomock.Any()).Return(errors.New("commit error"))
				tx.EXPECT().Rollback(gomock.Any()).Return(nil)
			},
			want: collectionstore.Character{
				ID:    1,
				Name:  "Char1",
				Image: "img1",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewMockProfileStore(ctrl)
			tx := NewMockProfileStore(ctrl)
			coll := NewMockCollectionQuerier(ctrl)
			user := NewMockUserQuerier(ctrl)
			tt.setupMocks(store, tx, coll, user)

			got, err := Exchange(context.Background(), store, tt.userID, tt.charID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Exchange() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.ID != tt.want.ID || got.Name != tt.want.Name || got.Image != tt.want.Image {
					t.Errorf("Exchange() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}
