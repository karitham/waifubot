package wishlist

import (
	"context"
	"time"

	"github.com/karitham/waifubot/catalog"
	"github.com/karitham/waifubot/collection"
	"github.com/karitham/waifubot/guild"
	"github.com/karitham/waifubot/tracker"
)

// Entry is a character on a user's wishlist.
type Entry struct {
	catalog.Character
	AddedAt time.Time
}

type UserCharacterSet struct {
	UserID     collection.UserID `json:"user_id"`
	Characters []Entry           `json:"characters"`
}

type WishlistComparison struct {
	UserHasCharacters   []Entry `json:"user_has_characters"`
	UserWantsCharacters []Entry `json:"user_wants_characters"`
	MutualMatches       int     `json:"mutual_matches"`
}

type Store interface {
	AddCharactersToWishlist(ctx context.Context, userID collection.UserID, characterIDs []catalog.CharacterID) error
	RemoveCharactersFromWishlist(ctx context.Context, userID collection.UserID, characterIDs []catalog.CharacterID) error
	RemoveAllFromWishlist(ctx context.Context, userID collection.UserID) error
	GetUserCharacterWishlist(ctx context.Context, userID collection.UserID) ([]Entry, error)

	// Discovery
	GetWishlistHolders(ctx context.Context, characterIDs []catalog.CharacterID, userID collection.UserID, guildID guild.ID) ([]UserCharacterSet, error)
	GetWantedCharacters(ctx context.Context, userID collection.UserID, guildID guild.ID) ([]UserCharacterSet, error)
	CompareWithUser(ctx context.Context, userID1, userID2 collection.UserID) (WishlistComparison, error)
	GetUsersWantingCharacter(ctx context.Context, charID catalog.CharacterID, guildID guild.ID, excludeUserID collection.UserID) ([]guild.MemberID, error)
}

// MediaService defines the interface for media operations.
type MediaService interface {
	SearchMedia(ctx context.Context, search string) ([]tracker.Media, error)
	GetMediaCharacters(ctx context.Context, mediaID tracker.MediaID) ([]tracker.Character, error)
}

// AddMediaToWishlist adds all characters from a media to the user's wishlist, filtering out owned characters.
func AddMediaToWishlist(ctx context.Context, wishlistStore Store, mediaService MediaService, store collection.Store, userID collection.UserID, mediaID tracker.MediaID) (int, error) {
	characters, err := mediaService.GetMediaCharacters(ctx, mediaID)
	if err != nil {
		return 0, err
	}

	if len(characters) == 0 {
		return 0, nil
	}

	for _, char := range characters {
		_ = store.UpsertCharacter(ctx, char.ToCatalog())
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
		if !ownedSet[char.ID] {
			characterIDs = append(characterIDs, char.ID)
		}
	}

	if len(characterIDs) == 0 {
		return 0, nil
	}

	batchSize := 50
	added := 0
	for i := 0; i < len(characterIDs); i += batchSize {
		end := min(i+batchSize, len(characterIDs))
		batch := characterIDs[i:end]
		err = wishlistStore.AddCharactersToWishlist(ctx, userID, batch)
		if err != nil {
			return added, err
		}
		added += len(batch)
	}

	return added, nil
}
