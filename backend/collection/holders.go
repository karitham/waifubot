package collection

import (
	"context"
	"fmt"

	"github.com/Karitham/corde"

	"github.com/karitham/waifubot/storage/guildstore"
)

// CharacterHolders retrieves users in a guild who own a specific character
func CharacterHolders(ctx context.Context, store Store, guildID corde.Snowflake, charID int64) (string, []corde.Snowflake, error) {
	if guildID == 0 {
		return "", nil, fmt.Errorf("this command can only be used in servers")
	}

	indexed, err := store.GuildStore().IsGuildIndexed(ctx, uint64(guildID))
	if err != nil || indexed.Status != guildstore.IndexingStatusCompleted {
		return "", nil, fmt.Errorf("guild members not indexed yet, please try again later")
	}

	charRow, err := store.CollectionStore().GetByID(ctx, charID)
	if err != nil {
		return "", nil, fmt.Errorf("no one in this server has %d", charID)
	}

	holderIDsInt, err := store.GuildStore().UsersOwningCharInGuild(ctx, guildstore.UsersOwningCharInGuildParams{CharacterID: charID, GuildID: uint64(guildID)})
	if err != nil {
		return "", nil, fmt.Errorf("failed to fetch character holders: %w", err)
	}

	holderIDs := make([]corde.Snowflake, len(holderIDsInt))
	for i, id := range holderIDsInt {
		holderIDs[i] = corde.Snowflake(id)
	}

	return charRow.Name, holderIDs, nil
}
