package collection

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/karitham/waifubot/catalog"
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
func (s *RollService) Roll(ctx context.Context, userID UserID) (catalog.Character, error) {
	return acquire(ctx, s.store, userID, acquisition{
		source: SourceRoll,
		check: func(user User, now time.Time) error {
			cooldownUntil := user.Date.Add(s.config.RollCooldown)
			if now.Before(cooldownUntil) {
				return ErrRollCooldown{Until: cooldownUntil}
			}
			return nil
		},
		pick: func(ctx context.Context, store Store, userID UserID) (catalog.Character, error) {
			char, err := store.RandomCharNotOwned(ctx, userID, RollWeightExponent)
			if errors.Is(err, ErrNotFound) {
				return catalog.Character{}, ErrNoUnownedCharacters
			}
			return char, err
		},
		charge: func(ctx context.Context, tx Store, now time.Time) error {
			return tx.RecordRoll(ctx, userID, now, s.config.RollCooldown)
		},
	})
}
