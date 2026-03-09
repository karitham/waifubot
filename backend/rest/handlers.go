//go:generate go run github.com/ogen-go/ogen/cmd/ogen@latest --target ./api --clean ../../openapi.yaml
//go:generate sqlc generate -f ../storage/sqlc.yaml

package rest

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/Karitham/corde"

	"github.com/karitham/waifubot/discord"
	"github.com/karitham/waifubot/services"
	"github.com/karitham/waifubot/storage"
	"github.com/karitham/waifubot/storage/collectionstore"
	"github.com/karitham/waifubot/storage/userstore"
	"github.com/karitham/waifubot/wishlist"

	"github.com/karitham/waifubot/rest/api"
)

type Server struct {
	db             storage.Store
	discordService *services.DiscordService
}

func New(db storage.Store, discordService *services.DiscordService) *Server {
	return &Server{
		db:             db,
		discordService: discordService,
	}
}

func (s *Server) NewError(ctx context.Context, err error) *api.InternalErrorStatusCode {
	return &api.InternalErrorStatusCode{
		StatusCode: 500,
		Response: api.Error{
			Message:    err.Error(),
			ErrorCode:  "internal_error",
			StatusCode: 500,
		},
	}
}

// GetUser returns a user's profile metadata (resource-oriented, no collection/favorite)
func (s *Server) GetUser(ctx context.Context, params api.GetUserParams) (api.GetUserRes, error) {
	id, err := strconv.ParseUint(params.UserID, 10, 64)
	if err != nil || id == 0 {
		return &api.GetUserBadRequest{Message: "invalid id provided", ErrorCode: "invalid_id", StatusCode: 400}, nil
	}

	user, err := s.getUserData(ctx, id)
	if err != nil {
		if errors.Is(err, errUserNotFound) {
			return &api.GetUserNotFound{Message: "user not found", ErrorCode: "user_not_found", StatusCode: 404}, nil
		}
		return nil, err
	}

	return user, nil
}

// ListUsers lists users with optional filtering and pagination
func (s *Server) ListUsers(ctx context.Context, params api.ListUsersParams) (api.ListUsersRes, error) {
	// Build query params
	listParams := userstore.ListParams{
		UserID:          0,
		DiscordUsername: "",
		AnilistUrl:      "",
		UsernamePrefix:  "",
		PageSize:        20,
		PageOffset:      0,
	}

	if v, ok := params.ID.Get(); ok && v != "" {
		id, err := strconv.ParseUint(v, 10, 64)
		if err == nil {
			listParams.UserID = int64(id)
		}
	}

	if v, ok := params.DiscordUsername.Get(); ok {
		listParams.DiscordUsername = v
	}

	if v, ok := params.AnilistURL.Get(); ok {
		listParams.AnilistUrl = v.String()
	}

	if v, ok := params.UsernamePrefix.Get(); ok {
		listParams.UsernamePrefix = v
	}

	// Handle pagination
	if v, ok := params.PageSize.Get(); ok && v > 0 {
		if v > 100 {
			v = 100
		}
		listParams.PageSize = int32(v)
	}

	if v, ok := params.PageToken.Get(); ok && v != "" {
		offset, err := decodePageToken(v)
		if err == nil {
			listParams.PageOffset = int32(offset)
		}
	}

	// Get total count
	countParams := userstore.CountFilteredParams{
		UserID:          listParams.UserID,
		DiscordUsername: listParams.DiscordUsername,
		AnilistUrl:      listParams.AnilistUrl,
		UsernamePrefix:  listParams.UsernamePrefix,
	}
	total, err := s.db.UserStore().CountFiltered(ctx, countParams)
	if err != nil {
		return nil, err
	}

	// Get users
	users, err := s.db.UserStore().List(ctx, listParams)
	if err != nil {
		return nil, err
	}

	// Build response
	userList := make([]api.User, len(users))
	for i, u := range users {
		userList[i] = s.mapUserToAPI(u)
	}

	result := &api.UserList{
		Users: userList,
		Total: int(total),
	}

	// Add next page token if there are more results
	if int(total) > len(users)+int(listParams.PageOffset) {
		nextOffset := int(listParams.PageOffset) + len(users)
		result.NextPageToken = api.NewOptString(encodePageToken(nextOffset))
	}

	return result, nil
}

// GetUserCollection returns a user's character collection with cursor-based pagination
func (s *Server) GetUserCollection(ctx context.Context, params api.GetUserCollectionParams) (api.GetUserCollectionRes, error) {
	id, err := strconv.ParseUint(params.UserID, 10, 64)
	if err != nil || id == 0 {
		return &api.GetUserCollectionBadRequest{
			Message:    "invalid id provided",
			ErrorCode:  "invalid_id",
			StatusCode: 400,
		}, nil
	}

	// Check if user exists
	_, err = s.db.UserStore().Get(ctx, id)
	if err != nil {
		return &api.GetUserCollectionNotFound{
			Message:    "user not found",
			ErrorCode:  "user_not_found",
			StatusCode: 404,
		}, nil
	}

	// Build pagination params
	pageSize := 50
	if v, ok := params.PageSize.Get(); ok && v > 0 {
		if v > 100 {
			v = 100
		}
		pageSize = v
	}

	// Get search query (q parameter)
	searchTerm := ""
	if v, ok := params.Q.Get(); ok {
		searchTerm = strings.TrimSpace(v)
	}

	// Get sort order
	orderBy := api.OrderByDate
	if v, ok := params.OrderBy.Get(); ok {
		orderBy = v
	}

	// Ensure we have a valid orderBy
	if orderBy == "" {
		orderBy = api.OrderByDate
	}

	// Get sort direction
	direction := api.DirectionDesc
	if v, ok := params.Direction.Get(); ok {
		direction = v
	}

	// Ensure we have a valid direction
	if direction == "" {
		direction = api.DirectionDesc
	}

	// Parse cursor if provided
	var cursor collectionPageToken
	if v, ok := params.PageToken.Get(); ok && v != "" {
		cursor, err = decodeCollectionPageToken(v)
		if err != nil {
			return &api.GetUserCollectionBadRequest{
				Message:    "invalid page_token",
				ErrorCode:  "invalid_token",
				StatusCode: 400,
			}, nil
		}
	}

	// Fetch one extra to determine if there's a next page
	limit := pageSize + 1

	// Parse order_by into column and direction
	orderColumn := "date"
	ascending := direction == api.DirectionAsc
	switch orderBy {
	case api.OrderByName:
		orderColumn = "name"
	case api.OrderByAnilistID:
		orderColumn = "id"
	case api.OrderByDate:
		orderColumn = "date"
	}

	// Execute dynamic paginated query
	// Type assert to access the concrete *Queries type with custom methods
	collectionQueries, ok := s.db.CollectionStore().(*collectionstore.Queries)
	if !ok {
		return nil, fmt.Errorf("collection store type assertion failed")
	}
	rows, err := collectionQueries.ListPaginatedDynamic(ctx, collectionstore.ListPaginatedParams{
		UserID:     id,
		SearchTerm: searchTerm,
		OrderBy:    orderColumn,
		Ascending:  ascending,
		CursorDate: &cursor.LastDate,
		CursorID:   cursor.LastID,
		CursorName: cursor.LastName,
		Limit:      int32(limit),
	})
	if err != nil {
		return nil, err
	}

	// Determine if there's a next page
	hasMore := len(rows) > pageSize
	if hasMore {
		rows = rows[:pageSize]
	}

	// Convert to API characters
	characters := make([]api.Character, len(rows))
	for i, c := range rows {
		characters[i] = api.Character{
			ID:    c.ID,
			Name:  c.Name,
			Image: c.Image,
			Type:  api.CharacterType(c.Source),
			Date:  c.Date.Time,
		}
	}

	// Get true total count with search filter
	totalCount, err := collectionQueries.CountFiltered(ctx, id, searchTerm)
	if err != nil {
		return nil, fmt.Errorf("failed to get total count: %w", err)
	}

	result := &api.Collection{
		Characters: characters,
		Total:      int(totalCount),
	}

	// Generate next page token if there are more results
	if hasMore && len(rows) > 0 {
		last := rows[len(rows)-1]
		nextToken := collectionPageToken{
			LastDate: last.Date.Time,
			LastID:   last.ID,
			LastName: last.Name,
			OrderBy:  string(orderBy),
			Search:   searchTerm,
		}
		result.NextPageToken = api.NewOptString(encodeCollectionPageToken(nextToken))
	}

	return result, nil
}

// GetUserFavorite returns a user's favorite character (204 if none)
func (s *Server) GetUserFavorite(ctx context.Context, params api.GetUserFavoriteParams) (api.GetUserFavoriteRes, error) {
	id, err := strconv.ParseUint(params.UserID, 10, 64)
	if err != nil || id == 0 {
		return &api.GetUserFavoriteBadRequest{
			Message:    "invalid id provided",
			ErrorCode:  "invalid_id",
			StatusCode: 400,
		}, nil
	}

	// Check if user exists and get their favorite
	u, err := s.db.UserStore().Get(ctx, id)
	if err != nil {
		return &api.GetUserFavoriteNotFound{
			Message:    "user not found",
			ErrorCode:  "user_not_found",
			StatusCode: 404,
		}, nil
	}

	// If no favorite set, return 204 No Content
	if !u.Favorite.Valid {
		return &api.GetUserFavoriteNoContent{}, nil
	}

	// Get the favorite character details from collection
	char, err := s.db.CollectionStore().Get(ctx, collectionstore.GetParams{
		ID:     u.Favorite.Int64,
		UserID: id,
	})
	if err != nil {
		return nil, err
	}

	return &api.Character{
		ID:    char.ID,
		Name:  char.Name,
		Image: char.Image,
		Type:  api.CharacterType(char.Source),
		Date:  char.Date.Time,
	}, nil
}

// GetUserWishlist returns a user's wishlist
func (s *Server) GetUserWishlist(ctx context.Context, params api.GetUserWishlistParams) (api.GetUserWishlistRes, error) {
	id, err := strconv.ParseUint(params.UserID, 10, 64)
	if err != nil || id == 0 {
		return &api.GetUserWishlistBadRequest{
			Message:    "invalid id provided",
			ErrorCode:  "invalid_id",
			StatusCode: 400,
		}, nil
	}

	// Check if user exists
	_, err = s.db.UserStore().Get(ctx, id)
	if err != nil {
		return &api.GetUserWishlistNotFound{
			Message:    "user not found",
			ErrorCode:  "user_not_found",
			StatusCode: 404,
		}, nil
	}

	// Get wishlist
	wishlistChars, err := wishlist.GetUserWishlist(ctx, wishlist.New(s.db.WishlistStore()), id)
	if err != nil {
		return nil, err
	}

	// Build pagination params
	pageSize := 20
	if v, ok := params.PageSize.Get(); ok && v > 0 {
		if v > 100 {
			v = 100
		}
		pageSize = v
	}

	offset := 0
	if v, ok := params.PageToken.Get(); ok && v != "" {
		offset, _ = decodePageToken(v)
	}

	// Apply pagination
	end := min(offset+pageSize, len(wishlistChars))
	paginatedChars := wishlistChars[offset:end]

	characters := make([]api.Character, len(paginatedChars))
	for i, c := range paginatedChars {
		t, _ := time.Parse(time.RFC3339, c.Date)
		characters[i] = api.Character{
			Date:  t,
			Name:  c.Name,
			Image: c.Image,
			ID:    c.ID,
		}
	}

	result := &api.Wishlist{
		Characters: characters,
		Total:      len(wishlistChars),
	}

	if end < len(wishlistChars) {
		result.NextPageToken = api.NewOptString(encodePageToken(end))
	}

	return result, nil
}

// GetUserLegacy is the deprecated /user/{userID} endpoint that returns full profile with collection
func (s *Server) GetUserLegacy(ctx context.Context, params api.GetUserLegacyParams) (api.GetUserLegacyRes, error) {
	id, err := strconv.ParseUint(params.UserID, 10, 64)
	if err != nil || id == 0 {
		return &api.GetUserLegacyBadRequest{
			Message:    "invalid id provided",
			ErrorCode:  "invalid_id",
			StatusCode: 400,
		}, nil
	}

	profile, err := s.getUserProfileData(ctx, id)
	if err != nil {
		if errors.Is(err, errUserNotFound) {
			return &api.GetUserLegacyNotFound{
				Message:    "user not found",
				ErrorCode:  "user_not_found",
				StatusCode: 404,
			}, nil
		}
		return nil, err
	}

	return profile, nil
}

// FindUserLegacy is the deprecated /user/find endpoint
func (s *Server) FindUserLegacy(ctx context.Context, params api.FindUserLegacyParams) (api.FindUserLegacyRes, error) {
	anilistVal, anilistSet := params.Anilist.Get()
	discordVal, discordSet := params.Discord.Get()

	if !anilistSet && !discordSet {
		return &api.FindUserLegacyBadRequest{
			Message:    "anilist or discord query param is required",
			ErrorCode:  "missing_query_param",
			StatusCode: 400,
		}, nil
	}

	resp, err := s.findUserByQuery(ctx, anilistVal, discordVal, anilistSet)
	if err != nil {
		if errors.Is(err, errUserNotFound) {
			return &api.FindUserLegacyNotFound{
				Message:    "user not found",
				ErrorCode:  "user_not_found",
				StatusCode: 404,
			}, nil
		}
		return nil, err
	}

	return resp, nil
}

var errUserNotFound = errors.New("user not found")

func (s *Server) getUserData(ctx context.Context, id uint64) (*api.User, error) {
	if s.discordService != nil {
		ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		if err := s.discordService.UpdateIfNeeded(ctx, s.db, corde.Snowflake(id)); err != nil {
			slog.With("err", err).Warn("failed to update profile data")
		}
	}

	u, err := s.db.UserStore().Get(ctx, id)
	if err != nil {
		return nil, errUserNotFound
	}

	return s.mapUserToAPIPtr(u), nil
}

func (s *Server) getUserProfileData(ctx context.Context, id uint64) (*api.Profile, error) {
	if s.discordService != nil {
		ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		if err := s.discordService.UpdateIfNeeded(ctx, s.db, corde.Snowflake(id)); err != nil {
			slog.With("err", err).Warn("failed to update profile data")
		}
	}

	u, err := s.db.UserStore().Get(ctx, id)
	if err != nil {
		return nil, err
	}

	chars, err := s.db.CollectionStore().List(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.mapUserToProfile(u, chars), nil
}

func (s *Server) findUserByQuery(ctx context.Context, anilist, discord string, useAnilist bool) (*api.UserIdResponse, error) {
	var user userstore.User
	var err error

	if useAnilist {
		user, err = s.db.UserStore().GetByAnilist(ctx, normalizeAnilistURL(anilist))
	} else {
		user, err = s.db.UserStore().GetByDiscordUsername(ctx, discord)
	}

	if err != nil || user.UserID == 0 {
		return nil, errUserNotFound
	}

	return &api.UserIdResponse{
		ID: fmt.Sprintf("%d", user.UserID),
	}, nil
}

func (s *Server) mapUserToAPI(u userstore.User) api.User {
	return api.User{
		Name:            fmt.Sprintf("users/%d", u.UserID),
		ID:              fmt.Sprintf("%d", u.UserID),
		Quote:           api.NewOptString(u.Quote),
		Tokens:          u.Tokens,
		AnilistURL:      api.NewOptString(u.AnilistUrl),
		DiscordUsername: u.DiscordUsername,
		DiscordAvatar:   api.NewOptString(discord.DiscordAvatarURL(u.UserID, u.DiscordAvatar)),
	}
}

func (s *Server) mapUserToAPIPtr(u userstore.User) *api.User {
	return &api.User{
		Name:            fmt.Sprintf("users/%d", u.UserID),
		ID:              fmt.Sprintf("%d", u.UserID),
		Quote:           api.NewOptString(u.Quote),
		Tokens:          u.Tokens,
		AnilistURL:      api.NewOptString(u.AnilistUrl),
		DiscordUsername: u.DiscordUsername,
		DiscordAvatar:   api.NewOptString(discord.DiscordAvatarURL(u.UserID, u.DiscordAvatar)),
	}
}

func (s *Server) mapUserToProfile(u userstore.User, list []collectionstore.ListRow) *api.Profile {
	waifus := make([]api.Character, 0, len(list))
	var fav api.OptCharacter
	for _, entry := range list {
		c := api.Character{
			ID:    entry.ID,
			Name:  entry.Name,
			Image: entry.Image,
			Type:  api.CharacterType(entry.Source),
			Date:  entry.Date.Time,
		}

		if u.Favorite.Valid && entry.ID == u.Favorite.Int64 {
			fav = api.NewOptCharacter(c)
		}

		waifus = append(waifus, c)
	}

	return &api.Profile{
		ID:              fmt.Sprintf("%d", u.UserID),
		Quote:           api.NewOptString(u.Quote),
		Tokens:          u.Tokens,
		Favorite:        fav,
		AnilistURL:      api.NewOptString(u.AnilistUrl),
		DiscordUsername: u.DiscordUsername,
		DiscordAvatar:   api.NewOptString(discord.DiscordAvatarURL(u.UserID, u.DiscordAvatar)),
		Waifus:          waifus,
	}
}

func normalizeAnilistURL(input string) string {
	input = strings.TrimSpace(input)
	if strings.HasPrefix(input, "https://anilist.co/user/") || strings.HasPrefix(input, "http://anilist.co/user/") {
		return input
	}

	return fmt.Sprintf("https://anilist.co/user/%s", input)
}

// collectionPageToken represents the opaque cursor for collection pagination
type collectionPageToken struct {
	LastDate time.Time `json:"d"`
	LastID   int64     `json:"id"`
	LastName string    `json:"n,omitempty"`
	OrderBy  string    `json:"o"`
	Search   string    `json:"q,omitempty"`
}

func encodeCollectionPageToken(t collectionPageToken) string {
	data, _ := json.Marshal(t)
	return base64.URLEncoding.EncodeToString(data)
}

func decodeCollectionPageToken(token string) (collectionPageToken, error) {
	data, err := base64.URLEncoding.DecodeString(token)
	if err != nil {
		return collectionPageToken{}, err
	}
	var t collectionPageToken
	if err := json.Unmarshal(data, &t); err != nil {
		return collectionPageToken{}, err
	}
	return t, nil
}

// offset-based pagination token helpers (for backward compatibility)
func encodePageToken(offset int) string {
	data, _ := json.Marshal(map[string]int{"offset": offset})
	return base64.URLEncoding.EncodeToString(data)
}

func decodePageToken(token string) (int, error) {
	data, err := base64.URLEncoding.DecodeString(token)
	if err != nil {
		return 0, err
	}
	var result map[string]int
	if err := json.Unmarshal(data, &result); err != nil {
		return 0, err
	}
	return result["offset"], nil
}
