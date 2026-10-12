package collection_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/karitham/waifubot/collection"
	"github.com/karitham/waifubot/collection/collectiontest"
)

func TestUserProfile_AbsentUser(t *testing.T) {
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

	profile, err := collection.UserProfile(t.Context(), store, 7)
	require.NoError(t, err)
	assert.Equal(t, collection.UserID(7), profile.User.UserID)
	assert.Equal(t, int32(0), profile.User.Tokens)
	assert.Nil(t, profile.FavoriteCharacter)
	assert.Equal(t, 0, profile.CharacterCount)
	assert.False(t, created, "reads must not create the user")
}
