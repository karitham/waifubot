package anilist

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Khan/genqlient/graphql"

	"github.com/karitham/waifubot/collection"
	"github.com/karitham/waifubot/discord"
)

// github.com/Khan/genqlient
//go:generate genqlient genqlient.yaml

// Anilist defines a common interface to interact with anilist
type Anilist struct {
	c             graphql.Client
	seed          rand.Source64
	MaxChars      int64
	internalCache map[string]querier
	cache         bool
}

// Check that anilist actually implements the interface
var _ discord.TrackingService = (*Anilist)(nil)

type querier struct {
	*sync.Mutex
	cache map[any]any
}

// New returns a new anilist client
func New(opts ...func(*Anilist)) *Anilist {
	const graphURL = "https://graphql.anilist.co"

	a := &Anilist{
		c:        graphql.NewClient(graphURL, &http.Client{Timeout: 5 * time.Second}),
		MaxChars: 50_000,
		seed:     rand.New(rand.NewSource(time.Now().Unix())),
		internalCache: map[string]querier{
			"random": {
				cache: make(map[any]any),
				Mutex: &sync.Mutex{},
			},
		},
		cache: true,
	}

	for _, opt := range opts {
		opt(a)
	}

	if a.cache {
		go a.randomCache(a.internalCache["random"])
	}

	return a
}

// NoCache disables the cache
func NoCache(a *Anilist) {
	a.cache = false
}

// MaxChar sets the maximum number of characters to return
func MaxChar(n int64) func(*Anilist) {
	return func(a *Anilist) {
		a.MaxChars = n
	}
}

// RandomChar returns a random char
func (a *Anilist) RandomChar(ctx context.Context, notIn ...int64) (collection.MediaCharacter, error) {
	if !a.cache {
		return a.randomChar(ctx, notIn...)
	}

	if notIn == nil {
		notIn = []int64{0}
	}

	c := a.internalCache["random"]
	c.Lock()
	defer c.Unlock()

	rest := []collection.MediaCharacter{}

two:
	for id, char := range c.cache {
		for _, n := range notIn {
			if n == id {
				continue two
			}
		}

		rest = append(rest, char.(collection.MediaCharacter))
	}

	if len(c.cache) < 100 {
		for range 5 {
			go func() {
				ch, err := a.randomChar(context.Background(), notIn...)
				if err != nil {
					slog.Warn("error getting random char", "error", err)
					return
				}
				c.Lock()
				defer c.Unlock()
				c.cache[ch.ID] = ch
			}()
		}
	}

	if len(rest) > 0 {
		char := rest[a.seed.Int63()%(int64(len(rest)))]
		slog.Debug("Hit cache", "char", char.Name, "cache_size", len(c.cache))
		delete(c.cache, char.ID)
		return char, nil
	}

	return a.randomChar(ctx, notIn...)
}

func (a *Anilist) randomChar(ctx context.Context, notIn ...int64) (collection.MediaCharacter, error) {
	r, err := charactersRandom(ctx, a.c, a.seed.Int63()%a.MaxChars, notIn)
	if err != nil {
		return collection.MediaCharacter{}, err
	}

	if len(r.Page.Characters) < 1 {
		return collection.MediaCharacter{}, errors.New("error querying random char")
	}

	c := r.Page.Characters[0]
	mediaTitle := ""
	if len(c.Media.Nodes) > 0 {
		mediaTitle = c.Media.Nodes[0].Title.Romaji
	}

	return collection.MediaCharacter{
		ID:         c.Id,
		Name:       strings.Join(strings.Fields(c.Name.Full), " "),
		ImageURL:   c.Image.Large,
		URL:        c.SiteUrl,
		MediaTitle: strings.Join(strings.Fields(mediaTitle), " "),
	}, nil
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
		}
	}

	return resp, nil
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
			Title:         m.Title.Romaji,
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

		for _, edge := range resp.Media.Characters.Edges {
			allCharacters = append(allCharacters, collection.MediaCharacter{
				ID:       edge.Node.Id,
				Name:     edge.Node.Name.Full,
				ImageURL: edge.Node.Image.Large,
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

func (a *Anilist) randomCache(c querier) {
	for range 5 {
		time.Sleep(500 * time.Millisecond)
		go func() {
			ch, err := a.randomChar(context.Background())
			if err != nil {
				return
			}
			c.Lock()
			defer c.Unlock()
			c.cache[ch.ID] = ch
		}()
	}
}
