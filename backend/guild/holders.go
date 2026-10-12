package guild

import (
	"context"
	"errors"
	"fmt"

	"github.com/karitham/waifubot/catalog"
)

var (
	// ErrNotInGuild is returned when the operation needs a guild context.
	ErrNotInGuild = errors.New("not in a guild")
	// ErrNotIndexed is returned when the guild's member index is not ready.
	ErrNotIndexed = errors.New("guild members not indexed")
)

// CharacterHolders retrieves users in a guild who own a specific character.
func CharacterHolders(ctx context.Context, store Querier, charStore catalog.Store, guildID ID, charID catalog.CharacterID) (string, []MemberID, error) {
	if guildID == 0 {
		return "", nil, ErrNotInGuild
	}

	status, err := store.IsGuildIndexed(ctx, guildID)
	if err != nil {
		return "", nil, fmt.Errorf("failed to read indexing status: %w", err)
	}
	if status.Status != IndexingCompleted {
		return "", nil, ErrNotIndexed
	}

	char, err := charStore.GetCharacterByID(ctx, charID)
	if err != nil {
		return "", nil, err
	}

	holderIDs, err := store.GetCharacterHoldersInGuild(ctx, guildID, charID)
	if err != nil {
		return "", nil, fmt.Errorf("failed to fetch character holders: %w", err)
	}

	return char.Name, holderIDs, nil
}
