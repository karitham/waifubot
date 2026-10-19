package rest

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/karitham/waifubot/collection"
	"github.com/karitham/waifubot/collection/collectiontest"
	"github.com/karitham/waifubot/rest/api"
)

func TestFindUserRejectsOverlongAniListQuery(t *testing.T) {
	store := &collectiontest.MockStore{
		GetUserByAnilistFunc: func(context.Context, string) (collection.User, error) {
			t.Fatal("overlong query must not reach the database")
			return collection.User{}, nil
		},
	}
	server := New(store, nil)

	res, err := server.FindUser(t.Context(), api.FindUserParams{
		Anilist: api.NewOptString(strings.Repeat("a", maxFindQueryLen+1)),
	})

	require.NoError(t, err)
	assert.IsType(t, &api.FindUserBadRequest{}, res)
}

func TestFindUserRejectsNormalizedAniListQueryOverflow(t *testing.T) {
	store := &collectiontest.MockStore{
		GetUserByAnilistFunc: func(context.Context, string) (collection.User, error) {
			t.Fatal("normalized overlong query must not reach the database")
			return collection.User{}, nil
		},
	}
	server := New(store, nil)

	res, err := server.FindUser(t.Context(), api.FindUserParams{
		Anilist: api.NewOptString(strings.Repeat("a", maxFindQueryLen-len("https://anilist.co/user/")+1)),
	})

	require.NoError(t, err)
	assert.IsType(t, &api.FindUserBadRequest{}, res)
}

func TestFindUserNormalizesAniListQuery(t *testing.T) {
	var got string
	store := &collectiontest.MockStore{
		GetUserByAnilistFunc: func(_ context.Context, anilistURL string) (collection.User, error) {
			got = anilistURL
			return collection.User{UserID: 42}, nil
		},
	}
	server := New(store, nil)

	res, err := server.FindUser(t.Context(), api.FindUserParams{
		Anilist: api.NewOptString("animefan"),
	})

	require.NoError(t, err)
	assert.Equal(t, "https://anilist.co/user/animefan", got)
	assert.Equal(t, "42", res.(*api.UserIdResponse).ID)
}
