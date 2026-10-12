package collection

import (
	"context"

	"github.com/karitham/waifubot/catalog"
)

// AddToWishlist adds characters to the user's wishlist, materializing the
// user first.
func AddToWishlist(ctx context.Context, store DB, userID UserID, characterIDs []catalog.CharacterID) error {
	return withUsers(ctx, store, []UserID{userID}, func(tx Tx) error {
		return tx.AddCharactersToWishlist(ctx, userID, characterIDs)
	})
}
