package collection

import (
	"context"

	"github.com/Karitham/corde"

	"github.com/karitham/waifubot/storage/collectionstore"
)

// Exchange executes the exchange logic for a character
func Exchange(ctx context.Context, store Store, userID corde.Snowflake, charID int64) (collectionstore.Character, error) {
	tx, err := store.Tx(ctx)
	if err != nil {
		return collectionstore.Character{}, err
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	char, err := tx.CollectionStore().Delete(ctx, collectionstore.DeleteParams{UserID: uint64(userID), ID: charID})
	if err != nil {
		return collectionstore.Character{}, err
	}

	err = tx.UserStore().IncTokens(ctx, uint64(userID))
	if err != nil {
		return collectionstore.Character{}, err
	}

	err = tx.Commit(ctx)
	return char, err
}
