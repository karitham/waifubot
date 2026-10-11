package guild

import (
	"context"
	"time"

	"github.com/karitham/waifubot/catalog"
)

// ID identifies a Discord guild.
type ID uint64

// MemberID identifies a guild member (a user, in guild context).
type MemberID uint64

// IndexingStatus is the state of a guild's member index.
type IndexingStatus int

const (
	IndexingPending IndexingStatus = iota
	IndexingInProgress
	IndexingCompleted
)

// IndexStatus reports the state and freshness of a guild's member index.
type IndexStatus struct {
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

// Querier provides guild indexing and holder lookups.
type Querier interface {
	IsGuildIndexed(ctx context.Context, guildID ID) (IndexStatus, error)
	StartIndexingJob(ctx context.Context, guildID ID) error
	CompleteIndexingJob(ctx context.Context, guildID ID) error
	UpsertGuildMembers(ctx context.Context, guildID ID, memberIDs []MemberID, indexedAt time.Time) error
	DeleteGuildMembersNotIn(ctx context.Context, guildID ID, memberIDs []MemberID) error
	GetCharacterHoldersInGuild(ctx context.Context, guildID ID, charID catalog.CharacterID) ([]MemberID, error)
}

// TxQuerier is a Querier inside a transaction.
type TxQuerier interface {
	Querier
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}
