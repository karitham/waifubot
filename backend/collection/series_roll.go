package collection

import (
	"context"
	"math/rand/v2"
	"time"

	"github.com/karitham/waifubot/catalog"
	"github.com/karitham/waifubot/tracker"
)

// SeriesRoll executes a paid, series-specific roll for a user, deducting tokens.
func (s *RollService) SeriesRoll(ctx context.Context, userID UserID, mediaID tracker.MediaID, seriesRollCost TokenAmount, anime AnimeService) (catalog.Character, error) {
	return acquire(ctx, s.store, userID, acquisition{
		source: SourceSeriesRoll,
		check: func(user User, _ time.Time) error {
			if user.Tokens < seriesRollCost.amount {
				return ErrInsufficientTokens
			}
			return nil
		},
		pick: func(ctx context.Context, store Store, userID UserID) (catalog.Character, error) {
			return pickSeriesCharacter(ctx, store, userID, mediaID, anime)
		},
		charge: func(ctx context.Context, tx Tx, now time.Time) error {
			_, err := tx.SpendTokens(ctx, userID, seriesRollCost)
			return err
		},
	})
}

// pickSeriesCharacter selects a random unowned character from a series,
// upserting it into the catalog.
func pickSeriesCharacter(ctx context.Context, store Store, userID UserID, mediaID tracker.MediaID, anime AnimeService) (catalog.Character, error) {
	ownedIDs, err := store.GetCollectionIDs(ctx, userID)
	if err != nil {
		return catalog.Character{}, err
	}

	allChars, err := anime.GetMediaCharacters(ctx, mediaID)
	if err != nil {
		return catalog.Character{}, err
	}

	if len(allChars) == 0 {
		return catalog.Character{}, ErrMediaNotFound
	}

	owned := make(map[catalog.CharacterID]struct{}, len(ownedIDs))
	for _, id := range ownedIDs {
		owned[id] = struct{}{}
	}

	var unowned []tracker.Character
	for _, c := range allChars {
		if _, ok := owned[c.ID]; !ok {
			unowned = append(unowned, c)
		}
	}

	if len(unowned) == 0 {
		return catalog.Character{}, ErrNoUnownedCharacters
	}

	char := unowned[rand.IntN(len(unowned))].ToCatalog()

	if err := store.UpsertCharacter(ctx, char); err != nil {
		return catalog.Character{}, err
	}

	return char, nil
}
