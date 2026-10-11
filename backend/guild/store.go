package guild

import (
	"context"
	"time"

	"github.com/karitham/waifubot/catalog"
)

// GuildID identifies a Discord guild.
type GuildID uint64

// MemberID identifies a guild member (a user, in guild context).
type MemberID uint64

type IndexingStatus int

const (
	IndexingPending IndexingStatus = iota
	IndexingInProgress
	IndexingCompleted
)

type GuildIndexStatus struct {
	Status    IndexingStatus
	UpdatedAt time.Time
}

// ConvertIndexingStatus maps a database indexing-status string to the domain type.
func ConvertIndexingStatus(s string) IndexingStatus {
	switch s {
	case "completed":
		return IndexingCompleted
	case "in_progress":
		return IndexingInProgress
	default:
		return IndexingPending
	}
}

// GuildQuerier provides guild indexing operations.
type GuildQuerier interface {
	IsGuildIndexed(ctx context.Context, guildID GuildID) (GuildIndexStatus, error)
	StartIndexingJob(ctx context.Context, guildID GuildID) error
	CompleteIndexingJob(ctx context.Context, guildID GuildID) error
	UpsertGuildMembers(ctx context.Context, guildID GuildID, memberIDs []MemberID, indexedAt time.Time) error
	DeleteGuildMembersNotIn(ctx context.Context, guildID GuildID, memberIDs []MemberID) error
}

// HolderQuerier looks up guild members owning a character.
type HolderQuerier interface {
	GetCharacterHoldersInGuild(ctx context.Context, guildID GuildID, charID catalog.CharacterID) ([]MemberID, error)
}

// TxQuerier is a GuildQuerier inside a transaction.
type TxQuerier interface {
	GuildQuerier
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}
