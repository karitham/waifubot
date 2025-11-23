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
)

func TestGive(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name        string
		from        corde.Snowflake
		to          corde.Snowflake
		charID      int64
		setupMocks  func(*MockProfileStore, *MockCollectionQuerier)
		want        Character
		wantErr     bool
		errContains string
	}{
		{
			name:   "success",
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
				coll.EXPECT().Give(gomock.Any(), gomock.Any()).Return(collectionstore.Collection{}, nil)
			},
			want: Character{
				ID:     1,
				UserID: 456,
				Date:   time.Now(),
				Image:  "img1",
				Name:   "Char1",
				Type:   "ROLL",
			},
			wantErr: false,
		},
		{
			name:   "from does not own",
			from:   123,
			to:     456,
			charID: 1,
			setupMocks: func(store *MockProfileStore, coll *MockCollectionQuerier) {
				store.EXPECT().CollectionStore().Return(coll)
				coll.EXPECT().Get(gomock.Any(), collectionstore.GetParams{ID: 1, UserID: uint64(123)}).Return(collectionstore.GetRow{}, errors.New("not found"))
			},
			want:        Character{},
			wantErr:     true,
			errContains: "from user does not own char 1",
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
			want:        Character{},
			wantErr:     true,
			errContains: "to user already owns char 1",
		},
		{
			name:   "give error",
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
			want:        Character{},
			wantErr:     true,
			errContains: "error giving char",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewMockProfileStore(ctrl)
			coll := NewMockCollectionQuerier(ctrl)
			tt.setupMocks(store, coll)

			got, err := Give(context.Background(), store, tt.from, tt.to, tt.charID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Give() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
				t.Errorf("Give() error = %v, want contains %v", err, tt.errContains)
			}
			if !tt.wantErr {
				// Approximate time check
				if got.ID != tt.want.ID || got.UserID != tt.want.UserID || got.Image != tt.want.Image || got.Name != tt.want.Name || got.Type != tt.want.Type {
					t.Errorf("Give() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}
