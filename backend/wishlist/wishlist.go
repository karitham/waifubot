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

type UserCharacterSet struct {
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
	RemoveAllFromWishlist(ctx context.Context, userID uint64) error
	GetUserCharacterWishlist(ctx context.Context, userID uint64) ([]Character, error)

	// Discovery
	GetWishlistHolders(ctx context.Context, characterIDs []int64, userID, guildID uint64) ([]UserCharacterSet, error)
	GetWantedCharacters(ctx context.Context, userID, guildID uint64) ([]UserCharacterSet, error)
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

	// Add characters to wishlist in batches of 50
	batchSize := 50
	added := 0
	for i := 0; i < len(characterIDs); i += batchSize {
		end := min(i+batchSize, len(characterIDs))
		batch := characterIDs[i:end]
		err = wishlistStore.AddMultipleCharactersToWishlist(ctx, uint64(userID), batch)
		if err != nil {
			return added, err
		}
		added += len(batch)
	}

	return added, nil
}
