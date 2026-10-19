package discord

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Karitham/corde"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/karitham/waifubot/collection"
	"github.com/karitham/waifubot/collection/collectiontest"
)

func staleUserStore() *collectiontest.MockStore {
	return &collectiontest.MockStore{
		GetUserFunc: func(_ context.Context, userID collection.UserID) (collection.User, error) {
			return collection.User{
				UserID:      userID,
				LastUpdated: time.Now().Add(-48 * time.Hour),
			}, nil
		},
	}
}

func TestUpdateIfNeededDedupesConcurrentReaders(t *testing.T) {
	var fetches atomic.Int32
	started := make(chan struct{})
	release := make(chan struct{})

	store := staleUserStore()
	service := newService(func(_ context.Context, userID string) (*DiscordUser, error) {
		if fetches.Add(1) == 1 {
			close(started)
		}
		<-release
		return &DiscordUser{ID: userID, Username: "user", Avatar: "hash"}, nil
	})

	leaderDone := make(chan error, 1)
	go func() {
		leaderDone <- service.UpdateIfNeeded(t.Context(), store, 1)
	}()
	<-started

	followerDone := make(chan error, 1)
	go func() {
		followerDone <- service.UpdateIfNeeded(t.Context(), store, 1)
	}()

	// The leader is already blocked in fetch; give the follower time to join
	// the in-flight call before releasing it.
	time.Sleep(50 * time.Millisecond)
	close(release)

	require.NoError(t, <-leaderDone)
	require.NoError(t, <-followerDone)
	assert.Equal(t, int32(1), fetches.Load(), "concurrent readers must share one fetch")
}

func TestUpdateIfNeededCachesFailures(t *testing.T) {
	var fetches atomic.Int32
	store := staleUserStore()
	service := newService(func(context.Context, string) (*DiscordUser, error) {
		fetches.Add(1)
		return nil, errors.New("discord unavailable")
	})

	require.Error(t, service.UpdateIfNeeded(t.Context(), store, 1))
	require.NoError(t, service.UpdateIfNeeded(t.Context(), store, 1))

	assert.Equal(t, int32(1), fetches.Load(), "the failure cache must suppress the second fetch")
}

func TestUpdateIfNeededSkipsFreshUser(t *testing.T) {
	var fetches atomic.Int32
	store := &collectiontest.MockStore{
		GetUserFunc: func(_ context.Context, userID collection.UserID) (collection.User, error) {
			return collection.User{UserID: userID, LastUpdated: time.Now()}, nil
		},
	}
	service := newService(func(context.Context, string) (*DiscordUser, error) {
		fetches.Add(1)
		return &DiscordUser{}, nil
	})

	require.NoError(t, service.UpdateIfNeeded(t.Context(), store, 1))
	assert.Zero(t, fetches.Load())
}

func TestUpdateIfNeededSkipsMissingUser(t *testing.T) {
	var fetches atomic.Int32
	store := &collectiontest.MockStore{
		GetUserFunc: func(context.Context, collection.UserID) (collection.User, error) {
			return collection.User{}, collection.ErrNotFound
		},
	}
	service := newService(func(context.Context, string) (*DiscordUser, error) {
		fetches.Add(1)
		return &DiscordUser{}, nil
	})

	require.NoError(t, service.UpdateIfNeeded(t.Context(), store, 1))
	assert.Zero(t, fetches.Load())
}

func TestUpdateIfNeededSurvivesCallerCancellation(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})
	store := staleUserStore()
	service := newService(func(_ context.Context, userID string) (*DiscordUser, error) {
		close(started)
		<-release
		return &DiscordUser{ID: userID, Username: "user", Avatar: "hash"}, nil
	})

	ctx, cancel := context.WithCancel(t.Context())
	errCh := make(chan error, 1)
	go func() {
		errCh <- service.UpdateIfNeeded(ctx, store, 1)
	}()

	<-started
	cancel()
	close(release)

	require.NoError(t, <-errCh)
}

func TestUpdateIfNeededRejectsUserZero(t *testing.T) {
	service := newService(func(context.Context, string) (*DiscordUser, error) {
		t.Fatal("user zero must not trigger a fetch")
		return nil, nil
	})

	err := service.UpdateIfNeeded(t.Context(), staleUserStore(), corde.Snowflake(0))
	require.ErrorIs(t, err, collection.ErrInvalidUserID)
}
