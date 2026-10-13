// Package tracker defines the records returned by a media tracker.
// These are provider-shaped DTOs: stored characters use catalog.Character.
// Character IDs are AniList IDs, which are also the catalog keys, so
// Character.ID reuses catalog.CharacterID.
package tracker

import (
	"context"

	"github.com/karitham/waifubot/catalog"
)

// Service is the media tracker.
type Service interface {
	GetMediaCharacters(ctx context.Context, mediaID MediaID) ([]Character, error)
	Anime(ctx context.Context, name string) ([]Media, error)
	Manga(ctx context.Context, name string) ([]Media, error)
	User(ctx context.Context, name string) ([]User, error)
	Character(ctx context.Context, name string) ([]Character, error)
	SearchMedia(ctx context.Context, search string) ([]Media, error)
}

// MediaID identifies an anime or manga on the tracker.
type MediaID int64

// Character is a character record from a tracker.
type Character struct {
	ID          catalog.CharacterID
	Name        string
	ImageURL    string
	URL         string
	Description string
	MediaTitle  string
	Popularity  int
}

// ToCatalog converts the provider record into the local catalog form.
func (c Character) ToCatalog() catalog.Character {
	return catalog.Character{
		ID:         c.ID,
		Name:       c.Name,
		Image:      c.ImageURL,
		MediaTitle: c.MediaTitle,
		Popularity: c.Popularity,
	}
}

// Media is an anime or manga record.
type Media struct {
	ID              MediaID
	Title           string
	URL             string
	CoverImageURL   string
	BannerImageURL  string
	CoverImageColor uint32
	Description     string
	Type            string
}

// User is a tracker user profile.
type User struct {
	Name     string
	URL      string
	ImageURL string
	About    string
}
