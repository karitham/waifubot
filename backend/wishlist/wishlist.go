package wishlist

import (
	"context"

	"github.com/Karitham/corde"

	"github.com/karitham/waifubot/collection"
	"github.com/karitham/waifubot/storage/collectionstore"
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
	AddMultipleCharactersToWishlist(ctx context.Context, userID uint64, characterIDs []int64) error
	RemoveMultipleCharactersFromWishlist(ctx context.Context, userID uint64, characterIDs []int64) error
	GetUserCharacterWishlist(ctx context.Context, userID uint64) ([]Character, error)

	// Discovery
	GetWishlistHolders(ctx context.Context, userID uint64, guildID uint64) ([]WishlistHolder, error)
	GetWantedCharacters(ctx context.Context, userID uint64) ([]WantedCharacter, error)
	CompareWithUser(ctx context.Context, userID1, userID2 uint64) (WishlistComparison, error)
}

// MediaService defines the interface for media operations
type MediaService interface {
	SearchMedia(ctx context.Context, search string) ([]collection.Media, error)
	GetMediaCharacters(ctx context.Context, mediaId int64) ([]collection.MediaCharacter, error)
}

// CollectionService defines the interface for collection operations
type CollectionService interface {
	CheckOwnership(ctx context.Context, userID corde.Snowflake, charID int64) (bool, collectionstore.Character, error)
	GetUserCollectionIDs(ctx context.Context, userID corde.Snowflake) ([]int64, error)
	UpsertCharacter(ctx context.Context, charID int64, name, image string) error
}

// AddCharacter adds a character to the user's wishlist
func AddCharacter(ctx context.Context, s Store, userID uint64, characterID int64) error {
	return s.AddMultipleCharactersToWishlist(ctx, userID, []int64{characterID})
}

// AddMultipleCharacters adds multiple characters to the user's wishlist
func AddMultipleCharacters(ctx context.Context, s Store, userID uint64, characterIDs []int64) error {
	return s.AddMultipleCharactersToWishlist(ctx, userID, characterIDs)
}

// RemoveCharacter removes a character from the user's wishlist
func RemoveCharacter(ctx context.Context, s Store, userID uint64, characterID int64) error {
	return s.RemoveMultipleCharactersFromWishlist(ctx, userID, []int64{characterID})
}

// RemoveMultipleCharacters removes multiple characters from the user's wishlist
func RemoveMultipleCharacters(ctx context.Context, s Store, userID uint64, characterIDs []int64) error {
	return s.RemoveMultipleCharactersFromWishlist(ctx, userID, characterIDs)
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

// AddMediaToWishlist adds all characters from a media to the user's wishlist, filtering out owned characters
func AddMediaToWishlist(ctx context.Context, wishlistStore Store, mediaService MediaService, collectionService CollectionService, userID corde.Snowflake, mediaID int64) (int, error) {
	// Get characters from the media
	characters, err := mediaService.GetMediaCharacters(ctx, mediaID)
	if err != nil {
		return 0, err
	}

	if len(characters) == 0 {
		return 0, nil
	}

	// Ensure all characters exist in the characters table
	for _, char := range characters {
		err := collectionService.UpsertCharacter(ctx, char.ID, char.Name, char.ImageURL)
		if err != nil {
			continue // Skip on error, don't fail the whole operation
		}
	}

	// Get all character IDs the user owns in one query
	ownedIDs, err := collectionService.GetUserCollectionIDs(ctx, userID)
	if err != nil {
		return 0, err
	}

	// Create a set of owned character IDs for efficient lookup
	ownedSet := make(map[int64]bool, len(ownedIDs))
	for _, id := range ownedIDs {
		ownedSet[id] = true
	}

	// Filter out characters the user already owns
	var characterIDs []int64
	for _, char := range characters {
		if !ownedSet[char.ID] {
			characterIDs = append(characterIDs, char.ID)
		}
	}

	if len(characterIDs) == 0 {
		return 0, nil
	}

	// Add characters to wishlist
	err = wishlistStore.AddMultipleCharactersToWishlist(ctx, uint64(userID), characterIDs)
	if err != nil {
		return 0, err
	}

	return len(characterIDs), nil
}
