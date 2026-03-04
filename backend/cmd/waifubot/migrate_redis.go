package main

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/Karitham/corde"
	"github.com/fxamacker/cbor/v2"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/redis/go-redis/v9"
	"github.com/urfave/cli/v2"

	"github.com/karitham/waifubot/collection"
	"github.com/karitham/waifubot/storage"
	"github.com/karitham/waifubot/storage/dropstore"
)

type InteractionDB interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

type RedisClient interface {
	Scan(ctx context.Context, cursor uint64, match string, count int64) *redis.ScanCmd
	Get(ctx context.Context, key string) *redis.StringCmd
}

type Migrator struct {
	Redis     RedisClient
	DropStore dropstore.Store
	DB        InteractionDB
}

var MigrateRedisCommand = &cli.Command{
	Name:  "migrate-redis",
	Usage: "Migrates data from Redis to PostgreSQL (one-time operation)",
	Flags: []cli.Flag{
		dbURLFlag,
		&cli.StringFlag{
			Name:     "redis-url",
			EnvVars:  []string{"REDIS_URL"},
			Required: true,
		},
	},
	Action: func(c *cli.Context) error {
		ctx := c.Context

		redisURL := c.String("redis-url")
		opts, err := redis.ParseURL(redisURL)
		if err != nil {
			return fmt.Errorf("error parsing redis url: %w", err)
		}
		rdb := redis.NewClient(opts)
		defer rdb.Close()

		if err := rdb.Ping(ctx).Err(); err != nil {
			return fmt.Errorf("cannot connect to redis: %w", err)
		}

		store, err := storage.NewStore(ctx, c.String(dbURLFlag.Name))
		if err != nil {
			return fmt.Errorf("error connecting to postgres db: %w", err)
		}

		m := &Migrator{
			Redis:     rdb,
			DropStore: dropstore.NewPostgresStore(store.DropStore()),
			DB:        store.DB(),
		}

		if err := m.MigrateInteractions(ctx); err != nil {
			return err
		}

		return m.MigrateDrops(ctx)
	},
}

func (m *Migrator) MigrateInteractions(ctx context.Context) error {
	slog.Info("Starting interaction migration")
	iter := m.Redis.Scan(ctx, 0, "channel:*:interactions", 0).Iterator()
	var count int
	for iter.Next(ctx) {
		if err := m.migrateInteraction(ctx, iter.Val()); err != nil {
			slog.Error("failed to migrate interaction", "err", err)
			continue
		}
		count++
	}
	if err := iter.Err(); err != nil {
		return fmt.Errorf("error iterating interaction keys: %w", err)
	}
	slog.Info("Interaction migration complete", "count", count)
	return nil
}

func (m *Migrator) migrateInteraction(ctx context.Context, key string) error {
	parts := strings.Split(key, ":")
	if len(parts) != 3 {
		return nil
	}

	idStr := parts[1]
	count, err := m.Redis.Get(ctx, key).Int64()
	if err != nil {
		return fmt.Errorf("failed to get count for key %s: %w", key, err)
	}

	sf := corde.SnowflakeFromString(idStr)
	_, err = m.DB.Exec(ctx,
		`INSERT INTO channel_interactions (channel_id, interaction_count)
		 VALUES ($1, $2)
		 ON CONFLICT (channel_id) DO UPDATE SET interaction_count = EXCLUDED.interaction_count`,
		uint64(sf), count,
	)
	if err != nil {
		return fmt.Errorf("failed to migrate interaction for channel %s: %w", sf, err)
	}
	return nil
}

func (m *Migrator) MigrateDrops(ctx context.Context) error {
	slog.Info("Starting drops migration")
	iter := m.Redis.Scan(ctx, 0, "channel:*:char", 0).Iterator()
	var count int
	for iter.Next(ctx) {
		if err := m.migrateDrop(ctx, iter.Val()); err != nil {
			slog.Error("failed to migrate drop", "err", err)
			continue
		}
		count++
	}
	if err := iter.Err(); err != nil {
		return fmt.Errorf("error iterating drop keys: %w", err)
	}
	slog.Info("Drops migration complete", "count", count)
	return nil
}

func (m *Migrator) migrateDrop(ctx context.Context, key string) error {
	parts := strings.Split(key, ":")
	if len(parts) != 3 {
		return nil
	}

	idStr := parts[1]
	s, err := m.Redis.Get(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("failed to get drop for key %s: %w", key, err)
	}

	var data collection.MediaCharacter
	if err := cbor.Unmarshal([]byte(s), &data); err != nil {
		return fmt.Errorf("failed to unmarshal cbor drop for key %s: %w", key, err)
	}

	sf := corde.SnowflakeFromString(idStr)
	err = m.DropStore.Set(ctx, sf, dropstore.Drop{
		ID:         data.ID,
		Name:       data.Name,
		ImageURL:   data.ImageURL,
		MediaTitle: data.MediaTitle,
	})
	if err != nil {
		return fmt.Errorf("failed to migrate drop for channel %s: %w", sf, err)
	}
	return nil
}
