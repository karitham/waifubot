package collection

import (
	"context"
)

// Characters retrieves a user's character collection.
func Characters(ctx context.Context, store Store, userID UserID) ([]OwnedCharacter, error) {
	return store.GetCollection(ctx, userID)
}
