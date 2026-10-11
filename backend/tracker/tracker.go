// Package tracker defines the records returned by a media tracker.
// These are provider-shaped DTOs: stored characters use catalog.Character.
package tracker

import "github.com/karitham/waifubot/catalog"

// Character is a character record from a tracker.
type Character struct {
	ID          int64
	Name        string
	ImageURL    string
	URL         string
	Description string
	MediaTitle  string
	Favorites   int
}

// ToCatalog converts the provider record into the local catalog form.
func (c Character) ToCatalog() catalog.Character {
	return catalog.Character{
		ID:         c.ID,
		Name:       c.Name,
		Image:      c.ImageURL,
		MediaTitle: c.MediaTitle,
		Favorites:  c.Favorites,
	}
}

// Media is an anime or manga record.
type Media struct {
	ID              int64
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
