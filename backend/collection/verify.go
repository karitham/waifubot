package collection

import (
	"context"

	"github.com/Karitham/corde"

	"github.com/karitham/waifubot/storage/collectionstore"
)

// CheckOwnership checks if a user owns a specific character
func CheckOwnership(ctx context.Context, store Store, userID corde.Snowflake, charID int64) (bool, collectionstore.Character, error) {
	char, err := store.CollectionStore().Get(ctx, collectionstore.GetParams{ID: charID, UserID: uint64(userID)})
	if err != nil {
		return false, collectionstore.Character{}, err
	}
	return char.ID == charID, collectionstore.Character{
		ID:    char.ID,
		Name:  char.Name,
		Image: char.Image,
	}, nil
}

// SearchGlobalCharacters searches for characters globally for autocomplete
func SearchGlobalCharacters(ctx context.Context, store Store, term string) ([]collectionstore.Character, error) {
	return store.CollectionStore().SearchGlobalCharacters(ctx, collectionstore.SearchGlobalCharactersParams{Term: term, Lim: 25})
}
