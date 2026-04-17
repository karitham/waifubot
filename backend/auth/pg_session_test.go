//go:build integration

package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/karitham/waifubot/auth"
	"github.com/karitham/waifubot/storage"
)

var testDBURL string

func TestMain(m *testing.M) {
	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx, "postgres:17-alpine",
		postgres.WithDatabase("waifubot_test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		panic("failed to start postgres container: " + err.Error())
	}
	defer func() {
		if err := testcontainers.TerminateContainer(pgContainer); err != nil {
			panic("failed to terminate container: " + err.Error())
		}
	}()

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		panic("failed to get connection string: " + err.Error())
	}
	testDBURL = connStr

	if err := storage.Migrate(testDBURL); err != nil {
		panic("migration failed: " + err.Error())
	}

	m.Run()
}

func setupStore(t *testing.T) (auth.SessionStore, storage.Store) {
	t.Helper()

	dbStore, err := storage.NewStore(t.Context(), testDBURL)
	require.NoError(t, err)

	txStore, err := dbStore.Tx(t.Context())
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = txStore.Rollback(t.Context())
	})

	return auth.NewPgSessionStore(txStore.SessionStore()), txStore
}

func seedUser(t *testing.T, s storage.Store, userID uint64) {
	t.Helper()
	require.NoError(t, s.UserStore().Create(t.Context(), userID))
}

func TestIntegration_CreateAndGetSession(t *testing.T) {
	const uid uint64 = 800001
	store, s := setupStore(t)
	seedUser(t, s, uid)
	ctx := t.Context()

	session, err := store.CreateSession(ctx, uid)
	require.NoError(t, err)
	assert.Len(t, session.Token, 64)
	assert.Equal(t, uid, session.UserID)
	assert.WithinDuration(t, time.Now(), session.CreatedAt, 5*time.Second)
	assert.WithinDuration(t, time.Now().Add(7*24*time.Hour), session.ExpiresAt, 5*time.Second)

	got, err := store.GetSession(ctx, session.Token)
	require.NoError(t, err)
	assert.Equal(t, session.Token, got.Token)
	assert.Equal(t, session.UserID, got.UserID)
	assert.WithinDuration(t, session.CreatedAt, got.CreatedAt, time.Second)
	assert.WithinDuration(t, session.ExpiresAt, got.ExpiresAt, time.Second)
}

func TestIntegration_GetExpiredSession(t *testing.T) {
	const uid uint64 = 800002
	store, s := setupStore(t)
	seedUser(t, s, uid)
	ctx := t.Context()

	session, err := store.CreateSession(ctx, uid)
	require.NoError(t, err)

	// Manually expire the session by updating expires_at
	_, err = s.DB().Exec(ctx, "UPDATE sessions SET expires_at = NOW() - INTERVAL '1 hour' WHERE token = $1", session.Token)
	require.NoError(t, err)

	_, err = store.GetSession(ctx, session.Token)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "session expired")
}

func TestIntegration_DeleteSession(t *testing.T) {
	const uid uint64 = 800003
	store, s := setupStore(t)
	seedUser(t, s, uid)
	ctx := t.Context()

	session, err := store.CreateSession(ctx, uid)
	require.NoError(t, err)

	require.NoError(t, store.DeleteSession(ctx, session.Token))

	_, err = store.GetSession(ctx, session.Token)
	require.Error(t, err)
}

func TestIntegration_DeleteExpiredSessions(t *testing.T) {
	const uid uint64 = 800004
	store, s := setupStore(t)
	seedUser(t, s, uid)
	ctx := t.Context()

	active, err := store.CreateSession(ctx, uid)
	require.NoError(t, err)

	// Create a second session and expire it
	expired, err := store.CreateSession(ctx, uid)
	require.NoError(t, err)
	_, err = s.DB().Exec(ctx, "UPDATE sessions SET expires_at = NOW() - INTERVAL '1 hour' WHERE token = $1", expired.Token)
	require.NoError(t, err)

	require.NoError(t, store.DeleteExpiredSessions(ctx))

	// Active session should still be retrievable
	_, err = store.GetSession(ctx, active.Token)
	require.NoError(t, err)

	// Expired session should be gone from DB
	_, err = store.GetSession(ctx, expired.Token)
	require.Error(t, err)
}
