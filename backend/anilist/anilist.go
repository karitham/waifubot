package anilist

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/karitham/waifubot/discord"
	"github.com/Khan/genqlient/graphql"
	"github.com/rs/zerolog/log"
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
func (a *Anilist) RandomChar(ctx context.Context, notIn ...int64) (discord.MediaCharacter, error) {
	if !a.cache {
		return a.randomChar(ctx, notIn...)
	}

	if notIn == nil {
		notIn = []int64{0}
	}

	c := a.internalCache["random"]
	c.Lock()
	defer c.Unlock()

	rest := []discord.MediaCharacter{}

two:
	for id, char := range c.cache {
		for _, n := range notIn {
			if n == id {
				continue two
			}
		}

		rest = append(rest, char.(discord.MediaCharacter))
	}

	if len(c.cache) < 100 {
		for i := 0; i < 5; i++ {
			go func() {
				ch, err := a.randomChar(context.Background(), notIn...)
				if err != nil {
					log.Warn().Err(err).Msg("error getting random char")
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
		log.Trace().Str("char", char.Name).Int("cache size", len(c.cache)).Msg("Hit cache")
		delete(c.cache, char.ID)
		return char, nil
	}

	return a.randomChar(ctx, notIn...)
}

func (a *Anilist) randomChar(ctx context.Context, notIn ...int64) (discord.MediaCharacter, error) {
	r, err := charactersRandom(ctx, a.c, a.seed.Int63()%a.MaxChars, notIn)
	if err != nil {
		return discord.MediaCharacter{}, err
	}

	if len(r.Page.Characters) < 1 {
		return discord.MediaCharacter{}, errors.New("error querying random char")
	}

	c := r.Page.Characters[0]
	mediaTitle := ""
	if len(c.Media.Nodes) > 0 {
		mediaTitle = c.Media.Nodes[0].Title.Romaji
	}

	return discord.MediaCharacter{
		ID:         c.Id,
		Name:       strings.Join(strings.Fields(c.Name.Full), " "),
		ImageURL:   c.Image.Large,
		URL:        c.SiteUrl,
		MediaTitle: strings.Join(strings.Fields(mediaTitle), " "),
	}, nil
}

// Anime returns an anime by title
func (a *Anilist) Anime(ctx context.Context, title string) ([]discord.Media, error) {
	return a.media(ctx, title, MediaTypeAnime)
}

func (a *Anilist) media(ctx context.Context, title string, t MediaType) ([]discord.Media, error) {
	media, err := media(ctx, a.c, title, t)
	if err != nil {
		return nil, err
	}
	resp := make([]discord.Media, len(media.Page.Media))
	for i, m := range media.Page.Media {
		resp[i] = discord.Media{
			CoverImageURL:   m.CoverImage.Large,
			BannerImageURL:  m.BannerImage,
			CoverImageColor: ColorUint(m.CoverImage.Color),
			Title:           m.Title.Romaji,
			URL:             m.SiteUrl,
			Description:     m.Description,
		}
	}

	return resp, err
}

// User returns a user by name
func (a *Anilist) User(ctx context.Context, name string) ([]discord.TrackerUser, error) {
	users, err := user(ctx, a.c, name)
	if err != nil {
		return nil, err
	}

	resp := make([]discord.TrackerUser, len(users.Page.Users))
	for i, u := range users.Page.Users {
		resp[i] = discord.TrackerUser{
			URL:      u.SiteUrl,
			Name:     u.Name,
			ImageURL: fmt.Sprintf("https://img.anili.st/user/%d", u.Id),
			About:    u.About,
		}
	}

	return resp, nil
}

// Character returns a character by name
func (a *Anilist) Character(ctx context.Context, name string) ([]discord.MediaCharacter, error) {
	char, err := character(ctx, a.c, name)
	if err != nil {
		return nil, err
	}

	resp := make([]discord.MediaCharacter, len(char.Page.Characters))
	for i, c := range char.Page.Characters {
		resp[i] = discord.MediaCharacter{
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
func (a *Anilist) Manga(ctx context.Context, title string) ([]discord.Media, error) {
	return a.media(ctx, title, MediaTypeManga)
}

// ColorUint
// Turn an hex color string beginning with a # into a uint32 representing a color.
func ColorUint(s string) uint32 {
	s = strings.Trim(s, "#")
	u, _ := strconv.ParseUint(s, 16, 32)
	return uint32(u)
}

func (a *Anilist) randomCache(c querier) {
	for i := 0; i < 5; i++ {
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
