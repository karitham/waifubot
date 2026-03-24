package discord

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatUsersWantingCharacter(t *testing.T) {
	tests := []struct {
		name          string
		userIDs       []uint64
		excludeUserID uint64
		want          string
	}{
		{
			name:          "empty list",
			userIDs:       []uint64{},
			excludeUserID: 0,
			want:          "",
		},
		{
			name:          "single user not excluded",
			userIDs:       []uint64{123},
			excludeUserID: 0,
			want:          "\n\n<@123> also want this character",
		},
		{
			name:          "single user excluded",
			userIDs:       []uint64{123},
			excludeUserID: 123,
			want:          "",
		},
		{
			name:          "two users none excluded",
			userIDs:       []uint64{123, 456},
			excludeUserID: 0,
			want:          "\n\n<@123> <@456> also want this character",
		},
		{
			name:          "three users",
			userIDs:       []uint64{123, 456, 789},
			excludeUserID: 0,
			want:          "\n\n<@123> <@456> <@789> also want this character",
		},
		{
			name:          "four users truncated",
			userIDs:       []uint64{123, 456, 789, 111},
			excludeUserID: 0,
			want:          "\n\n<@123> <@456> <@789>... also want this character",
		},
		{
			name:          "exclude user in middle",
			userIDs:       []uint64{123, 456, 789},
			excludeUserID: 456,
			want:          "\n\n<@123> <@789> also want this character",
		},
		{
			name:          "all users excluded",
			userIDs:       []uint64{123, 456, 789},
			excludeUserID: 123,
			// Note: 456 and 789 remain after filtering, so this won't be empty
			// This test verifies filtering works, but in practice SQL already excludes
			want: "\n\n<@456> <@789> also want this character",
		},
		{
			name:          "exclude all users returns empty",
			userIDs:       []uint64{123},
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
