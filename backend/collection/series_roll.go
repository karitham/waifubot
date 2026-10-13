package collection

import (
	"context"
	"math/rand/v2"
	"time"

	"github.com/karitham/waifubot/catalog"
	"github.com/karitham/waifubot/tracker"
)

// SeriesRoll executes a paid, series-specific roll for a user, deducting tokens.
func SeriesRoll(ctx context.Context, store DB, anime AnimeService, userID UserID, mediaID tracker.MediaID, config Config) (catalog.Character, error) {
	return acquire(ctx, store, userID, acquisition{
		source: SourceSeriesRoll,
		check: func(user User, _ time.Time) error {
			if user.Tokens < config.SeriesRollCost.amount {
				return ErrInsufficientTokens
			}
			return nil
		},
		pick: func(ctx context.Context, store Store, userID UserID) (catalog.Character, error) {
			return pickSeriesCharacter(ctx, store, userID, mediaID, anime)
		},
		charge: func(ctx context.Context, tx Tx, now time.Time) error {
			_, err := tx.SpendTokens(ctx, userID, config.SeriesRollCost)
			return err
		},
	})
}

// pickSeriesCharacter selects a random unowned character from a series among
// the characters the catalog knows.
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

	// Characters enter the catalog only through the hydrator; a tracker
	// character it has not reached yet is not ownable.
	ids := make([]catalog.CharacterID, len(unowned))
	for i, c := range unowned {
		ids[i] = c.ID
	}

	known, err := store.KnownCharacterIDs(ctx, ids)
	if err != nil {
		return catalog.Character{}, err
	}

	knownSet := make(map[catalog.CharacterID]struct{}, len(known))
	for _, id := range known {
		knownSet[id] = struct{}{}
	}

	var pickable []tracker.Character
	for _, c := range unowned {
		if _, ok := knownSet[c.ID]; ok {
			pickable = append(pickable, c)
		}
	}

	if len(pickable) == 0 {
		return catalog.Character{}, ErrCharactersNotSynced
	}

	return pickable[rand.IntN(len(pickable))].ToCatalog(), nil
}
