package collection

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// ErrRollCooldown is returned when a user is on roll cooldown.
type ErrRollCooldown struct {
	Until time.Time
}

func (e ErrRollCooldown) Error() string {
	return fmt.Sprintf("You can roll again in %s.", time.Until(e.Until).Round(time.Second))
}

// withTx runs fn inside a transaction, committing on success or rolling back on failure.
func withTx(ctx context.Context, store Store, fn func(tx Store) error) error {
	tx, err := store.WithTx(ctx)
	if err != nil {
		return err
	}
	if err := fn(tx); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	return nil
}

// RollService orchestrates roll operations.
type RollService struct {
	store  Store
	config RollConfig
}

func NewRollService(store Store, config RollConfig) *RollService {
	return &RollService{store: store, config: config}
}

// Roll executes the free roll for a user, enforcing the cooldown constraint.
func (s *RollService) Roll(ctx context.Context, userID UserID) (MediaCharacter, error) {
	// --- GATHER ---
	user, err := s.store.GetUser(ctx, userID)
	if err != nil {
		if !errors.Is(err, ErrNotFound) {
			return MediaCharacter{}, err
		}
		// User is new — skip cooldown check
	} else {
		now := time.Now()
		cooldownUntil := user.Date.Add(s.config.RollCooldown)
		if now.Before(cooldownUntil) {
			return MediaCharacter{}, ErrRollCooldown{Until: cooldownUntil}
		}
	}

	// --- PROCESS ---
	catChar, err := s.store.RandomCharNotOwned(ctx, userID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return MediaCharacter{}, ErrNoUnownedCharacters
		}
		return MediaCharacter{}, err
	}

	char := MediaCharacter{
		ID:         catChar.ID,
		Name:       catChar.Name,
		ImageURL:   catChar.Image,
		MediaTitle: catChar.MediaTitle,
		Favorites:  catChar.Favorites,
	}

	// --- COMMIT ---
	now := time.Now()
	err = withTx(ctx, s.store, func(tx Store) error {
		user, err := tx.GetUser(ctx, userID)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				if err = tx.CreateUser(ctx, userID); err != nil {
					return err
				}
			} else {
				return err
			}
		} else {
			// Re-check cooldown inside the transaction — a concurrent request
			// may have already updated last_roll since our earlier check.
			cooldownUntil := user.Date.Add(s.config.RollCooldown)
			if now.Before(cooldownUntil) {
				return ErrRollCooldown{Until: cooldownUntil}
			}
		}

		if err := tx.UpsertCharacter(ctx, Character{
			ID:         char.ID,
			Name:       char.Name,
			Image:      char.ImageURL,
			MediaTitle: char.MediaTitle,
			Favorites:  char.Favorites,
		}); err != nil {
			return err
		}

		if err := tx.AddToCollection(ctx, userID, Character{
			ID:         char.ID,
			Name:       char.Name,
			Image:      char.ImageURL,
			MediaTitle: char.MediaTitle,
		}, "ROLL", now); err != nil {
			return err
		}

		if err := tx.RemoveFromWishlist(ctx, userID, char.ID); err != nil {
			return err
		}

		return tx.UpdateLastRoll(ctx, userID, now)
	})
	if err != nil {
		return MediaCharacter{}, err
	}
	return char, nil
}
