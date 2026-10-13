package pg

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/karitham/waifubot/catalog"
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
