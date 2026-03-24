package discord

import (
	"context"
	"fmt"
	"strings"

	"github.com/karitham/waifubot/collection"
)

func collectionCheckOwnership(ctx context.Context, store collection.Store, userID uint64, charID int64) (bool, collection.Character, error) {
	return collection.CheckOwnership(ctx, store, userID, charID)
}

func formatUsersWantingCharacter(userIDs []uint64, excludeUserID uint64) string {
	var filtered []uint64
	for _, id := range userIDs {
		if id != excludeUserID {
			filtered = append(filtered, id)
		}
	}

	if len(filtered) == 0 {
		return ""
	}

	var (
		maxDisplay = 3

		sb strings.Builder
	)

	for i := range min(maxDisplay, len(filtered)) {
		if i > 0 {
			sb.WriteString(" ")
		}

		sb.WriteString(formatUserMention(filtered[i]))
	}

	if len(filtered) > maxDisplay {
		sb.WriteString(fmt.Sprintf(" +%d more", len(filtered)-maxDisplay))
	}

	return "\n\n🤝 " + sb.String() + " is looking to trade!"
}

func formatUserMention(userID uint64) string {
	return fmt.Sprintf("<@%d>", userID)
}
