package wishlist_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/karitham/waifubot/catalog"
	"github.com/karitham/waifubot/collection"
	"github.com/karitham/waifubot/collection/collectiontest"
	"github.com/karitham/waifubot/tracker"
	"github.com/karitham/waifubot/wishlist"
)

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
		KnownCharacterIDsFunc: func(_ context.Context, ids []catalog.CharacterID) ([]catalog.CharacterID, error) {
			return ids, nil
		},
		AddCharactersToWishlistFunc: func(_ context.Context, userID collection.UserID, ids []catalog.CharacterID) (int64, error) {
			calls++
			assert.Equal(t, collection.UserID(1), userID)
			assert.Len(t, ids, len(characters))
			return int64(len(ids)), nil
		},
	}

	count, err := wishlist.AddMediaToWishlist(t.Context(), mediaStub{characters: characters}, store, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, len(characters), count)
	assert.Equal(t, 1, calls, "the batch must be added in one transaction")
}

func TestAddMediaToWishlist_SkipsUnknown(t *testing.T) {
	characters := []tracker.Character{
		{ID: 1, Name: "known"},
		{ID: 2, Name: "unknown"},
	}

	var added []catalog.CharacterID
	store := &collectiontest.MockStore{
		KnownCharacterIDsFunc: func(_ context.Context, ids []catalog.CharacterID) ([]catalog.CharacterID, error) {
			assert.Len(t, ids, 2)
			return []catalog.CharacterID{1}, nil
		},
		AddCharactersToWishlistFunc: func(_ context.Context, userID collection.UserID, ids []catalog.CharacterID) (int64, error) {
			added = ids
			return int64(len(ids)), nil
		},
	}

	count, err := wishlist.AddMediaToWishlist(t.Context(), mediaStub{characters: characters}, store, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
	assert.Equal(t, []catalog.CharacterID{1}, added)
}
