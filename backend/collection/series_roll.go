package collection

import (
	"context"
	"errors"
	"math/rand"
	"time"
)

// SeriesRoll executes a series-specific roll for a user.
// It charges config.SeriesRollCost tokens, picks a random character from the
// given media (excluding owned), and adds it to the collection.
func SeriesRoll(
	ctx context.Context,
	store Store,
	animeService AnimeService,
	config Config,
	userID UserID,
	mediaID int64,
) (MediaCharacter, error) {
	tx, err := store.WithTx(ctx)
	if err != nil {
		return MediaCharacter{}, err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback(ctx)
		}
	}()

	user, err := tx.GetUser(ctx, userID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			if err = tx.CreateUser(ctx, userID); err != nil {
				return MediaCharacter{}, err
			}
			user, err = tx.GetUser(ctx, userID)
			if err != nil {
				return MediaCharacter{}, err
			}
		} else {
			return MediaCharacter{}, err
		}
	}

	if user.Tokens < config.SeriesRollCost {
		return MediaCharacter{}, ErrInsufficientTokens
	}

	ownedIDs, err := tx.GetCollectionIDs(ctx, userID)
	if err != nil {
		return MediaCharacter{}, err
	}

	allChars, err := animeService.GetMediaCharacters(ctx, mediaID)
	if err != nil {
		return MediaCharacter{}, err
	}

	if len(allChars) == 0 {
		return MediaCharacter{}, ErrMediaNotFound
	}

	// Filter out owned characters
	owned := make(map[int64]bool, len(ownedIDs))
	for _, id := range ownedIDs {
		owned[id] = true
	}

	var unowned []MediaCharacter
	for _, c := range allChars {
		if !owned[c.ID] {
			unowned = append(unowned, c)
		}
	}

	if len(unowned) == 0 {
		return MediaCharacter{}, ErrNoUnownedCharacters
	}

	char := unowned[rand.Intn(len(unowned))]

	now := time.Now()

	err = tx.UpsertCharacter(ctx, Character{
		ID:        char.ID,
		Name:      char.Name,
		Image:     char.ImageURL,
		Favorites: char.Favorites,
	})
	if err != nil {
		return MediaCharacter{}, err
	}

	err = tx.AddToCollection(ctx, userID, Character{
		ID:         char.ID,
		Name:       char.Name,
		Image:      char.ImageURL,
		MediaTitle: char.MediaTitle,
	}, "SERIES_ROLL", now)
	if err != nil {
		return MediaCharacter{}, err
	}

	_ = tx.RemoveFromWishlist(ctx, userID, char.ID)

	_, err = tx.SpendTokens(ctx, userID, config.SeriesRollCost)
	if err != nil {
		return MediaCharacter{}, err
	}

	err = tx.Commit(ctx)
	committed = err == nil
	return char, err
}
