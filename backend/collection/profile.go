package collection

import (
	"context"
	"errors"
	"time"
)

// AnimeService defines the interface for anime operations.
type AnimeService interface {
	GetMediaCharacters(ctx context.Context, mediaId int64) ([]MediaCharacter, error)
	Anime(ctx context.Context, name string) ([]Media, error)
	Manga(ctx context.Context, name string) ([]Media, error)
	User(ctx context.Context, name string) ([]TrackerUser, error)
	Character(ctx context.Context, name string) ([]MediaCharacter, error)
}

// Config holds configuration values.
type Config struct {
	RollCooldown   time.Duration
	SeriesRollCost int32
}

// RollConfig holds configuration for roll operations.
type RollConfig struct {
	RollCooldown time.Duration
}

// MediaCharacter represents a character from the anime service.
type MediaCharacter struct {
	ID          int64
	Name        string
	ImageURL    string
	URL         string
	Description string
	MediaTitle  string
	Favorites   int
}

// Media represents an anime or manga.
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

// TrackerUser represents an anime tracker user.
type TrackerUser struct {
	Name     string
	URL      string
	ImageURL string
	About    string
}

// Profile represents a complete user profile with character count and favorite.
type Profile struct {
	User
	CharacterCount int
	Favorite       Character
}

// UserProfile retrieves a user's profile.
func UserProfile(ctx context.Context, store Store, userID UserID) (Profile, error) {
	u, err := store.GetUser(ctx, userID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			if err = store.CreateUser(ctx, userID); err != nil {
				return Profile{}, err
			}
			u, err = store.GetUser(ctx, userID)
			if err != nil {
				return Profile{}, err
			}
		} else {
			return Profile{}, err
		}
	}

	var favorite Character
	if u.Favorite != 0 {
		fav, err := store.GetCharacterByID(ctx, u.Favorite)
		if err == nil {
			favorite = fav
		}
	}

	count, err := store.CountCollection(ctx, userID)
	if err != nil {
		return Profile{}, err
	}

	return Profile{
		User:           u,
		Favorite:       favorite,
		CharacterCount: int(count),
	}, nil
}
