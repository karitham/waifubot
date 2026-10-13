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

// Reader reads guild indexing state and holder lookups.
type Reader interface {
	IsGuildIndexed(ctx context.Context, guildID ID) (IndexStatus, error)
	GetCharacterHoldersInGuild(ctx context.Context, guildID ID, charID catalog.CharacterID) ([]MemberID, error)
}

// Writer holds the indexing writes; all run inside a transaction.
type Writer interface {
	StartIndexingJob(ctx context.Context, guildID ID) error
	CompleteIndexingJob(ctx context.Context, guildID ID) error
	UpsertGuildMembers(ctx context.Context, guildID ID, memberIDs []MemberID, indexedAt time.Time) error
	DeleteGuildMembersNotIn(ctx context.Context, guildID ID, memberIDs []MemberID) error
}

// TxQuerier is the full guild surface inside a transaction.
type TxQuerier interface {
	Reader
	Writer
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

// DB is the guild read surface plus transaction creation.
type DB interface {
	Reader
	Begin(ctx context.Context) (TxQuerier, error)
}
