package collection

import (
	"context"
	"errors"
	"fmt"
)

// Exchange sells a character for 1 token.
func Exchange(ctx context.Context, store Store, userID UserID, charID int64) (OwnedCharacter, error) {
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

	charInfo, err := tx.GetCharacterByID(ctx, charID)
	if err != nil {
		return OwnedCharacter{}, err
	}

	owned, err := tx.GetOwnedCharacter(ctx, userID, charID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return OwnedCharacter{}, fmt.Errorf("%w %d", ErrUserDoesNotOwnCharacter, charID)
		}
		return OwnedCharacter{}, err
	}

	err = tx.RemoveFromCollection(ctx, userID, charID)
	if err != nil {
		return OwnedCharacter{}, err
	}

	_, err = tx.AddTokens(ctx, userID, 1)
	if err != nil {
		return OwnedCharacter{}, err
	}

	err = tx.Commit(ctx)
	committed = err == nil

	return OwnedCharacter{
		Character: charInfo,
		Date:      owned.Date,
		Source:    "EXCHANGE",
		UserID:    userID,
	}, err
}
