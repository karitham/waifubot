package rest

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/karitham/waifubot/rest/api"
	"github.com/karitham/waifubot/storage/collectionstore"
	"github.com/karitham/waifubot/storage/userstore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testHandler struct {
	users       map[uint64]userstore.User
	collections map[uint64]map[int64]collectionstore.Collection
	characters  map[int64]collectionstore.Character
}

func newTestHandler() *testHandler {
	return &testHandler{
		users:       make(map[uint64]userstore.User),
		collections: make(map[uint64]map[int64]collectionstore.Collection),
		characters:  make(map[int64]collectionstore.Character),
	}
}

func (h *testHandler) addUser(id uint64, username, quote string, tokens int) {
	h.users[id] = userstore.User{
		ID:              int32(id),
		UserID:          id,
		DiscordUsername: username,
		Quote:           quote,
		Tokens:          int32(tokens),
	}
}

func (h *testHandler) addCharacter(id int64, name, image string) {
	h.characters[id] = collectionstore.Character{
		ID:    id,
		Name:  name,
		Image: image,
	}
}

func (h *testHandler) addToCollection(userID uint64, charID int64) {
	if h.collections[userID] == nil {
		h.collections[userID] = make(map[int64]collectionstore.Collection)
	}
	h.collections[userID][charID] = collectionstore.Collection{
		UserID:      userID,
		CharacterID: charID,
		Source:      "anilist",
		AcquiredAt:  pgtype.Timestamp{Time: time.Now(), Valid: true},
	}
}

func parseUserID(userID string) (uint64, error) {
	var id uint64
	for _, c := range userID {
		if c >= '0' && c <= '9' {
			id = id*10 + uint64(c-'0')
		} else {
			return 0, assert.AnError
		}
	}
	if id == 0 {
		return 0, assert.AnError
	}
	return id, nil
}

func (h *testHandler) GetUser(ctx context.Context, params api.GetUserParams) (api.GetUserRes, error) {
	id, err := parseUserID(params.UserID)
	if err != nil {
		return &api.GetUserBadRequest{Message: "invalid id provided", ErrorCode: "invalid_id", StatusCode: 400}, nil
	}

	user, ok := h.users[id]
	if !ok {
		return &api.GetUserNotFound{Message: "user not found", ErrorCode: "user_not_found", StatusCode: 404}, nil
	}

	return &api.User{
		ID:              params.UserID,
		Name:            "users/" + params.UserID,
		DiscordUsername: user.DiscordUsername,
		Quote:           api.NewOptString(user.Quote),
		Tokens:          user.Tokens,
		AnilistURL:      api.NewOptString(user.AnilistUrl),
		DiscordAvatar:   api.NewOptString(""),
	}, nil
}

func (h *testHandler) GetUserCollection(ctx context.Context, params api.GetUserCollectionParams) (api.GetUserCollectionRes, error) {
	id, err := parseUserID(params.UserID)
	if err != nil {
		return &api.GetUserCollectionBadRequest{Message: "invalid id provided", ErrorCode: "invalid_id", StatusCode: 400}, nil
	}

	if _, ok := h.users[id]; !ok {
		return &api.GetUserCollectionNotFound{Message: "user not found", ErrorCode: "user_not_found", StatusCode: 404}, nil
	}

	chars := h.collections[id]
	results := make([]api.Character, 0, len(chars))
	for _, col := range chars {
		char := h.characters[col.CharacterID]
		results = append(results, api.Character{
			ID:    char.ID,
			Name:  char.Name,
			Image: char.Image,
			Date:  col.AcquiredAt.Time,
		})
	}

	return &api.Collection{
		Characters: results,
		Total:      len(results),
	}, nil
}

func (h *testHandler) GetUserFavorite(ctx context.Context, params api.GetUserFavoriteParams) (api.GetUserFavoriteRes, error) {
	id, err := parseUserID(params.UserID)
	if err != nil {
		return &api.GetUserFavoriteBadRequest{Message: "invalid id provided", ErrorCode: "invalid_id", StatusCode: 400}, nil
	}

	user, ok := h.users[id]
	if !ok {
		return &api.GetUserFavoriteNotFound{Message: "user not found", ErrorCode: "user_not_found", StatusCode: 404}, nil
	}

	if !user.Favorite.Valid {
		return &api.GetUserFavoriteNoContent{}, nil
	}

	char := h.characters[user.Favorite.Int64]
	return &api.Character{
		ID:    char.ID,
		Name:  char.Name,
		Image: char.Image,
		Date:  time.Now(),
	}, nil
}

func (h *testHandler) GetUserWishlist(ctx context.Context, params api.GetUserWishlistParams) (api.GetUserWishlistRes, error) {
	id, err := parseUserID(params.UserID)
	if err != nil {
		return &api.GetUserWishlistBadRequest{Message: "invalid id provided", ErrorCode: "invalid_id", StatusCode: 400}, nil
	}

	if _, ok := h.users[id]; !ok {
		return &api.GetUserWishlistNotFound{Message: "user not found", ErrorCode: "user_not_found", StatusCode: 404}, nil
	}

	return &api.Wishlist{
		Characters: []api.Character{},
		Total:      0,
	}, nil
}

func (h *testHandler) ListUsers(ctx context.Context, params api.ListUsersParams) (api.ListUsersRes, error) {
	users := make([]api.User, 0, len(h.users))
	for _, u := range h.users {
		users = append(users, api.User{
			ID:              string(rune(u.UserID)),
			Name:            "users/" + string(rune(u.UserID)),
			DiscordUsername: u.DiscordUsername,
			Quote:           api.NewOptString(u.Quote),
			Tokens:          u.Tokens,
		})
	}
	return &api.UserList{Users: users, Total: len(users)}, nil
}

func (h *testHandler) GetUserLegacy(ctx context.Context, params api.GetUserLegacyParams) (api.GetUserLegacyRes, error) {
	userID := params.UserID
	var id uint64
	for _, c := range userID {
		if c >= '0' && c <= '9' {
			id = id*10 + uint64(c-'0')
		} else {
			return &api.GetUserLegacyBadRequest{Message: "invalid id provided", ErrorCode: "invalid_id", StatusCode: 400}, nil
		}
	}
	if id == 0 {
		return &api.GetUserLegacyBadRequest{Message: "invalid id provided", ErrorCode: "invalid_id", StatusCode: 400}, nil
	}

	user, ok := h.users[id]
	if !ok {
		return &api.GetUserLegacyNotFound{Message: "user not found", ErrorCode: "user_not_found", StatusCode: 404}, nil
	}

	return &api.Profile{
		ID:              params.UserID,
		DiscordUsername: user.DiscordUsername,
		Quote:           api.NewOptString(user.Quote),
		Tokens:          user.Tokens,
		AnilistURL:      api.NewOptString(user.AnilistUrl),
		DiscordAvatar:   api.NewOptString(""),
		Waifus:          []api.Character{},
	}, nil
}

func (h *testHandler) FindUserLegacy(ctx context.Context, params api.FindUserLegacyParams) (api.FindUserLegacyRes, error) {
	return &api.FindUserLegacyBadRequest{Message: "not implemented", ErrorCode: "not_implemented", StatusCode: 501}, nil
}

func (h *testHandler) NewError(ctx context.Context, err error) *api.InternalErrorStatusCode {
	return &api.InternalErrorStatusCode{}
}

func TestHTTP_GetUser_Success(t *testing.T) {
	h := newTestHandler()
	h.addUser(123, "testuser", "Hello World", 100)

	server, err := api.NewServer(h)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/123", nil)
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp api.User
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, "123", resp.ID)
	assert.Equal(t, "testuser", resp.DiscordUsername)
	quote, ok := resp.Quote.Get()
	assert.True(t, ok)
	assert.Equal(t, "Hello World", quote)
	assert.Equal(t, int32(100), resp.Tokens)
}

func TestHTTP_GetUser_NotFound(t *testing.T) {
	h := newTestHandler()

	server, err := api.NewServer(h)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/999", nil)
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	require.Equal(t, http.StatusNotFound, w.Code)
}

func TestHTTP_GetUser_InvalidID_NonNumeric(t *testing.T) {
	h := newTestHandler()

	server, err := api.NewServer(h)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/abc", nil)
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHTTP_GetUser_InvalidID_Zero(t *testing.T) {
	h := newTestHandler()

	server, err := api.NewServer(h)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/0", nil)
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHTTP_GetUserCollection_Success(t *testing.T) {
	h := newTestHandler()
	h.addUser(123, "testuser", "", 0)
	h.addCharacter(1, "Rem", "rem.jpg")
	h.addCharacter(2, "Emilia", "emilia.jpg")
	h.addToCollection(123, 1)
	h.addToCollection(123, 2)

	server, err := api.NewServer(h)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/123/collection", nil)
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp api.Collection
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, 2, resp.Total)
	assert.Len(t, resp.Characters, 2)
}

func TestHTTP_GetUserCollection_UserNotFound(t *testing.T) {
	h := newTestHandler()

	server, err := api.NewServer(h)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/999/collection", nil)
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	require.Equal(t, http.StatusNotFound, w.Code)
}

func TestHTTP_GetUserWishlist_Success(t *testing.T) {
	h := newTestHandler()
	h.addUser(123, "testuser", "", 0)

	server, err := api.NewServer(h)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/123/wishlist", nil)
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp api.Wishlist
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, 0, resp.Total)
}

func TestHTTP_ListUsers_Success(t *testing.T) {
	h := newTestHandler()
	h.addUser(1, "user1", "quote1", 10)
	h.addUser(2, "user2", "quote2", 20)

	server, err := api.NewServer(h)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp api.UserList
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, 2, resp.Total)
}

func TestHTTP_GetUserFavorite_NoFavorite(t *testing.T) {
	h := newTestHandler()
	h.addUser(123, "testuser", "", 0)

	server, err := api.NewServer(h)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/123/favorite", nil)
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	require.Equal(t, http.StatusNoContent, w.Code)
}
