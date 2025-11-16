package interactionstore

import (
	"context"
	"sync"

	"github.com/Karitham/corde"
	"github.com/redis/go-redis/v9"
)

type Store interface {
	Increment(ctx context.Context, channelID corde.Snowflake) error
	Get(ctx context.Context, channelID corde.Snowflake) (int64, error)
	Reset(ctx context.Context, channelID corde.Snowflake) error
}

type MemStore struct {
	mu     sync.Mutex
	values map[corde.Snowflake]int64
}

func NewMemStore() *MemStore {
	return &MemStore{
		values: map[corde.Snowflake]int64{},
	}
}

func (m *MemStore) Increment(ctx context.Context, channelID corde.Snowflake) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.values[channelID]++

	return nil
}

func (m *MemStore) Get(ctx context.Context, channelID corde.Snowflake) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.values[channelID], nil
}

func (m *MemStore) Reset(ctx context.Context, channelID corde.Snowflake) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.values, channelID)

	return nil
}

type RedisStore struct {
	client redis.UniversalClient
}

func NewRedis(c redis.UniversalClient) RedisStore {
	return RedisStore{
		client: c,
	}
}

func (r RedisStore) Increment(ctx context.Context, channelID corde.Snowflake) error {
	key := "channel:" + channelID.String() + ":interactions"
	return r.client.Incr(ctx, key).Err()
}

func (r RedisStore) Get(ctx context.Context, channelID corde.Snowflake) (int64, error) {
	key := "channel:" + channelID.String() + ":interactions"
	return r.client.Get(ctx, key).Int64()
}

func (r RedisStore) Reset(ctx context.Context, channelID corde.Snowflake) error {
	key := "channel:" + channelID.String() + ":interactions"
	return r.client.Del(ctx, key).Err()
}
