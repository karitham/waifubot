package collection

import (
	"context"
	"errors"
	"fmt"

	"github.com/karitham/waifubot/catalog"
)

// sellReward is granted for selling a character.
var sellReward = TokenAmount{amount: 1}

// Exchange sells a character for tokens.
func Exchange(ctx context.Context, store DB, userID UserID, charID catalog.CharacterID) (OwnedCharacter, error) {
	var (
		charInfo catalog.Character
		owned    OwnedCharacter
	)
	err := withTx(ctx, store, func(tx Tx) error {
		var err error
		charInfo, err = tx.GetCharacterByID(ctx, charID)
		if err != nil {
			return err
		}

		owned, err = tx.GetOwnedCharacter(ctx, userID, charID)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				return fmt.Errorf("%w %d", ErrUserDoesNotOwnCharacter, charID)
			}
			return err
		}

		if err := tx.RemoveFromCollection(ctx, userID, charID); err != nil {
			return err
		}

		if _, err := tx.AddTokens(ctx, userID, sellReward); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return OwnedCharacter{}, err
	}

	return OwnedCharacter{
		Character: charInfo,
		Date:      owned.Date,
		Source:    SourceExchange,
		UserID:    userID,
	}, nil
}
