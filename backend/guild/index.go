package guild

import (
	"context"
	"time"

	"github.com/Karitham/corde"
)

// MemberFetcher abstracts Discord guild member fetching.
type MemberFetcher interface {
	FetchMemberIDs(ctx context.Context, guildID corde.Snowflake) ([]corde.Snowflake, error)
}

const maxAge = 7 * 24 * time.Hour

// Indexer handles guild indexing operations.
type Indexer struct {
	db      DB
	fetcher MemberFetcher
}

// NewIndexer creates a new guild indexer.
func NewIndexer(db DB, fetcher MemberFetcher) *Indexer {
	return &Indexer{
		db:      db,
		fetcher: fetcher,
	}
}

// IndexGuildIfNeeded checks if a guild needs indexing and runs it if necessary.
func (i *Indexer) IndexGuildIfNeeded(ctx context.Context, guildID corde.Snowflake) error {
	id := ID(guildID)
	status, err := i.db.IsGuildIndexed(ctx, id)
	if err == nil && status.Status == IndexingCompleted && time.Since(status.UpdatedAt) <= maxAge {
		return nil
	}

	if err := startIndexingJobIfNeeded(ctx, i.db, id); err != nil {
		return err
	}

	return i.IndexGuild(ctx, guildID)
}

// IndexGuild fetches members and syncs the database.
func (i *Indexer) IndexGuild(ctx context.Context, guildID corde.Snowflake) error {
	memberIDs, err := i.fetcher.FetchMemberIDs(ctx, guildID)
	if err != nil {
		return err
	}

	ids := make([]MemberID, len(memberIDs))
	for idx, id := range memberIDs {
		ids[idx] = MemberID(id)
	}

	if err := i.db.DeleteGuildMembersNotIn(ctx, ID(guildID), ids); err != nil {
		return err
	}

	if len(ids) > 0 {
		if err := i.db.UpsertGuildMembers(ctx, ID(guildID), ids, time.Now()); err != nil {
			return err
		}
	}

	return i.db.CompleteIndexingJob(ctx, ID(guildID))
}

// startIndexingJobIfNeeded acquires a transaction lock and starts an indexing
// job if no other indexer is currently active.
func startIndexingJobIfNeeded(ctx context.Context, db DB, guildID ID) error {
	tx, err := db.Begin(ctx)
	if err != nil {
		return err
	}

	// Re-check inside transaction to prevent duplicate indexing
	status, err := tx.IsGuildIndexed(ctx, guildID)
	if err == nil && status.Status == IndexingInProgress && time.Since(status.UpdatedAt) < 10*time.Minute {
		_ = tx.Rollback(ctx)
		return nil
	}

	if err := tx.StartIndexingJob(ctx, guildID); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}

	return tx.Commit(ctx)
}
