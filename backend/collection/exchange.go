package collection

import (
	"context"

	"github.com/Karitham/corde"

	"github.com/karitham/waifubot/storage/collectionstore"
)

// Exchange executes the exchange logic for a character
func Exchange(ctx context.Context, store Store, userID corde.Snowflake, charID int64) (collectionstore.Character, error) {
	txI, err := store.Tx(ctx)
	if err != nil {
		return collectionstore.Character{}, err
	}
	tx := txI.(Store)

	var char collectionstore.Character
	err = func() error {
		c, err := tx.CollectionStore().Delete(ctx, collectionstore.DeleteParams{UserID: uint64(userID), ID: charID})
		if err != nil {
			return err
		}
		char = c
		return tx.UserStore().IncTokens(ctx, uint64(userID))
	}()
	if err != nil {
		_ = tx.Rollback(ctx)
		return collectionstore.Character{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return collectionstore.Character{}, err
	}

	return char, nil
}
