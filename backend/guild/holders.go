package guild

import (
	"context"
	"fmt"

	"github.com/karitham/waifubot/collection"
)

// CharacterHolders retrieves users in a guild who own a specific character.
func CharacterHolders(ctx context.Context, store GuildQuerier, charStore CharacterQuerier, guildID uint64, charID int64) (string, []uint64, error) {
	if guildID == 0 {
		return "", nil, fmt.Errorf("this command can only be used in servers")
	}

	status, err := store.IsGuildIndexed(ctx, guildID)
	if err != nil || status.Status != collection.IndexingCompleted {
		return "", nil, fmt.Errorf("guild members not indexed yet, please try again later")
	}

	char, err := charStore.GetCharacterByID(ctx, charID)
	if err != nil {
		return "", nil, fmt.Errorf("no one in this server has %d", charID)
	}

	holderIDs, err := charStore.GetCharacterHoldersInGuild(ctx, guildID, charID)
	if err != nil {
		return "", nil, fmt.Errorf("failed to fetch character holders: %w", err)
	}

	return char.Name, holderIDs, nil
}
