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
	return fmt.Sprintf("roll cooldown until %s", e.Until.UTC().Format(time.RFC3339))
}

// Roll executes the free roll for a user, enforcing the cooldown constraint.
func Roll(ctx context.Context, store DB, userID UserID, config Config) (catalog.Character, error) {
	return acquire(ctx, store, userID, acquisition{
		source: SourceRoll,
		check: func(user User, now time.Time) error {
			cooldownUntil := user.LastRoll.Add(config.RollCooldown)
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
		charge: func(ctx context.Context, tx Tx, now time.Time) error {
			return tx.RecordRoll(ctx, userID, now, config.RollCooldown)
		},
	})
}
