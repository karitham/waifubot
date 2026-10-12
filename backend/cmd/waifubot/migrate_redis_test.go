package main

import (
	"context"
	"strconv"
	"testing"

	"github.com/fxamacker/cbor/v2"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"

	"github.com/karitham/waifubot/catalog"
	"github.com/karitham/waifubot/collection"
	"github.com/karitham/waifubot/storage/dropstore"
)

type mockRedis struct {
	RedisClient
	scanFunc func(ctx context.Context, cursor uint64, match string, count int64) *redis.ScanCmd
	getFunc  func(ctx context.Context, key string) *redis.StringCmd
}

func (m *mockRedis) Scan(ctx context.Context, cursor uint64, match string, count int64) *redis.ScanCmd {
	return m.scanFunc(ctx, cursor, match, count)
}

func (m *mockRedis) Get(ctx context.Context, key string) *redis.StringCmd {
	return m.getFunc(ctx, key)
}

type mockDB struct {
	execFunc func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

func (m *mockDB) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	return m.execFunc(ctx, sql, args...)
}

type mockDropStore struct {
	dropstore.Store
	setFunc func(ctx context.Context, id collection.ChannelID, data catalog.Character) error
}

func (m *mockDropStore) Set(ctx context.Context, id collection.ChannelID, data catalog.Character) error {
	return m.setFunc(ctx, id, data)
}

func TestMigrateInteractions(t *testing.T) {
	ctx := t.Context()
	keys := []string{"channel:123:interactions", "channel:456:interactions"}
	counts := map[string]int64{
		"channel:123:interactions": 10,
		"channel:456:interactions": 20,
	}

	mr := &mockRedis{
		scanFunc: func(ctx context.Context, cursor uint64, match string, count int64) *redis.ScanCmd {
			cmd := redis.NewScanCmd(ctx, nil)
			cmd.SetVal(keys, 0)
			return cmd
		},
		getFunc: func(ctx context.Context, key string) *redis.StringCmd {
			cmd := redis.NewStringCmd(ctx)
			cmd.SetVal(strconv.FormatInt(counts[key], 10))
			return cmd
		},
	}

	executedQueries := []struct {
		sql  string
		args []any
	}{}
	mdb := &mockDB{
		execFunc: func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
			executedQueries = append(executedQueries, struct {
				sql  string
				args []any
			}{sql, args})
			return pgconn.NewCommandTag("INSERT 1"), nil
		},
	}

	m := &Migrator{
		Redis: mr,
		DB:    mdb,
	}

	err := m.MigrateInteractions(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(executedQueries))

	assert.Equal(t, uint64(123), executedQueries[0].args[0])
	assert.Equal(t, int64(10), executedQueries[0].args[1])
	assert.Equal(t, uint64(456), executedQueries[1].args[0])
	assert.Equal(t, int64(20), executedQueries[1].args[1])
}

func TestMigrateDrops(t *testing.T) {
	ctx := t.Context()
	keys := []string{"channel:123:char"}

	// The payload was written before the Popularity rename, so the key
	// must stay Favorites.
	legacy := struct {
		ID         int64
		Name       string
		ImageURL   string
		MediaTitle string
		Favorites  int
	}{
		ID:         1,
		Name:       "Test Char",
		ImageURL:   "http://example.com/img.png",
		MediaTitle: "Media",
		Favorites:  4242,
	}
	data, _ := cbor.Marshal(legacy)

	mr := &mockRedis{
		scanFunc: func(ctx context.Context, cursor uint64, match string, count int64) *redis.ScanCmd {
			cmd := redis.NewScanCmd(ctx, nil)
			cmd.SetVal(keys, 0)
			return cmd
		},
		getFunc: func(ctx context.Context, key string) *redis.StringCmd {
			cmd := redis.NewStringCmd(ctx)
			cmd.SetVal(string(data))
			return cmd
		},
	}

	var capturedDrop catalog.Character
	ds := &mockDropStore{
		setFunc: func(ctx context.Context, id collection.ChannelID, data catalog.Character) error {
			capturedDrop = data
			return nil
		},
	}

	m := &Migrator{
		Redis:     mr,
		DropStore: ds,
		DB:        &mockDB{},
	}

	err := m.MigrateDrops(ctx)
	assert.NoError(t, err)

	assert.Equal(t, int64(1), int64(capturedDrop.ID))
	assert.Equal(t, "Test Char", capturedDrop.Name)
	assert.Equal(t, 4242, capturedDrop.Popularity)
}
