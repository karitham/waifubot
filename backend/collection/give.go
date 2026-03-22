package collection

import (
	"context"
	"errors"
	"fmt"
)

// Give executes the give logic from one user to another.
func Give(ctx context.Context, store Store, from, to UserID, charID int64) (OwnedCharacter, error) {
	tx, err := store.WithTx(ctx)
	if err != nil {
		return OwnedCharacter{}, err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback(ctx)
		}
	}()

	_, err = tx.GetOwnedCharacter(ctx, from, charID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return OwnedCharacter{}, fmt.Errorf("%w %d", ErrUserDoesNotOwnCharacter, charID)
		}
		return OwnedCharacter{}, fmt.Errorf("error checking ownership: %w", err)
	}

	_, err = tx.GetOwnedCharacter(ctx, to, charID)
	if err == nil {
		return OwnedCharacter{}, fmt.Errorf("to user already owns char %d", charID)
	}

	given, err := tx.GiveCharacter(ctx, from, to, charID)
	if err != nil {
		return OwnedCharacter{}, fmt.Errorf("error giving char: %w", err)
	}

	_ = tx.RemoveFromWishlist(ctx, to, charID)

	err = tx.Commit(ctx)
	committed = err == nil
	return given, err
}
