package collection

import (
	"context"
	"errors"
)

// CheckOwnership checks if a user owns a specific character.
func CheckOwnership(ctx context.Context, store Store, userID UserID, charID int64) (bool, Character, error) {
	char, err := store.GetOwnedCharacter(ctx, userID, charID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return false, Character{}, nil
		}
		return false, Character{}, err
	}
	return true, char.Character, nil
}
