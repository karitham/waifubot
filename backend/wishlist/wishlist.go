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

// AddMediaToWishlist adds all characters from a media to the user's wishlist,
// filtering out characters the catalog does not know and ones the user
// already owns. The user is materialized and the list is added in one
// transaction.
func AddMediaToWishlist(ctx context.Context, mediaService MediaService, store collection.DB, userID collection.UserID, mediaID tracker.MediaID) (int, error) {
	characters, err := mediaService.GetMediaCharacters(ctx, mediaID)
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

	added, err := collection.AddToWishlist(ctx, store, userID, characterIDs)
	if err != nil {
		return 0, err
	}

	return int(added), nil
}
