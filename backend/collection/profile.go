package collection

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/karitham/waifubot/catalog"
)

// Config holds configuration values.
type Config struct {
	RollCooldown   time.Duration
	SeriesRollCost TokenAmount
}

// Profile represents a complete user profile with character count and favorite.
type Profile struct {
	User
	CharacterCount    int
	FavoriteCharacter *catalog.Character
}

// UserProfile retrieves a user's profile. A user with no row has an empty
// profile; reads never create the user.
func UserProfile(ctx context.Context, store Store, userID UserID) (Profile, error) {
	u, err := store.GetUser(ctx, userID)
	if errors.Is(err, ErrNotFound) {
		u = User{UserID: userID}
	} else if err != nil {
		return Profile{}, err
	}

	var favorite *catalog.Character
	if u.FavoriteCharacterID != 0 {
		fav, err := store.GetCharacterByID(ctx, u.FavoriteCharacterID)
		if err == nil {
			favorite = &fav
		} else if !errors.Is(err, ErrNotFound) {
			return Profile{}, fmt.Errorf("failed to get favorite character: %w", err)
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
