package wishlistpg_test

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/karitham/waifubot/catalog"
	"github.com/karitham/waifubot/storage/wishlistpg"
	"github.com/karitham/waifubot/storage/wishliststore"
	"github.com/karitham/waifubot/storage/wishlisttest"
)

func TestWishlist(t *testing.T) {
	tests := []struct {
		name  string
		setup func(m *wishlisttest.MockQuerier)
		run   func(t *testing.T, s *wishlistpg.Pg)
	}{
		{
			name: "get_user_character_wishlist",
			setup: func(m *wishlisttest.MockQuerier) {
				m.GetUserCharacterWishlistFunc = func(_ context.Context, _ uint64) ([]wishliststore.GetUserCharacterWishlistRow, error) {
					return []wishliststore.GetUserCharacterWishlistRow{
						{
							ID:    456,
							Name:  "Test Character",
							Image: "http://example.com/image.jpg",
							Date:  pgtype.Timestamp{Time: time.Now().UTC(), Valid: true},
						},
					}, nil
				}
			},
			run: func(t *testing.T, s *wishlistpg.Pg) {
				chars, err := s.GetUserCharacterWishlist(t.Context(), 123)
				require.NoError(t, err)
				assert.Len(t, chars, 1)
				assert.Equal(t, catalog.CharacterID(456), chars[0].ID)
				assert.Equal(t, "Test Character", chars[0].Name)
			},
		},
		{
			name: "remove_all_from_wishlist",
			setup: func(m *wishlisttest.MockQuerier) {
				m.RemoveAllFromWishlistFunc = func(_ context.Context, userID uint64) error {
					assert.Equal(t, uint64(123), userID)
					return nil
				}
			},
			run: func(t *testing.T, s *wishlistpg.Pg) {
				err := s.RemoveAllFromWishlist(t.Context(), 123)
				require.NoError(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &wishlisttest.MockQuerier{}
			if tt.setup != nil {
				tt.setup(mock)
			}

			s := wishlistpg.New(mock)
			tt.run(t, s)
		})
	}
}
