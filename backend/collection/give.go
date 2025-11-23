package collection

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Karitham/corde"

	"github.com/karitham/waifubot/storage/collectionstore"
)

// Give executes the give logic from one user to another
func Give(ctx context.Context, store Store, from, to corde.Snowflake, charID int64) (Character, error) {
	c, err := store.CollectionStore().Get(ctx, collectionstore.GetParams{ID: charID, UserID: uint64(from)})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Character{}, fmt.Errorf("%w %d", ErrUserDoesNotOwnCharacter, charID)
		}
		return Character{}, fmt.Errorf("error checking ownership: %w", err)
	}

	_, err = store.CollectionStore().Get(ctx, collectionstore.GetParams{ID: charID, UserID: uint64(to)})
	if err == nil {
		return Character{}, fmt.Errorf("to user already owns char %d", charID)
	}

	_, err = store.CollectionStore().Give(ctx, collectionstore.GiveParams{UserID: uint64(to), CharacterID: charID, UserID_2: uint64(from)})
	if err != nil {
		return Character{}, fmt.Errorf("error giving char: %w", err)
	}

	return Character{
		Date:   c.Date.Time,
		Image:  c.Image,
		Name:   c.Name,
		Type:   c.Source,
		UserID: corde.Snowflake(to),
		ID:     c.ID,
	}, nil
}
