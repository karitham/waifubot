package collection

import (
	"context"
	"errors"
	"math/rand/v2"
	"time"
)

// SeriesRoll executes a paid, series-specific roll for a user, deducting tokens.
func (s *RollService) SeriesRoll(ctx context.Context, userID UserID, mediaID int64, seriesRollCost int32, anime AnimeService) (MediaCharacter, error) {
	// --- GATHER ---
	user, err := s.store.GetUser(ctx, userID)
	if err != nil {
		if !errors.Is(err, ErrNotFound) {
			return MediaCharacter{}, err
		}
		if err := s.store.CreateUser(ctx, userID); err != nil {
			return MediaCharacter{}, err
		}
		user, err = s.store.GetUser(ctx, userID)
		if err != nil {
			return MediaCharacter{}, err
		}
	}

	if user.Tokens < seriesRollCost {
		return MediaCharacter{}, ErrInsufficientTokens
	}

	ownedIDs, err := s.store.GetCollectionIDs(ctx, userID)
	if err != nil {
		return MediaCharacter{}, err
	}

	allChars, err := anime.GetMediaCharacters(ctx, mediaID)
	if err != nil {
		return MediaCharacter{}, err
	}

	if len(allChars) == 0 {
		return MediaCharacter{}, ErrMediaNotFound
	}

	// --- PROCESS ---
	owned := make(map[int64]struct{}, len(ownedIDs))
	for _, id := range ownedIDs {
		owned[id] = struct{}{}
	}

	var unowned []MediaCharacter
	for _, c := range allChars {
		if _, ok := owned[c.ID]; !ok {
			unowned = append(unowned, c)
		}
	}

	if len(unowned) == 0 {
		return MediaCharacter{}, ErrNoUnownedCharacters
	}

	char := unowned[rand.IntN(len(unowned))]

	if err := s.store.UpsertCharacter(ctx, Character{
		ID:        char.ID,
		Name:      char.Name,
		Image:     char.ImageURL,
		Favorites: char.Favorites,
	}); err != nil {
		return MediaCharacter{}, err
	}

	// --- COMMIT ---
	now := time.Now()
	err = withTx(ctx, s.store, func(tx Store) error {
		// Re-fetch user and re-check tokens to prevent double-spend
		user, err := tx.GetUser(ctx, userID)
		if err != nil {
			return err
		}
		if user.Tokens < seriesRollCost {
			return ErrInsufficientTokens
		}

		if err := tx.AddToCollection(ctx, userID, Character{
			ID:         char.ID,
			Name:       char.Name,
			Image:      char.ImageURL,
			MediaTitle: char.MediaTitle,
		}, "SERIES_ROLL", now); err != nil {
			return err
		}

		if err := tx.RemoveFromWishlist(ctx, userID, char.ID); err != nil {
			return err
		}

		if _, err := tx.SpendTokens(ctx, userID, seriesRollCost); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return MediaCharacter{}, err
	}
	return char, nil
}
