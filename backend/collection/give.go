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
			return fmt.Errorf("%w %d", ErrAlreadyOwned, charID)
		} else if !errors.Is(err, ErrNotFound) {
			return fmt.Errorf("error checking recipient ownership: %w", err)
		}

		var err error
		given, err = tx.GiveCharacter(ctx, from, to, charID)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				return fmt.Errorf("%w %d", ErrUserDoesNotOwnCharacter, charID)
			}
			return fmt.Errorf("error giving char: %w", err)
		}

		if err := tx.RemoveFromWishlist(ctx, to, charID); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return OwnedCharacter{}, err
	}

	return given, nil
}
