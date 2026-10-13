package pg

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/karitham/waifubot/catalog"
	"github.com/karitham/waifubot/collection"
	"github.com/karitham/waifubot/storage"
)

// mockQuerier embeds the generated querier and overrides the methods the
// adapter tests exercise.
type mockQuerier struct {
	storage.Querier
	getUserCharacterWishlistFunc func(ctx context.Context, userID uint64) ([]storage.GetUserCharacterWishlistRow, error)
}

func (m *mockQuerier) GetUserCharacterWishlist(ctx context.Context, userID uint64) ([]storage.GetUserCharacterWishlistRow, error) {
	return m.getUserCharacterWishlistFunc(ctx, userID)
}

func TestGetUserCharacterWishlist(t *testing.T) {
	m := &mockQuerier{
		getUserCharacterWishlistFunc: func(_ context.Context, userID uint64) ([]storage.GetUserCharacterWishlistRow, error) {
			assert.Equal(t, uint64(123), userID)
			return []storage.GetUserCharacterWishlistRow{
				{
					ID:    456,
					Name:  "Test Character",
					Image: "http://example.com/image.jpg",
					Date:  pgtype.Timestamp{Time: time.Now().UTC(), Valid: true},
				},
			}, nil
		},
	}
	b := base{q: m}

	chars, err := b.GetUserCharacterWishlist(t.Context(), 123)
	require.NoError(t, err)
	assert.Len(t, chars, 1)
	assert.Equal(t, catalog.CharacterID(456), chars[0].ID)
	assert.Equal(t, "Test Character", chars[0].Name)
}

func TestGroupByUser(t *testing.T) {
	tests := []struct {
		name       string
		rows       []wishlistRow
		wantByUser map[collection.UserID][]catalog.CharacterID
	}{
		{
			name:       "no rows",
			rows:       nil,
			wantByUser: map[collection.UserID][]catalog.CharacterID{},
		},
		{
			name: "single user keeps insertion order",
			rows: []wishlistRow{
				{userID: 1, characterID: 10, name: "A", image: "a"},
				{userID: 1, characterID: 30, name: "C", image: "c"},
			},
			wantByUser: map[collection.UserID][]catalog.CharacterID{1: {10, 30}},
		},
		{
			name: "several users group and keep order",
			rows: []wishlistRow{
				{userID: 1, characterID: 10, name: "A", image: "a"},
				{userID: 2, characterID: 20, name: "B", image: "b"},
				{userID: 1, characterID: 30, name: "C", image: "c"},
			},
			wantByUser: map[collection.UserID][]catalog.CharacterID{1: {10, 30}, 2: {20}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			groups := groupByUser(tt.rows, func(r wishlistRow) wishlistRow { return r })

			require.Len(t, groups, len(tt.wantByUser))
			for _, g := range groups {
				var ids []catalog.CharacterID
				for _, c := range g.Characters {
					ids = append(ids, c.ID)
				}
				assert.Equal(t, tt.wantByUser[g.UserID], ids)
			}
		})
	}
}
