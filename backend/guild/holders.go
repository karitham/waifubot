package guild

import (
	"context"
	"fmt"

	"github.com/karitham/waifubot/catalog"
)

// CharacterHolders retrieves users in a guild who own a specific character.
func CharacterHolders(ctx context.Context, store Querier, charStore catalog.Store, guildID ID, charID catalog.CharacterID) (string, []MemberID, error) {
	if guildID == 0 {
		return "", nil, fmt.Errorf("this command can only be used in servers")
	}

	status, err := store.IsGuildIndexed(ctx, guildID)
	if err != nil || status.Status != IndexingCompleted {
		return "", nil, fmt.Errorf("guild members not indexed yet, please try again later")
	}

	char, err := charStore.GetCharacterByID(ctx, charID)
	if err != nil {
		return "", nil, fmt.Errorf("no one in this server has %d", charID)
	}

	holderIDs, err := store.GetCharacterHoldersInGuild(ctx, guildID, charID)
	if err != nil {
		return "", nil, fmt.Errorf("failed to fetch character holders: %w", err)
	}

	return char.Name, holderIDs, nil
}
