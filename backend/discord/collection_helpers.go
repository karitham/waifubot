package discord

import (
	"context"

	"github.com/karitham/waifubot/collection"
)

// collectionCheckOwnership wraps collection.CheckOwnership for discord handlers.
func collectionCheckOwnership(ctx context.Context, store collection.Store, userID uint64, charID int64) (bool, collection.Character, error) {
	return collection.CheckOwnership(ctx, store, userID, charID)
}
