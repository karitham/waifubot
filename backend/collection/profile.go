package collection

import (
	"context"
	"errors"
	"fmt"
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
	CharacterCount    int
	FavoriteCharacter *catalog.Character
}

// SetFavorite sets the user's favorite character, validating ownership.
// A zero charID clears the favorite.
func SetFavorite(ctx context.Context, store Store, userID UserID, charID catalog.CharacterID) error {
	if charID != 0 {
		if _, err := store.GetOwnedCharacter(ctx, userID, charID); err != nil {
			if errors.Is(err, ErrNotFound) {
				return fmt.Errorf("%w %d", ErrUserDoesNotOwnCharacter, charID)
			}
			return err
		}
	}

	return store.UpdateFavorite(ctx, userID, charID)
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

	var favorite *catalog.Character
	if u.FavoriteCharacterID != 0 {
		fav, err := store.GetCharacterByID(ctx, u.FavoriteCharacterID)
		if err == nil {
			favorite = &fav
		}
	}

	count, err := store.CountCollection(ctx, userID)
	if err != nil {
		return Profile{}, err
	}

	return Profile{
		User:              u,
		FavoriteCharacter: favorite,
		CharacterCount:    int(count),
	}, nil
}
