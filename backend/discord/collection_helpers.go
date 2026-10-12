package discord

import (
	"fmt"
	"strings"

	"github.com/karitham/waifubot/collection"
	"github.com/karitham/waifubot/guild"
)

func formatUsersWantingCharacter(memberIDs []guild.MemberID, excludeUserID collection.UserID) string {
	var filtered []guild.MemberID
	for _, id := range memberIDs {
		if id != guild.MemberID(excludeUserID) {
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
		fmt.Fprintf(&sb, " +%d more", len(filtered)-maxDisplay)
	}

	return "\n\n🤝 " + sb.String() + " is looking to trade!"
}

func formatUserMention(memberID guild.MemberID) string {
	return fmt.Sprintf("<@%d>", memberID)
}
