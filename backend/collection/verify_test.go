package collection

import (
	"context"
	"database/sql"
	"errors"
	"reflect"
	"testing"

	"github.com/Karitham/corde"
	"go.uber.org/mock/gomock"

	"github.com/karitham/waifubot/storage/collectionstore"
)

func TestCheckOwnership(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name       string
		userID     corde.Snowflake
		charID     int64
		setupMocks func(*MockProfileStore, *MockCollectionQuerier)
		want       bool
		wantChar   collectionstore.Character
		assertErr  func(error) bool
	}{
		{
			name:   "success owns",
			userID: 123,
			charID: 1,
			setupMocks: func(store *MockProfileStore, coll *MockCollectionQuerier) {
				store.EXPECT().CollectionStore().Return(coll)
				coll.EXPECT().Get(gomock.Any(), collectionstore.GetParams{ID: 1, UserID: 123}).Return(collectionstore.GetRow{
					ID:    1,
					Name:  "Char1",
					Image: "img1",
				}, nil)
			},
			want: true,
			wantChar: collectionstore.Character{
				ID:    1,
				Name:  "Char1",
				Image: "img1",
			},
			assertErr: nil,
		},
		{
			name:   "does not own",
			userID: 123,
			charID: 2,
			setupMocks: func(store *MockProfileStore, coll *MockCollectionQuerier) {
				store.EXPECT().CollectionStore().Return(coll)
				coll.EXPECT().Get(gomock.Any(), collectionstore.GetParams{ID: 2, UserID: 123}).Return(collectionstore.GetRow{}, sql.ErrNoRows)
			},
			want:      false,
			wantChar:  collectionstore.Character{},
			assertErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewMockProfileStore(ctrl)
			coll := NewMockCollectionQuerier(ctrl)
			tt.setupMocks(store, coll)

			got, gotChar, err := CheckOwnership(context.Background(), store, tt.userID, tt.charID)
			if tt.assertErr == nil {
				if err != nil {
					t.Errorf("CheckOwnership() expected no error, got %v", err)
					return
				}
				if got != tt.want {
					t.Errorf("CheckOwnership() got = %v, want %v", got, tt.want)
				}
				if gotChar != tt.wantChar {
					t.Errorf("CheckOwnership() gotChar = %v, want %v", gotChar, tt.wantChar)
				}
			} else {
				if err == nil {
					t.Errorf("CheckOwnership() expected error, got nil")
					return
				}
				if !tt.assertErr(err) {
					t.Errorf("CheckOwnership() error assertion failed: %v", err)
				}
			}
		})
	}
}

func TestSearchGlobalCharacters(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name       string
		term       string
		setupMocks func(*MockProfileStore, *MockCollectionQuerier)
		want       []collectionstore.Character
		wantErr    bool
	}{
		{
			name: "success with results",
			term: "test",
			setupMocks: func(store *MockProfileStore, coll *MockCollectionQuerier) {
				store.EXPECT().CollectionStore().Return(coll)
				coll.EXPECT().SearchGlobalCharacters(gomock.Any(), collectionstore.SearchGlobalCharactersParams{Term: "test", Lim: 25}).Return([]collectionstore.Character{
					{ID: 1, Name: "Char1", Image: "img1"},
					{ID: 2, Name: "Char2", Image: "img2"},
				}, nil)
			},
			want: []collectionstore.Character{
				{ID: 1, Name: "Char1", Image: "img1"},
				{ID: 2, Name: "Char2", Image: "img2"},
			},
			wantErr: false,
		},
		{
			name: "success empty",
			term: "empty",
			setupMocks: func(store *MockProfileStore, coll *MockCollectionQuerier) {
				store.EXPECT().CollectionStore().Return(coll)
				coll.EXPECT().SearchGlobalCharacters(gomock.Any(), collectionstore.SearchGlobalCharactersParams{Term: "empty", Lim: 25}).Return([]collectionstore.Character{}, nil)
			},
			want:    []collectionstore.Character{},
			wantErr: false,
		},
		{
			name: "error",
			term: "error",
			setupMocks: func(store *MockProfileStore, coll *MockCollectionQuerier) {
				store.EXPECT().CollectionStore().Return(coll)
				coll.EXPECT().SearchGlobalCharacters(gomock.Any(), collectionstore.SearchGlobalCharactersParams{Term: "error", Lim: 25}).Return(nil, errors.New("search failed"))
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewMockProfileStore(ctrl)
			coll := NewMockCollectionQuerier(ctrl)
			tt.setupMocks(store, coll)

			got, err := SearchGlobalCharacters(context.Background(), store, tt.term)
			if (err != nil) != tt.wantErr {
				t.Errorf("SearchGlobalCharacters() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SearchGlobalCharacters() got = %v, want %v", got, tt.want)
			}
		})
	}
}
