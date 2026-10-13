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
	tx     guild.Tx

	deleteCalls   int
	upsertCalls   int
	completeCalls int
}

func (m *mockQuerier) Begin(ctx context.Context) (guild.Tx, error) {
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
	m.completeCalls++
	return nil
}
func (m *mockQuerier) Commit(ctx context.Context) error   { return nil }
func (m *mockQuerier) Rollback(ctx context.Context) error { return nil }
func (m *mockQuerier) UpsertGuildMembers(ctx context.Context, guildID guild.ID, memberIDs []guild.MemberID, indexedAt time.Time) error {
	m.upsertCalls++
	return nil
}

func (m *mockQuerier) DeleteGuildMembersNotIn(ctx context.Context, guildID guild.ID, memberIDs []guild.MemberID) error {
	m.deleteCalls++
	return nil
}

func (m *mockQuerier) GetCharacterHoldersInGuild(ctx context.Context, guildID guild.ID, charID catalog.CharacterID) ([]guild.MemberID, error) {
	return nil, nil
}

func (m *mockQuerier) GetCharacterByID(ctx context.Context, charID catalog.CharacterID) (catalog.Character, error) {
	return catalog.Character{}, nil
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
	require.Zero(t, txQuerier.deleteCalls, "a skipped run must not touch members")
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
	require.Equal(t, 1, txQuerier.deleteCalls, "the member sync must go through the transaction")
	require.Equal(t, 1, txQuerier.upsertCalls)
	require.Equal(t, 1, txQuerier.completeCalls)
}

func TestCharacterHolders_GuildNotIndexed(t *testing.T) {
	querier := &mockQuerier{
		status: guild.IndexStatus{
			Status: guild.IndexingPending,
		},
	}

	_, _, err := guild.CharacterHolders(t.Context(), querier, 123, 456)
	require.ErrorIs(t, err, guild.ErrNotIndexed)
}

func TestCharacterHolders_GuildIDZero(t *testing.T) {
	_, _, err := guild.CharacterHolders(t.Context(), nil, 0, 456)
	require.ErrorIs(t, err, guild.ErrNotInGuild)
}
