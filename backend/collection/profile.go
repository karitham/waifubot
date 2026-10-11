package collection

import (
	"context"
	"errors"
	"time"

	"github.com/karitham/waifubot/catalog"
	"github.com/karitham/waifubot/tracker"
)

// AnimeService defines the interface for anime operations.
type AnimeService interface {
	GetMediaCharacters(ctx context.Context, mediaID tracker.MediaID) ([]tracker.Character, error)
	Anime(ctx context.Context, name string) ([]tracker.Media, error)
	Manga(ctx context.Context, name string) ([]tracker.Media, error)
	User(ctx context.Context, name string) ([]tracker.User, error)
	Character(ctx context.Context, name string) ([]tracker.Character, error)
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

// Profile represents a complete user profile with character count and favorite.
type Profile struct {
	User
	CharacterCount int
	Favorite       catalog.Character
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

	var favorite catalog.Character
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
