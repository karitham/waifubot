package collection

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Karitham/corde"
	"github.com/jackc/pgx/v5/pgtype"
	"go.uber.org/mock/gomock"

	"github.com/karitham/waifubot/storage/collectionstore"
)

func TestCharacters(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name       string
		userID     corde.Snowflake
		setupMocks func(*MockProfileStore, *MockCollectionQuerier)
		want       []Character
		wantErr    bool
	}{
		{
			name:   "success with characters",
			userID: 123,
			setupMocks: func(store *MockProfileStore, coll *MockCollectionQuerier) {
				store.EXPECT().CollectionStore().Return(coll)
				coll.EXPECT().List(gomock.Any(), uint64(123)).Return([]collectionstore.ListRow{
					{
						ID:     1,
						Name:   "Char1",
						Image:  "img1",
						Source: "ROLL",
						Date:   pgtype.Timestamp{Time: time.Now(), Valid: true},
					},
					{
						ID:     2,
						Name:   "Char2",
						Image:  "img2",
						Source: "GIVE",
						Date:   pgtype.Timestamp{Time: time.Now(), Valid: true},
					},
				}, nil)
			},
			want: []Character{
				{
					ID:     1,
					Name:   "Char1",
					Image:  "img1",
					Type:   "ROLL",
					UserID: 123,
					Date:   time.Now(),
				},
				{
					ID:     2,
					Name:   "Char2",
					Image:  "img2",
					Type:   "GIVE",
					UserID: 123,
					Date:   time.Now(),
				},
			},
			wantErr: false,
		},
		{
			name:   "success empty",
			userID: 123,
			setupMocks: func(store *MockProfileStore, coll *MockCollectionQuerier) {
				store.EXPECT().CollectionStore().Return(coll)
				coll.EXPECT().List(gomock.Any(), uint64(123)).Return([]collectionstore.ListRow{}, nil)
			},
			want:    []Character{},
			wantErr: false,
		},
		{
			name:   "list error",
			userID: 123,
			setupMocks: func(store *MockProfileStore, coll *MockCollectionQuerier) {
				store.EXPECT().CollectionStore().Return(coll)
				coll.EXPECT().List(gomock.Any(), uint64(123)).Return(nil, errors.New("list error"))
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

			got, err := Characters(context.Background(), store, tt.userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Characters() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if len(got) != len(tt.want) {
					t.Errorf("Characters() len = %v, want %v", len(got), len(tt.want))
				}
				for i, c := range got {
					if c.ID != tt.want[i].ID || c.Name != tt.want[i].Name || c.Image != tt.want[i].Image || c.Type != tt.want[i].Type || c.UserID != tt.want[i].UserID {
						t.Errorf("Characters()[%d] = %v, want %v", i, c, tt.want[i])
					}
				}
			}
		})
	}
}
