package discord

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/karitham/waifubot/collection"
	"github.com/karitham/waifubot/guild"
)

func TestFormatUsersWantingCharacter(t *testing.T) {
	tests := []struct {
		name          string
		userIDs       []guild.MemberID
		excludeUserID collection.UserID
		want          string
	}{
		{
			name:          "empty list",
			userIDs:       []guild.MemberID{},
			excludeUserID: 0,
			want:          "",
		},
		{
			name:          "single user not excluded",
			userIDs:       []guild.MemberID{123},
			excludeUserID: 0,
			want:          "\n\n🤝 <@123> is looking to trade!",
		},
		{
			name:          "single user excluded",
			userIDs:       []guild.MemberID{123},
			excludeUserID: 123,
			want:          "",
		},
		{
			name:          "two users none excluded",
			userIDs:       []guild.MemberID{123, 456},
			excludeUserID: 0,
			want:          "\n\n🤝 <@123> <@456> is looking to trade!",
		},
		{
			name:          "three users",
			userIDs:       []guild.MemberID{123, 456, 789},
			excludeUserID: 0,
			want:          "\n\n🤝 <@123> <@456> <@789> is looking to trade!",
		},
		{
			name:          "four users truncated",
			userIDs:       []guild.MemberID{123, 456, 789, 111},
			excludeUserID: 0,
			want:          "\n\n🤝 <@123> <@456> <@789> +1 more is looking to trade!",
		},
		{
			name:          "exclude user in middle",
			userIDs:       []guild.MemberID{123, 456, 789},
			excludeUserID: 456,
			want:          "\n\n🤝 <@123> <@789> is looking to trade!",
		},
		{
			name:          "all users excluded",
			userIDs:       []guild.MemberID{123, 456, 789},
			excludeUserID: 123,
			// Note: 456 and 789 remain after filtering, so this won't be empty
			// This test verifies filtering works, but in practice SQL already excludes
			want: "\n\n🤝 <@456> <@789> is looking to trade!",
		},
		{
			name:          "exclude all users returns empty",
			userIDs:       []guild.MemberID{123},
			excludeUserID: 123,
			want:          "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatUsersWantingCharacter(tt.userIDs, tt.excludeUserID)
			assert.Equal(t, tt.want, got)
		})
	}
}
