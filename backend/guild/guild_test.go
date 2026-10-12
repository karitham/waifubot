package guild_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/Karitham/corde"

	"github.com/karitham/waifubot/catalog"
	"github.com/karitham/waifubot/guild"
)

// mockQuerier implements guild.DB for testing.
type mockQuerier struct {
	status guild.IndexStatus
	err    error
	tx     guild.TxQuerier
}

func (m *mockQuerier) Begin(ctx context.Context) (guild.TxQuerier, error) {
	if m.tx != nil {
		return m.tx, nil
	}
	return m, nil
}

func (m *mockQuerier) IsGuildIndexed(ctx context.Context, guildID guild.ID) (guild.IndexStatus, error) {
	return m.status, m.err
}

func (m *mockQuerier) StartIndexingJob(ctx context.Context, guildID guild.ID) error { return nil }

func (m *mockQuerier) CompleteIndexingJob(ctx context.Context, guildID guild.ID) error {
	return nil
}
func (m *mockQuerier) Commit(ctx context.Context) error   { return nil }
func (m *mockQuerier) Rollback(ctx context.Context) error { return nil }
func (m *mockQuerier) UpsertGuildMembers(ctx context.Context, guildID guild.ID, memberIDs []guild.MemberID, indexedAt time.Time) error {
	return nil
}

func (m *mockQuerier) DeleteGuildMembersNotIn(ctx context.Context, guildID guild.ID, memberIDs []guild.MemberID) error {
	return nil
}

func (m *mockQuerier) GetCharacterHoldersInGuild(ctx context.Context, guildID guild.ID, charID catalog.CharacterID) ([]guild.MemberID, error) {
	return nil, nil
}

// mockFetcher implements guild.MemberFetcher for testing.
type mockFetcher struct {
	ids    []corde.Snowflake
	err    error
	called bool
}

func (m *mockFetcher) FetchMemberIDs(ctx context.Context, guildID corde.Snowflake) ([]corde.Snowflake, error) {
	m.called = true
	return m.ids, m.err
}

func TestIndexGuildIfNeeded_AlreadyIndexed(t *testing.T) {
	querier := &mockQuerier{
		status: guild.IndexStatus{
			Status:    guild.IndexingCompleted,
			UpdatedAt: time.Now(),
		},
	}
	fetcher := &mockFetcher{}

	indexer := guild.NewIndexer(querier, fetcher)
	err := indexer.IndexGuildIfNeeded(t.Context(), corde.Snowflake(123))
	require.NoError(t, err)
}

func TestIndexGuildIfNeeded_AnotherIndexerActive(t *testing.T) {
	txQuerier := &mockQuerier{
		status: guild.IndexStatus{
			Status:    guild.IndexingInProgress,
			UpdatedAt: time.Now(),
		},
	}
	querier := &mockQuerier{
		status: guild.IndexStatus{
			Status:    guild.IndexingCompleted,
			UpdatedAt: time.Now().Add(-8 * 24 * time.Hour),
		},
		tx: txQuerier,
	}

	fetcher := &mockFetcher{ids: []corde.Snowflake{100, 200}}

	indexer := guild.NewIndexer(querier, fetcher)
	err := indexer.IndexGuildIfNeeded(t.Context(), corde.Snowflake(123))
	require.NoError(t, err)
	require.False(t, fetcher.called, "a fresh in-progress job must not be indexed again")
}

func TestIndexGuildIfNeeded_ExpiredIndex(t *testing.T) {
	txQuerier := &mockQuerier{
		status: guild.IndexStatus{
			Status:    guild.IndexingInProgress,
			UpdatedAt: time.Now().Add(-11 * time.Minute),
		},
	}
	querier := &mockQuerier{
		status: guild.IndexStatus{
			Status:    guild.IndexingCompleted,
			UpdatedAt: time.Now().Add(-8 * 24 * time.Hour),
		},
		tx: txQuerier,
	}

	fetcher := &mockFetcher{ids: []corde.Snowflake{100, 200}}

	indexer := guild.NewIndexer(querier, fetcher)
	err := indexer.IndexGuildIfNeeded(t.Context(), corde.Snowflake(123))
	require.NoError(t, err)
	require.True(t, fetcher.called, "a stale in-progress job must be taken over")
}

func TestCharacterHolders_GuildNotIndexed(t *testing.T) {
	querier := &mockQuerier{
		status: guild.IndexStatus{
			Status: guild.IndexingPending,
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
