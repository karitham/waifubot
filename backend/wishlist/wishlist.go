package wishlist

import (
	"context"
)

type Character struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Image string `json:"image"`
	Date  string `json:"date"`
}

type WishlistHolder struct {
	UserID     uint64      `json:"user_id"`
	Characters []Character `json:"characters"`
}

type WantedCharacter struct {
	UserID     uint64      `json:"user_id"`
	Characters []Character `json:"characters"`
}

type WishlistComparison struct {
	UserHasCharacters   []Character `json:"user_has_characters"`
	UserWantsCharacters []Character `json:"user_wants_characters"`
	MutualMatches       int         `json:"mutual_matches"`
}

type Store interface {
	// Character Wishlist
	AddCharacterToWishlist(ctx context.Context, userID uint64, characterID int64) error
	AddMultipleCharactersToWishlist(ctx context.Context, userID uint64, characterIDs []int64) error
	RemoveCharacterFromWishlist(ctx context.Context, userID uint64, characterID int64) error
	GetUserCharacterWishlist(ctx context.Context, userID uint64) ([]Character, error)

	// Discovery
	GetWishlistHolders(ctx context.Context, userID uint64, guildID uint64) ([]WishlistHolder, error)
	GetWantedCharacters(ctx context.Context, userID uint64) ([]WantedCharacter, error)
	CompareWithUser(ctx context.Context, userID1, userID2 uint64) (WishlistComparison, error)
}

// AddCharacter adds a character to the user's wishlist
func AddCharacter(ctx context.Context, s Store, userID uint64, characterID int64) error {
	return s.AddCharacterToWishlist(ctx, userID, characterID)
}

// RemoveCharacter removes a character from the user's wishlist
func RemoveCharacter(ctx context.Context, s Store, userID uint64, characterID int64) error {
	return s.RemoveCharacterFromWishlist(ctx, userID, characterID)
}

// GetUserWishlist gets the user's wishlist
func GetUserWishlist(ctx context.Context, s Store, userID uint64) ([]Character, error) {
	return s.GetUserCharacterWishlist(ctx, userID)
}

// GetWishlistHolders gets users who have characters from the user's wishlist
func GetWishlistHolders(ctx context.Context, s Store, userID uint64, guildID uint64) ([]WishlistHolder, error) {
	return s.GetWishlistHolders(ctx, userID, guildID)
}

// GetWantedCharacters gets users who want characters the user owns
func GetWantedCharacters(ctx context.Context, s Store, userID uint64) ([]WantedCharacter, error) {
	return s.GetWantedCharacters(ctx, userID)
}

// CompareWithUser compares the user's collection with another user's wishlist
func CompareWithUser(ctx context.Context, s Store, userID1, userID2 uint64) (WishlistComparison, error) {
	return s.CompareWithUser(ctx, userID1, userID2)
}
