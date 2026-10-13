package collection

import (
	"context"
	"fmt"

	"github.com/karitham/waifubot/catalog"
)

// CreateDrop picks a random active character and stores it as the channel's
// drop, replacing any previous one. The pick is a catalog read, so the drop
// only ever references an existing character.
func CreateDrop(ctx context.Context, store DB, channelID ChannelID) (catalog.Character, error) {
	char, err := store.RandomActiveChar(ctx, DropWeightExponent)
	if err != nil {
		return catalog.Character{}, fmt.Errorf("failed to pick a drop character: %w", err)
	}

	err = withTx(ctx, store, func(tx Tx) error {
		return tx.SetDrop(ctx, channelID, char)
	})
	if err != nil {
		return catalog.Character{}, fmt.Errorf("failed to store the drop: %w", err)
	}
	return char, nil
}
