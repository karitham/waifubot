package collection

import (
	"context"
	"errors"
	"fmt"

	"github.com/karitham/waifubot/catalog"
)

// Give executes the give logic from one user to another.
func Give(ctx context.Context, store DB, from, to UserID, charID catalog.CharacterID) (OwnedCharacter, error) {
	var given OwnedCharacter
	err := withUsers(ctx, store, []UserID{from, to}, func(tx Tx) error {
		if _, err := tx.GetOwnedCharacter(ctx, from, charID); err != nil {
			if errors.Is(err, ErrNotFound) {
				return fmt.Errorf("%w %d", ErrUserDoesNotOwnCharacter, charID)
			}
			return fmt.Errorf("error checking ownership: %w", err)
		}

		if _, err := tx.GetOwnedCharacter(ctx, to, charID); err == nil {
			return fmt.Errorf("to user already owns char %d", charID)
		}

		var err error
		given, err = tx.GiveCharacter(ctx, from, to, charID)
		if err != nil {
			return fmt.Errorf("error giving char: %w", err)
		}

		_ = tx.RemoveFromWishlist(ctx, to, charID)
		return nil
	})
	if err != nil {
		return OwnedCharacter{}, err
	}

	return given, nil
}
