package anilist

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Khan/genqlient/graphql"

	"github.com/karitham/waifubot/collection"
	"github.com/karitham/waifubot/discord"
)

// getTitle returns the best title to display, preferring romaji over english.
func getTitle(title getMediaCharactersMediaTitle) string {
	if title.GetRomaji() != "" {
		return title.GetRomaji()
	}
	return title.GetEnglish()
}

// getSearchTitle returns the best title to display for search results, preferring romaji over english.
func getSearchTitle(title searchMediaPageMediaTitle) string {
	if title.GetRomaji() != "" {
		return title.GetRomaji()
	}
	return title.GetEnglish()
}

// github.com/Khan/genqlient
//go:generate genqlient genqlient.yaml

// Anilist defines a common interface to interact with anilist
type Anilist struct {
	c graphql.Client
}

// Check that anilist actually implements the interface
var _ discord.TrackingService = (*Anilist)(nil)

// New returns a new anilist client
func New() *Anilist {
	const graphURL = "https://graphql.anilist.co"

	a := &Anilist{
		c: graphql.NewClient(graphURL, &http.Client{Timeout: 5 * time.Second}),
	}

	return a
}

// Anime returns an anime by title
func (a *Anilist) Anime(ctx context.Context, title string) ([]collection.Media, error) {
	return a.media(ctx, title, MediaTypeAnime)
}

func (a *Anilist) media(ctx context.Context, title string, t MediaType) ([]collection.Media, error) {
	media, err := media(ctx, a.c, title, t)
	if err != nil {
		return nil, err
	}
	resp := make([]collection.Media, len(media.Page.Media))
	for i, m := range media.Page.Media {
		resp[i] = collection.Media{
			ID:              m.Id,
			CoverImageURL:   m.CoverImage.Large,
			BannerImageURL:  m.BannerImage,
			CoverImageColor: ColorUint(m.CoverImage.Color),
			Title:           m.Title.Romaji,
			URL:             m.SiteUrl,
			Description:     m.Description,
			Type:            string(t),
		}
	}

	return resp, err
}

// User returns a user by name
func (a *Anilist) User(ctx context.Context, name string) ([]collection.TrackerUser, error) {
	users, err := user(ctx, a.c, name)
	if err != nil {
		return nil, err
	}

	resp := make([]collection.TrackerUser, len(users.Page.Users))
	for i, u := range users.Page.Users {
		resp[i] = collection.TrackerUser{
			URL:      u.SiteUrl,
			Name:     u.Name,
			ImageURL: fmt.Sprintf("https://img.anili.st/user/%d", u.Id),
			About:    u.About,
		}
	}

	return resp, nil
}

// Character returns a character by name
func (a *Anilist) Character(ctx context.Context, name string) ([]collection.MediaCharacter, error) {
	char, err := character(ctx, a.c, name)
	if err != nil {
		return nil, err
	}

	resp := make([]collection.MediaCharacter, len(char.Page.Characters))
	for i, c := range char.Page.Characters {
		resp[i] = collection.MediaCharacter{
			ID:          c.Id,
			Name:        c.Name.Full,
			ImageURL:    c.Image.Large,
			URL:         c.SiteUrl,
			Description: c.Description,
			Favorites:   int(c.Favourites),
		}
	}

	return resp, nil
}

// CharactersByIDs fetches multiple characters by their AniList IDs
func (a *Anilist) CharactersByIDs(ctx context.Context, ids []int64) ([]collection.MediaCharacter, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	if len(ids) > 50 {
		slog.Warn("charactersByIds called with more than 50 IDs, truncating", "count", len(ids))
		ids = ids[:50]
	}

	resp, err := charactersByIds(ctx, a.c, ids)
	if err != nil {
		return nil, err
	}

	result := make([]collection.MediaCharacter, 0, len(resp.Page.Characters))
	for _, c := range resp.Page.Characters {
		mediaTitle := ""
		if len(c.Media.Nodes) > 0 {
			mediaTitle = c.Media.Nodes[0].Title.Romaji
		}

		result = append(result, collection.MediaCharacter{
			ID:         c.Id,
			Name:       c.Name.Full,
			ImageURL:   c.Image.Large,
			Favorites:  int(c.Favourites),
			MediaTitle: mediaTitle,
		})
	}

	return result, nil
}

// MaxCharacterID returns the highest character ID on AniList.
// Returns 0 if the response is empty (no characters exist).
func (a *Anilist) MaxCharacterID(ctx context.Context) (int64, error) {
	resp, err := maxCharacterID(ctx, a.c)
	if err != nil {
		return 0, err
	}

	if len(resp.Page.Characters) == 0 {
		return 0, nil
	}

	return resp.Page.Characters[0].Id, nil
}

// Manga returns a manga by title
func (a *Anilist) Manga(ctx context.Context, title string) ([]collection.Media, error) {
	return a.media(ctx, title, MediaTypeManga)
}

// SearchMedia returns media (anime and manga) by search term
func (a *Anilist) SearchMedia(ctx context.Context, search string) ([]collection.Media, error) {
	resp, err := searchMedia(ctx, a.c, search)
	if err != nil {
		return nil, err
	}
	media := make([]collection.Media, len(resp.Page.Media))
	for i, m := range resp.Page.Media {
		media[i] = collection.Media{
			ID:            m.Id,
			CoverImageURL: m.CoverImage.Large,
			Title:         getSearchTitle(m.Title),
			Type:          string(m.Type),
		}
	}
	return media, nil
}

// GetMediaCharacters returns up to 100 characters from a media, paginating as needed
func (a *Anilist) GetMediaCharacters(ctx context.Context, mediaId int64) ([]collection.MediaCharacter, error) {
	var allCharacters []collection.MediaCharacter
	page := int64(1)
	maxPages := int64(4) // 25 * 4 = 100 characters max

	for page <= maxPages {
		resp, err := getMediaCharacters(ctx, a.c, mediaId, page)
		if err != nil {
			return nil, err
		}

		if resp.Media.Id == 0 {
			return nil, fmt.Errorf("media not found: %d", mediaId)
		}

		// Get media title from first page
		var mediaTitle string
		if page == 1 {
			mediaTitle = getTitle(resp.Media.Title)
		}

		for _, edge := range resp.Media.Characters.Edges {
			allCharacters = append(allCharacters, collection.MediaCharacter{
				ID:         edge.Node.Id,
				Name:       edge.Node.Name.Full,
				ImageURL:   edge.Node.Image.Large,
				Favorites:  int(edge.Node.Favourites),
				MediaTitle: strings.Join(strings.Fields(mediaTitle), " "),
			})
		}

		if !resp.Media.Characters.PageInfo.HasNextPage {
			break
		}
		page++
	}

	return allCharacters, nil
}

// ColorUint
// Turn an hex color string beginning with a # into a uint32 representing a color.
func ColorUint(s string) uint32 {
	s = strings.Trim(s, "#")
	u, _ := strconv.ParseUint(s, 16, 32)
	return uint32(u)
}
