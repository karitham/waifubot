package guildpg

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/karitham/waifubot/guild"
	"github.com/karitham/waifubot/storage"
)

func TestToIndexingStatus(t *testing.T) {
	tests := []struct {
		name  string
		input storage.IndexingStatus
		want  guild.IndexingStatus
	}{
		{name: "completed", input: storage.IndexingStatusCompleted, want: guild.IndexingCompleted},
		{name: "in progress", input: storage.IndexingStatusInProgress, want: guild.IndexingInProgress},
		{name: "pending", input: storage.IndexingStatusPending, want: guild.IndexingPending},
		{name: "empty", input: "", want: guild.IndexingPending},
		{name: "unknown", input: "unknown_status", want: guild.IndexingPending},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, toIndexingStatus(tt.input))
		})
	}
}
