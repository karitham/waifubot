package guild

import (
	"context"
	"testing"

	"github.com/Karitham/corde"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/karitham/waifubot/storage/guildstore"
)

func TestIndexGuildIfNeeded_InMemory(t *testing.T) {
	store := NewInMemoryStore()
	ctx := context.Background()
	guildID := uint64(123)

	indexer := NewIndexer(store, "test_token")

	err := indexer.IndexGuildIfNeeded(ctx, corde.Snowflake(guildID))
	require.NoError(t, err)

	row, err := store.GuildStore().IsGuildIndexed(ctx, guildID)
	require.NoError(t, err)
	assert.Equal(t, guildstore.IndexingStatusPending, row.Status)

	err = store.GuildStore().CompleteIndexingJob(ctx, guildID)
	require.NoError(t, err)

	row, err = store.GuildStore().IsGuildIndexed(ctx, guildID)
	require.NoError(t, err)
	assert.Equal(t, guildstore.IndexingStatusCompleted, row.Status)
}

func TestIndexGuildIfNeeded_AlreadyIndexed_InMemory(t *testing.T) {
	store := NewInMemoryStore()
	ctx := context.Background()
	guildID := uint64(123)

	err := store.GuildStore().StartIndexingJob(ctx, guildID)
	require.NoError(t, err)
	err = store.GuildStore().CompleteIndexingJob(ctx, guildID)
	require.NoError(t, err)

	indexer := NewIndexer(store, "test_token")

	err = indexer.IndexGuildIfNeeded(ctx, corde.Snowflake(guildID))
	require.NoError(t, err)
}

func TestIndexGuildIfNeeded_StaleIndex_InMemory(t *testing.T) {
	store := NewInMemoryStore()
	ctx := context.Background()
	guildID := uint64(123)

	err := store.GuildStore().StartIndexingJob(ctx, guildID)
	require.NoError(t, err)
	err = store.GuildStore().CompleteIndexingJob(ctx, guildID)
	require.NoError(t, err)

	indexer := NewIndexer(store, "test_token")

	err = indexer.IndexGuildIfNeeded(ctx, corde.Snowflake(guildID))
	require.NoError(t, err)
}

func TestIndexGuild_Members_InMemory(t *testing.T) {
	store := NewInMemoryStore()
	ctx := context.Background()
	guildID := uint64(123)
	userIDs := []uint64{1, 2, 3, 4, 5}

	err := store.GuildStore().StartIndexingJob(ctx, guildID)
	require.NoError(t, err)

	intUserIDs := make([]int64, len(userIDs))
	for i, id := range userIDs {
		intUserIDs[i] = int64(id)
	}

	err = store.GuildStore().UpsertGuildMembers(ctx, guildstore.UpsertGuildMembersParams{
		GuildID: guildID,
		Column2: intUserIDs,
	})
	require.NoError(t, err)

	err = store.GuildStore().CompleteIndexingJob(ctx, guildID)
	require.NoError(t, err)

	members, err := store.GuildStore().GetGuildMembers(ctx, guildID)
	require.NoError(t, err)
	assert.Len(t, members, 5)
}

func TestIndexGuild_DeleteMembersNotIn_InMemory(t *testing.T) {
	store := NewInMemoryStore()
	ctx := context.Background()
	guildID := uint64(123)
	initialIDs := []int64{1, 2, 3, 4, 5}
	newIDs := []int64{2, 3, 6}

	err := store.GuildStore().UpsertGuildMembers(ctx, guildstore.UpsertGuildMembersParams{
		GuildID: guildID,
		Column2: initialIDs,
	})
	require.NoError(t, err)

	err = store.GuildStore().DeleteGuildMembersNotIn(ctx, guildstore.DeleteGuildMembersNotInParams{
		GuildID: guildID,
		Column2: newIDs,
	})
	require.NoError(t, err)

	members, err := store.GuildStore().GetGuildMembers(ctx, guildID)
	require.NoError(t, err)
	assert.Len(t, members, 2)
	assert.Contains(t, members, uint64(2))
	assert.Contains(t, members, uint64(3))
}
