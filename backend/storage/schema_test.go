//go:build integration

package storage_test

import (
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcexec "github.com/testcontainers/testcontainers-go/exec"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/karitham/waifubot/storage"
)

// TestSchemaSnapshotMatchesMigrations checks that schema.sql, the source
// sqlc generates from, still describes a database built by the migrations.
// When it fails after a postgres image bump, regenerate schema.sql with the
// container's pg_dump.
func TestSchemaSnapshotMatchesMigrations(t *testing.T) {
	ctx := t.Context()

	pgContainer, err := postgres.Run(ctx, "postgres:18-alpine",
		postgres.WithDatabase("waifubot_schema"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, testcontainers.TerminateContainer(pgContainer))
	})

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)
	require.NoError(t, storage.Migrate(connStr))

	code, out, err := pgContainer.Exec(ctx, []string{
		"pg_dump", "--schema-only", "--no-owner", "--no-privileges", "-U", "postgres", "waifubot_schema",
	}, tcexec.Multiplexed())
	require.NoError(t, err)
	require.Zero(t, code, "pg_dump failed")

	dump, err := io.ReadAll(out)
	require.NoError(t, err)

	committed, err := os.ReadFile("schema.sql")
	require.NoError(t, err)

	assert.Equal(t, normalizeSchema(string(committed)), normalizeSchema(string(dump)),
		"schema.sql drifted from the migrations; regenerate it with the container's pg_dump")
}

// normalizeSchema reduces a schema dump to its statement tokens so that
// layout and pg_dump scaffolding do not affect the comparison. The dbmate
// bookkeeping table is not part of the snapshot.
func normalizeSchema(s string) string {
	var tokens []string
	inMigrationsTable := false
	skipNext := false
	for _, line := range strings.Split(s, "\n") {
		l := strings.TrimSpace(line)
		switch {
		case skipNext:
			skipNext = false
			continue
		case strings.HasPrefix(l, "CREATE TABLE public.schema_migrations"):
			inMigrationsTable = true
			continue
		case inMigrationsTable:
			if strings.HasPrefix(l, ");") {
				inMigrationsTable = false
			}
			continue
		case l == "ALTER TABLE ONLY public.schema_migrations":
			skipNext = true
			continue
		case l == "" || strings.HasPrefix(l, "--") || strings.HasPrefix(l, "SET ") ||
			strings.HasPrefix(l, "SELECT pg_catalog") || strings.HasPrefix(l, "COMMENT ON") ||
			strings.HasPrefix(l, "\\"):
			continue
		}
		tokens = append(tokens, strings.Join(strings.Fields(l), " "))
	}
	return strings.Join(tokens, " ")
}
