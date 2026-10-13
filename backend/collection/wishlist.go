package collection

import (
	"context"
	"time"

	"github.com/karitham/waifubot/catalog"
	"github.com/karitham/waifubot/tracker"
)

// WishlistEntry is a character on a user's wishlist.
type WishlistEntry struct {
	catalog.Character
	AddedAt time.Time
}

// WishlistUserSet is a user and the characters they have or want.
type WishlistUserSet struct {
	UserID     UserID          `json:"user_id"`
	Characters []WishlistEntry `json:"characters"`
}

// WishlistComparison is the overlap between two users' collections and
// wishlists.
type WishlistComparison struct {
	UserHasCharacters   []WishlistEntry `json:"user_has_characters"`
	UserWantsCharacters []WishlistEntry `json:"user_wants_characters"`
	MutualMatches       int             `json:"mutual_matches"`
}

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

// AddMediaToWishlist adds all characters from a media to the user's wishlist,
// filtering out characters absent from the catalog and ones the user
// already owns. The user is materialized and the list is added in one
// transaction.
func AddMediaToWishlist(ctx context.Context, store DB, anime tracker.Service, userID UserID, mediaID tracker.MediaID) (int, error) {
	characters, err := anime.GetMediaCharacters(ctx, mediaID)
	if err != nil {
		return 0, err
	}

	if len(characters) == 0 {
		return 0, nil
	}

	ids := make([]catalog.CharacterID, len(characters))
	for i, char := range characters {
		ids[i] = char.ID
	}

	// Only the hydrator writes characters to the catalog; entries for
	// characters not yet written are invisible to the reads that join
	// characters.
	known, err := store.KnownCharacterIDs(ctx, ids)
	if err != nil {
		return 0, err
	}

	knownSet := make(map[catalog.CharacterID]bool, len(known))
	for _, id := range known {
		knownSet[id] = true
	}

	ownedIDs, err := store.GetCollectionIDs(ctx, userID)
	if err != nil {
		return 0, err
	}

	ownedSet := make(map[catalog.CharacterID]bool, len(ownedIDs))
	for _, id := range ownedIDs {
		ownedSet[id] = true
	}

	var characterIDs []catalog.CharacterID
	for _, char := range characters {
		if knownSet[char.ID] && !ownedSet[char.ID] {
			characterIDs = append(characterIDs, char.ID)
		}
	}

	if len(characterIDs) == 0 {
		return 0, nil
	}

	added, err := AddToWishlist(ctx, store, userID, characterIDs)
	if err != nil {
		return 0, err
	}

	return int(added), nil
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
