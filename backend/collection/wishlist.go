package collection

import (
	"context"

	"github.com/karitham/waifubot/catalog"
)

// RemoveFromWishlist removes a character from the user's wishlist. It does
// not materialize the user: removing from an absent wishlist is a no-op.
func RemoveFromWishlist(ctx context.Context, store DB, userID UserID, charID catalog.CharacterID) error {
	return withTx(ctx, store, func(tx Tx) error {
		return tx.RemoveFromWishlist(ctx, userID, charID)
	})
}

// RemoveAllFromWishlist clears the user's wishlist without materializing the
// user.
func RemoveAllFromWishlist(ctx context.Context, store DB, userID UserID) error {
	return withTx(ctx, store, func(tx Tx) error {
		return tx.RemoveAllFromWishlist(ctx, userID)
	})
}

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
