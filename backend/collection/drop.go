package collection

import (
	"context"
	"fmt"

	"github.com/karitham/waifubot/catalog"
)

// CreateDrop picks a random active character and stores it as the channel's
// drop, replacing any previous one. The pick reads the catalog, so the drop
// only ever references an existing character.
func CreateDrop(ctx context.Context, store DB, channelID ChannelID) (catalog.Character, error) {
	char, err := store.RandomActiveChar(ctx, DropWeightExponent)
	if err != nil {
		return catalog.Character{}, fmt.Errorf("failed to pick a drop character: %w", err)
	}

	tx, err := store.Begin(ctx)
	if err != nil {
		return catalog.Character{}, err
	}

	if err := tx.SetDrop(ctx, channelID, char); err != nil {
		_ = tx.Rollback(ctx)
		return catalog.Character{}, fmt.Errorf("failed to store the drop: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		_ = tx.Rollback(ctx)
		return catalog.Character{}, err
	}
	return char, nil
}
