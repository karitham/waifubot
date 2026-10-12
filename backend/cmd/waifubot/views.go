package main

import (
	"time"

	"github.com/karitham/waifubot/catalog"
	"github.com/karitham/waifubot/collection"
	"github.com/karitham/waifubot/tracker"
	"github.com/karitham/waifubot/wishlist"
)

// CLI JSON views. Commands encode these instead of domain types so that
// storage bookkeeping (cursor timestamps, active flags) and Go field
// names never leak into the output.

type cliCharacter struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	Image      string `json:"image"`
	MediaTitle string `json:"media_title"`
	Favorites  int    `json:"favorites"`
}

func characterView(c catalog.Character) cliCharacter {
	return cliCharacter{
		ID:         int64(c.ID),
		Name:       c.Name,
		Image:      c.Image,
		MediaTitle: c.MediaTitle,
		Favorites:  c.Popularity,
	}
}

type cliOwnedCharacter struct {
	cliCharacter
	Date   time.Time         `json:"date"`
	Source collection.Source `json:"source"`
	UserID uint64            `json:"user_id"`
}

func ownedCharacterView(c collection.OwnedCharacter) cliOwnedCharacter {
	return cliOwnedCharacter{
		cliCharacter: characterView(c.Character),
		Date:         c.Date,
		Source:       c.Source,
		UserID:       uint64(c.UserID),
	}
}

type cliWishlistEntry struct {
	cliCharacter
	AddedAt time.Time `json:"added_at"`
}

func wishlistEntryView(e wishlist.Entry) cliWishlistEntry {
	return cliWishlistEntry{
		cliCharacter: characterView(e.Character),
		AddedAt:      e.AddedAt,
	}
}

type cliUserCharacterSet struct {
	UserID     uint64             `json:"user_id"`
	Characters []cliWishlistEntry `json:"characters"`
}

func userCharacterSetView(s wishlist.UserCharacterSet) cliUserCharacterSet {
	chars := make([]cliWishlistEntry, len(s.Characters))
	for i, c := range s.Characters {
		chars[i] = wishlistEntryView(c)
	}
	return cliUserCharacterSet{
		UserID:     uint64(s.UserID),
		Characters: chars,
	}
}

type cliProfile struct {
	ID                uint64        `json:"id"`
	Quote             string        `json:"quote"`
	Tokens            int32         `json:"tokens"`
	AnilistURL        string        `json:"anilist_url"`
	DiscordUsername   string        `json:"discord_username"`
	DiscordAvatar     string        `json:"discord_avatar"`
	CharacterCount    int           `json:"character_count"`
	FavoriteCharacter *cliCharacter `json:"favorite_character,omitempty"`
}

func profileView(p collection.Profile) cliProfile {
	var favorite *cliCharacter
	if p.FavoriteCharacter != nil {
		v := characterView(*p.FavoriteCharacter)
		favorite = &v
	}
	return cliProfile{
		ID:                uint64(p.UserID),
		Quote:             p.Quote,
		Tokens:            p.Tokens,
		AnilistURL:        p.AnilistURL,
		DiscordUsername:   p.DiscordUsername,
		DiscordAvatar:     p.DiscordAvatar,
		CharacterCount:    p.CharacterCount,
		FavoriteCharacter: favorite,
	}
}

type cliMedia struct {
	ID              int64  `json:"id"`
	Title           string `json:"title"`
	URL             string `json:"url"`
	CoverImageURL   string `json:"cover_image_url"`
	BannerImageURL  string `json:"banner_image_url"`
	CoverImageColor uint32 `json:"cover_image_color"`
	Description     string `json:"description"`
	Type            string `json:"type"`
}

func mediaView(m tracker.Media) cliMedia {
	return cliMedia{
		ID:              int64(m.ID),
		Title:           m.Title,
		URL:             m.URL,
		CoverImageURL:   m.CoverImageURL,
		BannerImageURL:  m.BannerImageURL,
		CoverImageColor: m.CoverImageColor,
		Description:     m.Description,
		Type:            m.Type,
	}
}

func mediaViews(media []tracker.Media) []cliMedia {
	views := make([]cliMedia, len(media))
	for i, m := range media {
		views[i] = mediaView(m)
	}
	return views
}
