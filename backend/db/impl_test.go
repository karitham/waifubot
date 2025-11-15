package db

import (
	"context"
	"net/url"
	"testing"
	"time"

	"github.com/Karitham/corde"
	"github.com/amacneil/dbmate/v2/pkg/dbmate"
	_ "github.com/amacneil/dbmate/v2/pkg/driver/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestStore_GuildOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("skip testing guild ops")
	}

	ctx := context.Background()

	pgContainer, connStr := setupPostgres(t, ctx)
	defer pgContainer.Terminate(ctx)

	store := setupStore(t, connStr)

	t.Run("GetGuildMembers empty", func(t *testing.T) {
		members, err := store.GetGuildMembers(ctx, 123)
		require.NoError(t, err)
		assert.Len(t, members, 0)
	})

	t.Run("InsertGuildMembers and GetGuildMembers", func(t *testing.T) {
		err := store.InsertGuildMembers(ctx, 123, []corde.Snowflake{456, 789})
		require.NoError(t, err)

		members, err := store.GetGuildMembers(ctx, 123)
		require.NoError(t, err)
		assert.Len(t, members, 2)
		assert.Contains(t, members, corde.Snowflake(456))
		assert.Contains(t, members, corde.Snowflake(789))
	})

	t.Run("UsersOwningCharInGuild", func(t *testing.T) {
		// Insert user and char
		_, err := store.db.Exec(ctx, "INSERT INTO users (user_id) VALUES ($1)", 456)
		require.NoError(t, err)
		_, err = store.db.Exec(ctx, "INSERT INTO characters (user_id, id, name, image, type) VALUES ($1, $2, $3, $4, $5)", 456, 1001, "Test Char", "image.png", "anime")
		require.NoError(t, err)

		owners, err := store.UsersOwningCharInGuild(ctx, 1001, 123)
		require.NoError(t, err)
		assert.Len(t, owners, 1)
		assert.Equal(t, corde.Snowflake(456), owners[0])
	})

	t.Run("IsGuildIndexed not indexed", func(t *testing.T) {
		indexed, err := store.IsGuildIndexed(ctx, 999, 7*24*time.Hour)
		require.NoError(t, err)
		assert.False(t, indexed)
	})

	t.Run("GetIndexingStatus no job", func(t *testing.T) {
		status, updatedAt, err := store.GetIndexingStatus(ctx, 999)
		require.NoError(t, err)
		assert.Equal(t, "pending", status)
		assert.True(t, updatedAt.IsZero())
	})

	t.Run("StartIndexingJob", func(t *testing.T) {
		err := store.StartIndexingJob(ctx, 123)
		require.NoError(t, err)

		status, updatedAt, err := store.GetIndexingStatus(ctx, 123)
		require.NoError(t, err)
		assert.Equal(t, "in_progress", status)
		assert.False(t, updatedAt.IsZero())
	})

	t.Run("CompleteIndexingJob", func(t *testing.T) {
		err := store.CompleteIndexingJob(ctx, 123)
		require.NoError(t, err)

		status, updatedAt, err := store.GetIndexingStatus(ctx, 123)
		require.NoError(t, err)
		assert.Equal(t, "completed", status)
		assert.False(t, updatedAt.IsZero())
	})

	t.Run("IsGuildIndexed completed", func(t *testing.T) {
		indexed, err := store.IsGuildIndexed(ctx, 123, 7*24*time.Hour)
		require.NoError(t, err)
		assert.True(t, indexed)
	})

	t.Run("InsertGuildMembers replace", func(t *testing.T) {
		err := store.InsertGuildMembers(ctx, 123, []corde.Snowflake{101112})
		require.NoError(t, err)

		members, err := store.GetGuildMembers(ctx, 123)
		require.NoError(t, err)
		assert.Len(t, members, 1)
		assert.Equal(t, corde.Snowflake(101112), members[0])
	})

	t.Run("InsertGuildMembers partial update", func(t *testing.T) {
		// Insert initial members
		err := store.InsertGuildMembers(ctx, 456, []corde.Snowflake{1, 2, 3})
		require.NoError(t, err)

		members, err := store.GetGuildMembers(ctx, 456)
		require.NoError(t, err)
		assert.Len(t, members, 3)
		assert.Contains(t, members, corde.Snowflake(1))
		assert.Contains(t, members, corde.Snowflake(2))
		assert.Contains(t, members, corde.Snowflake(3))

		// Update to [2, 3, 4]: should delete 1, keep 2 and 3 (update indexed_at), add 4
		err = store.InsertGuildMembers(ctx, 456, []corde.Snowflake{2, 3, 4})
		require.NoError(t, err)

		members, err = store.GetGuildMembers(ctx, 456)
		require.NoError(t, err)
		assert.Len(t, members, 3)
		assert.NotContains(t, members, corde.Snowflake(1)) // deleted
		assert.Contains(t, members, corde.Snowflake(2))    // kept
		assert.Contains(t, members, corde.Snowflake(3))    // kept
		assert.Contains(t, members, corde.Snowflake(4))    // added
	})
}

func setupPostgres(t *testing.T, ctx context.Context) (*postgres.PostgresContainer, string) {
	pgContainer, err := postgres.Run(ctx,
		"postgres:17-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	)
	require.NoError(t, err)

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	return pgContainer, connStr
}

func setupStore(t *testing.T, connStr string) *Store {
	u, err := url.Parse(connStr)
	require.NoError(t, err)

	db := dbmate.New(u)
	db.MigrationsDir = []string{"migrations"}
	db.SchemaFile = "schema.sql"

	err = db.CreateAndMigrate()
	require.NoError(t, err)

	store, err := NewStore(context.Background(), connStr)
	require.NoError(t, err)

	return store
}
