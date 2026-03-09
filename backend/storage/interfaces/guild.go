package interfaces

import (
	"context"
	"time"
)

type IndexingStatus string

const (
	IndexingStatusPending   IndexingStatus = "pending"
	IndexingStatusRunning   IndexingStatus = "running"
	IndexingStatusCompleted IndexingStatus = "completed"
	IndexingStatusFailed    IndexingStatus = "failed"
)

type GuildRepository interface {
	CompleteIndexingJob(ctx context.Context, guildID uint64) error
	DeleteGuildMembers(ctx context.Context, guildID uint64) error
	DeleteGuildMembersNotIn(ctx context.Context, guildID uint64, memberIDs []uint64) error
	GetGuildMembers(ctx context.Context, guildID uint64) ([]uint64, error)
	GetIndexingStatus(ctx context.Context, guildID uint64) (IndexStatus, error)
	IsGuildIndexed(ctx context.Context, guildID uint64) (bool, time.Time, error)
	StartIndexingJob(ctx context.Context, guildID uint64) error
	UpsertGuildMembers(ctx context.Context, guildID uint64, memberIDs []uint64) error
	UsersOwningCharInGuild(ctx context.Context, guildID uint64, charID int64) ([]uint64, error)
}

type IndexStatus struct {
	Status    IndexingStatus
	UpdatedAt time.Time
}
