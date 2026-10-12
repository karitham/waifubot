package wishlist_test

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/karitham/waifubot/catalog"
	"github.com/karitham/waifubot/collection"
	"github.com/karitham/waifubot/collection/collectiontest"
	"github.com/karitham/waifubot/storage/wishliststore"
	"github.com/karitham/waifubot/storage/wishlisttest"
	"github.com/karitham/waifubot/tracker"
	"github.com/karitham/waifubot/wishlist"
)

func TestWishlist(t *testing.T) {
	tests := []struct {
		name  string
		setup func(m *wishlisttest.MockQuerier)
		run   func(t *testing.T, s wishlist.Store)
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
			run: func(t *testing.T, s wishlist.Store) {
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
			run: func(t *testing.T, s wishlist.Store) {
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

			s := wishlist.New(mock)
			tt.run(t, s)
		})
	}
}

// mediaStub implements wishlist.MediaService for tests.
type mediaStub struct {
	characters []tracker.Character
}

func (s mediaStub) SearchMedia(context.Context, string) ([]tracker.Media, error) {
	return nil, nil
}

func (s mediaStub) GetMediaCharacters(context.Context, tracker.MediaID) ([]tracker.Character, error) {
	return s.characters, nil
}

func TestAddMediaToWishlist(t *testing.T) {
	characters := make([]tracker.Character, 120)
	for i := range characters {
		characters[i] = tracker.Character{ID: catalog.CharacterID(i + 1), Name: "char"}
	}

	calls := 0
	store := &collectiontest.MockStore{
		AddCharactersToWishlistFunc: func(_ context.Context, userID collection.UserID, ids []catalog.CharacterID) error {
			calls++
			assert.Equal(t, collection.UserID(1), userID)
			assert.Len(t, ids, len(characters))
			return nil
		},
	}

	count, err := wishlist.AddMediaToWishlist(t.Context(), mediaStub{characters: characters}, store, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, len(characters), count)
	assert.Equal(t, 1, calls, "the batch must be added in one transaction")
}
