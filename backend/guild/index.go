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

	started, err := startIndexingJobIfNeeded(ctx, i.db, id)
	if err != nil {
		return err
	}
	if !started {
		return nil
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
// job unless another indexer is currently active. It reports whether this
// caller owns the job and may run the index.
func startIndexingJobIfNeeded(ctx context.Context, db DB, guildID ID) (bool, error) {
	tx, err := db.Begin(ctx)
	if err != nil {
		return false, err
	}

	// Re-check inside the transaction: a fresh in-progress job belongs to
	// another indexer, and running a second one would duplicate the fetch.
	status, err := tx.IsGuildIndexed(ctx, guildID)
	if err == nil && status.Status == IndexingInProgress && time.Since(status.UpdatedAt) < 10*time.Minute {
		_ = tx.Rollback(ctx)
		return false, nil
	}

	if err := tx.StartIndexingJob(ctx, guildID); err != nil {
		_ = tx.Rollback(ctx)
		return false, err
	}

	if err := tx.Commit(ctx); err != nil {
		return false, err
	}
	return true, nil
}
