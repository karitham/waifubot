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
	store   GuildQuerier
	fetcher MemberFetcher
}

// NewIndexer creates a new guild indexer.
// Store must be a GuildQuerier. If it's also a Transactional, pass WithTx as the tx factory.
func NewIndexer(store GuildQuerier, fetcher MemberFetcher) *Indexer {
	return &Indexer{
		store:   store,
		fetcher: fetcher,
	}
}

// IndexGuildIfNeeded checks if a guild needs indexing and runs it if necessary.
func (i *Indexer) IndexGuildIfNeeded(ctx context.Context, guildID corde.Snowflake, txFactory func(context.Context) (TxQuerier, error)) error {
	id := GuildID(guildID)
	status, err := i.store.IsGuildIndexed(ctx, id)
	if err == nil && status.Status == IndexingCompleted && time.Since(status.UpdatedAt) <= maxAge {
		return nil
	}

	if err := StartIndexingJobIfNeeded(ctx, i.store, id, txFactory); err != nil {
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

	if err := i.store.DeleteGuildMembersNotIn(ctx, GuildID(guildID), ids); err != nil {
		return err
	}

	if len(ids) > 0 {
		if err := i.store.UpsertGuildMembers(ctx, GuildID(guildID), ids, time.Now()); err != nil {
			return err
		}
	}

	return i.store.CompleteIndexingJob(ctx, GuildID(guildID))
}

// StartIndexingJobIfNeeded acquires a transaction lock and starts an indexing job
// if no other indexer is currently active.
func StartIndexingJobIfNeeded(ctx context.Context, store GuildQuerier, guildID GuildID, txFactory func(context.Context) (TxQuerier, error)) error {
	tx, err := txFactory(ctx)
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
