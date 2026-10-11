package collection

import (
	"context"
	"errors"

	"github.com/karitham/waifubot/catalog"
)

// CheckOwnership checks if a user owns a specific character.
func CheckOwnership(ctx context.Context, store Store, userID UserID, charID catalog.CharacterID) (bool, catalog.Character, error) {
	char, err := store.GetOwnedCharacter(ctx, userID, charID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return false, catalog.Character{}, nil
		}
		return false, catalog.Character{}, err
	}
	return true, char.Character, nil
}
