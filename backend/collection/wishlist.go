package collection

import (
	"context"

	"github.com/karitham/waifubot/catalog"
)

// AddToWishlist adds characters to the user's wishlist, materializing the
// user first. It reports how many characters were actually added.
func AddToWishlist(ctx context.Context, store DB, userID UserID, characterIDs []catalog.CharacterID) (int64, error) {
	var added int64
	err := withUsers(ctx, store, []UserID{userID}, func(tx Tx) error {
		var err error
		added, err = tx.AddCharactersToWishlist(ctx, userID, characterIDs)
		return err
	})
	if err != nil {
		return 0, err
	}
	return added, nil
}
