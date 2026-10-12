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
)

func TestUserProfile_Favorite(t *testing.T) {
	tests := []struct {
		name      string
		lookupErr error
		wantErr   bool
	}{
		{name: "lookup failure is returned", lookupErr: errors.New("database on fire"), wantErr: true},
		{name: "dangling favorite is tolerated", lookupErr: collection.ErrNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &collectiontest.MockStore{
				GetUserFunc: func(_ context.Context, userID collection.UserID) (collection.User, error) {
					return collection.User{UserID: userID, FavoriteCharacterID: 42}, nil
				},
				GetCharacterByIDFunc: func(_ context.Context, _ catalog.CharacterID) (catalog.Character, error) {
					return catalog.Character{}, tt.lookupErr
				},
			}

			profile, err := collection.UserProfile(t.Context(), store, 7)

			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Nil(t, profile.FavoriteCharacter)
		})
	}
}

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
