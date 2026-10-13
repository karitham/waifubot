package collection

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/karitham/waifubot/drops"
)

// Claim claims the channel's drop for the user when the name matches the
// dropped character. The drop row is locked for the transaction, so
// concurrent claims serialize and the drop is consumed exactly once.
func Claim(ctx context.Context, store DB, userID UserID, channel drops.ChannelID, charName string) (drops.Drop, error) {
	var drop drops.Drop
	err := withUsers(ctx, store, []UserID{userID}, func(tx Tx) error {
		var err error
		drop, err = tx.GetDropForUpdate(ctx, channel)
		if err != nil {
			return fmt.Errorf("failed to get drop: %w", err)
		}

		if !drop.Matches(charName) {
			return drops.ErrWrongCharacterName
		}

		err = tx.AddToCollection(ctx, userID, drop.Character.ID, SourceClaim, time.Now())
		if err != nil {
			if errors.Is(err, ErrAlreadyOwned) {
				return ErrAlreadyOwned
			}
			return fmt.Errorf("failed to insert character into collection: %w", err)
		}

		if err := tx.RemoveFromWishlist(ctx, userID, drop.Character.ID); err != nil {
			return err
		}

		if err := tx.DeleteDrop(ctx, channel); err != nil {
			return fmt.Errorf("failed to delete drop: %w", err)
		}
		return nil
	})
	if err != nil {
		return drops.Drop{}, err
	}

	return drop, nil
}
