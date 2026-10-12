package collection

import (
	"context"
	"fmt"

	"github.com/karitham/waifubot/catalog"
)

// CreateDrop picks a random active character and stores it as the channel's
// drop, replacing any previous one.
func CreateDrop(ctx context.Context, store Store, channelID ChannelID) (catalog.Character, error) {
	char, err := store.RandomActiveChar(ctx, DropWeightExponent)
	if err != nil {
		return catalog.Character{}, fmt.Errorf("failed to pick a drop character: %w", err)
	}

	if err := store.SetDrop(ctx, channelID, char); err != nil {
		return catalog.Character{}, fmt.Errorf("failed to store the drop: %w", err)
	}

	return char, nil
}
