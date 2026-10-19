package collection_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/karitham/waifubot/catalog"
	"github.com/karitham/waifubot/collection"
	"github.com/karitham/waifubot/collection/collectiontest"
)

func TestNormalizeAnilistURL(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
		ok   bool
	}{
		{name: "canonical", raw: "https://anilist.co/user/testuser", want: "https://anilist.co/user/testuser", ok: true},
		{name: "trailing slash", raw: "https://anilist.co/user/testuser/", want: "https://anilist.co/user/testuser/", ok: true},
		{name: "http", raw: "http://anilist.co/user/testuser", ok: false},
		{name: "foreign host", raw: "https://example.com/user/testuser", ok: false},
		{name: "query", raw: "https://anilist.co/user/testuser?x=1", ok: false},
		{name: "fragment", raw: "https://anilist.co/user/testuser#x", ok: false},
		{name: "markdown injection", raw: "https://anilist.co/user/x) [Click](https://evil.example)", ok: false},
		{name: "javascript scheme", raw: "javascript://anilist.co/user/testuser", ok: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := collection.NormalizeAnilistURL(tt.raw)
			assert.Equal(t, tt.ok, ok)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestWritesRejectZeroUserID(t *testing.T) {
	store := &collectiontest.MockStore{}

	err := collection.SetQuote(t.Context(), store, 0, "hi")

	require.ErrorIs(t, err, collection.ErrInvalidUserID)
	assert.Zero(t, store.CommitCalls)
}

func TestProfileEdits(t *testing.T) {
	tests := []struct {
		name  string
		setup func(m *collectiontest.MockStore)
		apply func(ctx context.Context, store collection.DB) error
	}{
		{
			name: "quote",
			setup: func(m *collectiontest.MockStore) {
				m.UpdateQuoteFunc = func(_ context.Context, userID collection.UserID, quote string) error {
					assert.Equal(t, collection.UserID(1), userID)
					assert.Equal(t, "hi", quote)
					return nil
				}
			},
			apply: func(ctx context.Context, store collection.DB) error {
				return collection.SetQuote(ctx, store, 1, "hi")
			},
		},
		{
			name: "anilist url",
			setup: func(m *collectiontest.MockStore) {
				m.UpdateAnilistURLFunc = func(_ context.Context, userID collection.UserID, url string) error {
					assert.Equal(t, collection.UserID(1), userID)
					assert.Equal(t, "https://anilist.co/user/x", url)
					return nil
				}
			},
			apply: func(ctx context.Context, store collection.DB) error {
				return collection.SetAnilistURL(ctx, store, 1, "https://anilist.co/user/x")
			},
		},
		{
			name: "favorite",
			setup: func(m *collectiontest.MockStore) {
				m.SetFavoriteFunc = func(_ context.Context, userID collection.UserID, charID catalog.CharacterID) error {
					assert.Equal(t, collection.UserID(1), userID)
					assert.Equal(t, catalog.CharacterID(42), charID)
					return nil
				}
			},
			apply: func(ctx context.Context, store collection.DB) error {
				return collection.SetFavorite(ctx, store, 1, 42)
			},
		},
		{
			name: "discord info",
			setup: func(m *collectiontest.MockStore) {
				m.UpdateDiscordInfoFunc = func(_ context.Context, userID collection.UserID, username, avatar string, _ time.Time) error {
					assert.Equal(t, collection.UserID(1), userID)
					assert.Equal(t, "testuser", username)
					assert.Equal(t, "avatar_hash", avatar)
					return nil
				}
			},
			apply: func(ctx context.Context, store collection.DB) error {
				return collection.SetDiscordInfo(ctx, store, 1, "testuser", "avatar_hash", time.Now())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			created := false
			store := &collectiontest.MockStore{
				GetUserFunc: func(_ context.Context, _ collection.UserID) (collection.User, error) {
					return collection.User{}, collection.ErrNotFound
				},
				CreateUserFunc: func(_ context.Context, _ collection.UserID) error {
					created = true
					return nil
				},
			}
			tt.setup(store)

			require.NoError(t, tt.apply(t.Context(), store))

			assert.True(t, created, "the edit must materialize the user")
			assert.Equal(t, 1, store.CommitCalls)
		})
	}
}
