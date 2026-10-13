package collection_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/karitham/waifubot/catalog"
	"github.com/karitham/waifubot/collection"
	"github.com/karitham/waifubot/collection/collectiontest"
	"github.com/karitham/waifubot/tracker"
)

func TestRemoveFromWishlist(t *testing.T) {
	created := false
	store := &collectiontest.MockStore{
		CreateUserFunc: func(_ context.Context, _ collection.UserID) error {
			created = true
			return nil
		},
		RemoveFromWishlistFunc: func(_ context.Context, userID collection.UserID, charID catalog.CharacterID) error {
			assert.Equal(t, collection.UserID(1), userID)
			assert.Equal(t, catalog.CharacterID(42), charID)
			return nil
		},
	}

	require.NoError(t, collection.RemoveFromWishlist(t.Context(), store, 1, 42))
	assert.False(t, created, "removal must not materialize the user")
	assert.Equal(t, 1, store.CommitCalls)
}

func TestRemoveAllFromWishlist(t *testing.T) {
	created := false
	store := &collectiontest.MockStore{
		CreateUserFunc: func(_ context.Context, _ collection.UserID) error {
			created = true
			return nil
		},
		RemoveAllFromWishlistFunc: func(_ context.Context, userID collection.UserID) error {
			assert.Equal(t, collection.UserID(1), userID)
			return nil
		},
	}

	require.NoError(t, collection.RemoveAllFromWishlist(t.Context(), store, 1))
	assert.False(t, created, "removal must not materialize the user")
	assert.Equal(t, 1, store.CommitCalls)
}

func TestAddToWishlist(t *testing.T) {
	tests := []struct {
		name         string
		addErr       error
		wantCommit   int
		wantRollback int
	}{
		{
			name:       "success",
			wantCommit: 1,
		},
		{
			name:         "insert failure rolls back",
			addErr:       errors.New("database on fire"),
			wantRollback: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &collectiontest.MockStore{
				AddCharactersToWishlistFunc: func(_ context.Context, userID collection.UserID, characterIDs []catalog.CharacterID) (int64, error) {
					assert.Equal(t, collection.UserID(1), userID)
					assert.Equal(t, []catalog.CharacterID{42}, characterIDs)
					return 1, tt.addErr
				},
			}

			added, err := collection.AddToWishlist(t.Context(), store, 1, []catalog.CharacterID{42})

			if tt.addErr != nil {
				require.ErrorIs(t, err, tt.addErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, int64(1), added)
			}
			assert.Equal(t, tt.wantCommit, store.CommitCalls)
			assert.Equal(t, tt.wantRollback, store.RollbackCalls)
		})
	}
}

// mediaStub implements collection.AnimeService for tests.
type mediaStub struct {
	characters []tracker.Character
}

func (s mediaStub) GetMediaCharacters(context.Context, tracker.MediaID) ([]tracker.Character, error) {
	return s.characters, nil
}

func (s mediaStub) SearchMedia(context.Context, string) ([]tracker.Media, error) { return nil, nil }

func (s mediaStub) Anime(context.Context, string) ([]tracker.Media, error) { return nil, nil }

func (s mediaStub) Manga(context.Context, string) ([]tracker.Media, error) { return nil, nil }

func (s mediaStub) User(context.Context, string) ([]tracker.User, error) { return nil, nil }

func (s mediaStub) Character(context.Context, string) ([]tracker.Character, error) {
	return nil, nil
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

	count, err := collection.AddMediaToWishlist(t.Context(), store, mediaStub{characters: characters}, 1, 10)
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

	count, err := collection.AddMediaToWishlist(t.Context(), store, mediaStub{characters: characters}, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
	assert.Equal(t, []catalog.CharacterID{1}, added)
}
