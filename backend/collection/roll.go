package collection

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// ErrRollCooldown is returned when a user is on roll cooldown.
type ErrRollCooldown struct {
	Until         time.Time
	MissingTokens int
}

func (e ErrRollCooldown) Error() string {
	return fmt.Sprintf("You need another %d tokens to roll, or you can wait %s until next free roll.", e.MissingTokens, time.Until(e.Until).Round(time.Second))
}

// Roll executes the roll logic for a user.
func Roll(ctx context.Context, store Store, animeService AnimeService, config Config, userID UserID) (MediaCharacter, error) {
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

	now := time.Now()
	canRollFree := now.After(user.Date.Add(config.RollCooldown))
	hasTokens := user.Tokens >= config.TokensNeeded

	if !canRollFree && !hasTokens {
		return MediaCharacter{}, ErrRollCooldown{
			Until:         user.Date.Add(config.RollCooldown),
			MissingTokens: int(config.TokensNeeded - user.Tokens),
		}
	}

	charsIDs, err := tx.GetCollectionIDs(ctx, userID)
	if err != nil {
		return MediaCharacter{}, err
	}

	char, err := animeService.RandomChar(ctx, charsIDs...)
	if err != nil {
		return MediaCharacter{}, err
	}

	err = tx.UpsertCharacter(ctx, Character{
		ID:    char.ID,
		Name:  char.Name,
		Image: char.ImageURL,
	})
	if err != nil {
		return MediaCharacter{}, err
	}

	err = tx.AddToCollection(ctx, userID, Character{
		ID:         char.ID,
		Name:       char.Name,
		Image:      char.ImageURL,
		MediaTitle: char.MediaTitle,
	}, "ROLL", now)
	if err != nil {
		return MediaCharacter{}, err
	}

	_ = tx.RemoveFromWishlist(ctx, userID, char.ID)

	if canRollFree {
		err = tx.UpdateLastRoll(ctx, userID, now)
	} else {
		_, err = tx.SpendTokens(ctx, userID, config.TokensNeeded)
	}

	if err != nil {
		return MediaCharacter{}, err
	}

	err = tx.Commit(ctx)
	committed = err == nil
	return char, err
}
