package guild_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/Karitham/corde"

	"github.com/karitham/waifubot/collection"
	"github.com/karitham/waifubot/guild"
)

// mockQuerier implements guild.TxQuerier for testing.
type mockQuerier struct {
	status collection.GuildIndexStatus
	err    error
}

func (m *mockQuerier) IsGuildIndexed(ctx context.Context, guildID uint64) (collection.GuildIndexStatus, error) {
	return m.status, m.err
}
func (m *mockQuerier) StartIndexingJob(ctx context.Context, guildID uint64) error    { return nil }
func (m *mockQuerier) CompleteIndexingJob(ctx context.Context, guildID uint64) error { return nil }
func (m *mockQuerier) Commit(ctx context.Context) error                              { return nil }
func (m *mockQuerier) Rollback(ctx context.Context) error                            { return nil }
func (m *mockQuerier) UpsertGuildMembers(ctx context.Context, guildID uint64, memberIDs []uint64, indexedAt time.Time) error {
	return nil
}
func (m *mockQuerier) DeleteGuildMembersNotIn(ctx context.Context, guildID uint64, memberIDs []uint64) error {
	return nil
}

// mockFetcher implements guild.MemberFetcher for testing.
type mockFetcher struct {
	ids []corde.Snowflake
	err error
}

func (m *mockFetcher) FetchMemberIDs(ctx context.Context, guildID corde.Snowflake) ([]corde.Snowflake, error) {
	return m.ids, m.err
}

func TestIndexGuildIfNeeded_AlreadyIndexed(t *testing.T) {
	querier := &mockQuerier{
		status: collection.GuildIndexStatus{
			Status:    collection.IndexingCompleted,
			UpdatedAt: time.Now(),
		},
	}
	fetcher := &mockFetcher{}

	indexer := guild.NewIndexer(querier, fetcher)
	err := indexer.IndexGuildIfNeeded(t.Context(), corde.Snowflake(123), nil)
	require.NoError(t, err)
}

func TestIndexGuildIfNeeded_ExpiredIndex(t *testing.T) {
	querier := &mockQuerier{
		status: collection.GuildIndexStatus{
			Status:    collection.IndexingCompleted,
			UpdatedAt: time.Now().Add(-8 * 24 * time.Hour),
		},
	}

	fetcher := &mockFetcher{ids: []corde.Snowflake{100, 200}}

	indexer := guild.NewIndexer(querier, fetcher)

	txQuerier := &mockQuerier{
		status: collection.GuildIndexStatus{
			Status:    collection.IndexingInProgress,
			UpdatedAt: time.Now(),
		},
	}
	txFactory := func(ctx context.Context) (guild.TxQuerier, error) {
		return txQuerier, nil
	}

	err := indexer.IndexGuildIfNeeded(t.Context(), corde.Snowflake(123), txFactory)
	require.NoError(t, err)
}

func TestCharacterHolders_GuildNotIndexed(t *testing.T) {
	querier := &mockQuerier{
		status: collection.GuildIndexStatus{
			Status: collection.IndexingPending,
		},
	}

	_, _, err := guild.CharacterHolders(t.Context(), querier, nil, 123, 456)
	require.Error(t, err)
	require.Contains(t, err.Error(), "not indexed")
}

func TestCharacterHolders_GuildIDZero(t *testing.T) {
	_, _, err := guild.CharacterHolders(t.Context(), nil, nil, 0, 456)
	require.Error(t, err)
	require.Contains(t, err.Error(), "only be used in servers")
}
