//go:generate ogen ../../openapi.yaml

package rest

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/Karitham/corde"

	"github.com/karitham/waifubot/collection"
	"github.com/karitham/waifubot/discord"
	"github.com/karitham/waifubot/services"
	"github.com/karitham/waifubot/wishlist"

	"github.com/karitham/waifubot/rest/api"
)

type Server struct {
	db             collection.Store
	wishlistStore  wishlist.Store
	discordService *services.DiscordService
}

func New(db collection.Store, ws wishlist.Store, discordService *services.DiscordService) *Server {
	return &Server{
		db:             db,
		wishlistStore:  ws,
		discordService: discordService,
	}
}

func parseUserID(userID string) (uint64, error) {
	id, err := strconv.ParseUint(userID, 10, 64)
	if err != nil || id == 0 {
		return 0, fmt.Errorf("invalid user id")
	}
	return id, nil
}

func (s *Server) GetUser(ctx context.Context, params api.GetUserParams) (api.GetUserRes, error) {
	id, err := parseUserID(params.UserID)
	if err != nil {
		return &api.GetUserBadRequest{
			Message:    "invalid id provided",
			ErrorCode:  "invalid_id",
			StatusCode: 400,
		}, nil
	}

	return s.getUserProfile(ctx, id)
}

func (s *Server) GetUserV1(ctx context.Context, params api.GetUserV1Params) (api.GetUserV1Res, error) {
	id, err := parseUserID(string(params.UserID))
	if err != nil {
		return &api.GetUserV1BadRequest{
			Message:    "invalid id provided",
			ErrorCode:  "invalid_id",
			StatusCode: 400,
		}, nil
	}

	profile, err := s.getUserProfileData(ctx, id)
	if err != nil {
		return &api.GetUserV1NotFound{
			Message:    "user not found",
			ErrorCode:  "user_not_found",
			StatusCode: 404,
		}, nil
	}
	return profile, nil
}

func (s *Server) GetProfileV1(ctx context.Context, params api.GetProfileV1Params) (api.GetProfileV1Res, error) {
	id, err := parseUserID(string(params.UserID))
	if err != nil {
		return &api.GetProfileV1BadRequest{
			Message:    "invalid id provided",
			ErrorCode:  "invalid_id",
			StatusCode: 400,
		}, nil
	}

	if s.discordService != nil {
		ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		if err := s.discordService.UpdateIfNeeded(ctx, s.db, corde.Snowflake(id)); err != nil {
			slog.With("err", err).Warn("failed to update profile data")
		}
	}

	u, err := s.db.GetUser(ctx, id)
	if err != nil {
		return &api.GetProfileV1NotFound{
			Message:    "user not found",
			ErrorCode:  "user_not_found",
			StatusCode: 404,
		}, nil
	}

	var fav api.OptCharacter
	if u.Favorite != 0 {
		favChar, err := s.db.GetCharacterByID(ctx, u.Favorite)
		if err == nil {
			fav = api.NewOptCharacter(api.Character{
				ID:        favChar.ID,
				Name:      favChar.Name,
				Image:     favChar.Image,
				Favorites: favChar.Favorites,
				// date and type are null — character comes from catalog, not collection
			})
		}
	}

	return &api.UserProfile{
		ID:              fmt.Sprintf("%d", u.UserID),
		Quote:           api.NewOptString(u.Quote),
		Tokens:          u.Tokens,
		AnilistURL:      api.NewOptString(u.AnilistURL),
		DiscordUsername: u.DiscordUsername,
		DiscordAvatar:   api.NewOptString(discord.DiscordAvatarURL(u.UserID, u.DiscordAvatar)),
		Favorite:        fav,
	}, nil
}

func (s *Server) GetCollectionV1(ctx context.Context, params api.GetCollectionV1Params) (api.GetCollectionV1Res, error) {
	id, err := parseUserID(string(params.UserID))
	if err != nil {
		return &api.GetCollectionV1BadRequest{
			Message:    "invalid id provided",
			ErrorCode:  "invalid_id",
			StatusCode: 400,
		}, nil
	}

	chars, err := s.db.GetCollection(ctx, id)
	if err != nil {
		return &api.GetCollectionV1NotFound{
			Message:    "user not found",
			ErrorCode:  "user_not_found",
			StatusCode: 404,
		}, nil
	}

	characters := make([]api.Character, len(chars))
	for i, entry := range chars {
		characters[i] = mapCharacter(entry.ID, entry.Name, entry.Image, entry.Favorites, entry.Source, entry.Date)
	}

	return &api.CollectionResponse{
		Characters: characters,
		Total:      len(characters),
	}, nil
}

func (s *Server) getUserProfile(ctx context.Context, id uint64) (api.GetUserRes, error) {
	profile, err := s.getUserProfileData(ctx, id)
	if err != nil {
		return &api.GetUserNotFound{
			Message:    "user not found",
			ErrorCode:  "user_not_found",
			StatusCode: 404,
		}, nil
	}
	return profile, nil
}

func (s *Server) getUserProfileData(ctx context.Context, id uint64) (*api.Profile, error) {
	if s.discordService != nil {
		ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		if err := s.discordService.UpdateIfNeeded(ctx, s.db, corde.Snowflake(id)); err != nil {
			slog.With("err", err).Warn("failed to update profile data")
		}
	}

	u, err := s.db.GetUser(ctx, id)
	if err != nil {
		return nil, err
	}

	chars, err := s.db.GetCollection(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.mapUser(u, chars), nil
}

func (s *Server) FindUser(ctx context.Context, params api.FindUserParams) (api.FindUserRes, error) {
	anilistVal, anilistSet := params.Anilist.Get()
	discordVal, discordSet := params.Discord.Get()

	if !anilistSet && !discordSet {
		return &api.FindUserBadRequest{
			Message:    "anilist or discord query param is required",
			ErrorCode:  "missing_query_param",
			StatusCode: 400,
		}, nil
	}

	resp, err := s.findUserByQuery(ctx, anilistVal, discordVal, anilistSet)

	if err != nil {
		if errors.Is(err, errUserNotFound) {
			return &api.FindUserNotFound{
				Message:    "user not found",
				ErrorCode:  "user_not_found",
				StatusCode: 404,
			}, nil
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	return resp, nil
}

func (s *Server) FindUserV1(ctx context.Context, params api.FindUserV1Params) (api.FindUserV1Res, error) {
	anilistVal, anilistSet := params.Anilist.Get()
	discordVal, discordSet := params.Discord.Get()

	if !anilistSet && !discordSet {
		return &api.FindUserV1BadRequest{
			Message:    "anilist or discord query param is required",
			ErrorCode:  "missing_query_param",
			StatusCode: 400,
		}, nil
	}

	resp, err := s.findUserByQuery(ctx, anilistVal, discordVal, anilistSet)

	if err != nil {
		if errors.Is(err, errUserNotFound) {
			return &api.FindUserV1NotFound{
				Message:    "user not found",
				ErrorCode:  "user_not_found",
				StatusCode: 404,
			}, nil
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	return resp, nil
}

var (
	errUserNotFound = errors.New("user not found")
)

func (s *Server) findUserByQuery(ctx context.Context, anilist, discord string, useAnilist bool) (*api.UserIdResponse, error) {
	var user collection.User
	var err error

	if useAnilist {
		user, err = s.db.GetUserByAnilist(ctx, normalizeAnilistURL(anilist))
	} else {
		user, err = s.db.GetUserByDiscordUsername(ctx, discord)
	}

	if err != nil || user.UserID == 0 {
		return nil, errUserNotFound
	}

	return &api.UserIdResponse{
		ID: fmt.Sprintf("%d", user.UserID),
	}, nil
}

func (s *Server) GetWishlist(ctx context.Context, params api.GetWishlistParams) (api.GetWishlistRes, error) {
	id, err := strconv.ParseUint(string(params.UserID), 10, 64)
	if err != nil || id == 0 {
		return &api.GetWishlistBadRequest{
			Message:    "invalid id provided",
			ErrorCode:  "invalid_id",
			StatusCode: 400,
		}, nil
	}

	chars, err := s.wishlistStore.GetUserCharacterWishlist(ctx, id)
	if err != nil {
		return &api.GetWishlistNotFound{
			Message:    "user not found",
			ErrorCode:  "user_not_found",
			StatusCode: 404,
		}, nil
	}

	characters := make([]api.Character, len(chars))
	for i, c := range chars {
		t, _ := time.Parse(time.RFC3339, c.Date)
		characters[i] = api.Character{
			Date:      api.NewOptNilDateTime(t),
			Name:      c.Name,
			Image:     c.Image,
			ID:        c.ID,
			Favorites: c.Favorites,
		}
	}

	return &api.WishlistResponse{
		Characters: characters,
		Total:      len(characters),
	}, nil
}

// mapCharacter builds an api.Character from common fields.
func mapCharacter(id int64, name, image string, favorites int, source string, date time.Time) api.Character {
	return api.Character{
		ID:        id,
		Name:      name,
		Image:     image,
		Favorites: favorites,
		Type:      api.NewOptNilCharacterType(api.CharacterType(source)),
		Date:      api.NewOptNilDateTime(date),
	}
}

func (s *Server) mapUser(u collection.User, list []collection.OwnedCharacter) *api.Profile {
	waifus := make([]api.Character, 0, len(list))
	var fav api.OptCharacter
	for _, entry := range list {
		c := mapCharacter(entry.ID, entry.Name, entry.Image, entry.Favorites, entry.Source, entry.Date)

		if u.Favorite != 0 && entry.ID == u.Favorite {
			fav = api.NewOptCharacter(c)
		}

		waifus = append(waifus, c)
	}

	return &api.Profile{
		ID:              fmt.Sprintf("%d", u.UserID),
		Quote:           api.NewOptString(u.Quote),
		Tokens:          u.Tokens,
		Favorite:        fav,
		AnilistURL:      api.NewOptString(u.AnilistURL),
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
