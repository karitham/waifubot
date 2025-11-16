package dropstore

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/Karitham/corde"
	"github.com/fxamacker/cbor/v2"
	"github.com/redis/go-redis/v9"
)

type Store[T any] interface {
	Delete(ctx context.Context, id corde.Snowflake) error
	Get(ctx context.Context, id corde.Snowflake) (*T, error)
	Set(ctx context.Context, id corde.Snowflake, data T) error
}

type MemStore[T any] struct {
	mu     sync.Mutex
	values map[corde.Snowflake]T
}

func NewMemStore[T any]() *MemStore[T] {
	return &MemStore[T]{
		values: map[corde.Snowflake]T{},
	}
}

func (m *MemStore[T]) Delete(ctx context.Context, id corde.Snowflake) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.values, id)

	return nil
}

func (m *MemStore[T]) Get(ctx context.Context, id corde.Snowflake) (*T, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	v := m.values[id]
	return &v, nil
}

func (m *MemStore[T]) Set(ctx context.Context, id corde.Snowflake, data T) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.values[id] = data

	return nil
}

type RedisStore[T any] struct {
	client         redis.UniversalClient
	prefix, suffix string
}

func NewRedis[T any](c redis.UniversalClient, prefix string, suffix string) Store[T] {
	return RedisStore[T]{
		client: c,
		prefix: prefix,
		suffix: suffix,
	}
}

func (r RedisStore[T]) Set(ctx context.Context, id corde.Snowflake, data T) error {
	b, err := cbor.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal char: %w", err)
	}

	return r.client.Set(ctx, redisKey(id, r.prefix, r.suffix), string(b), 0).Err()
}

func (r RedisStore[T]) Get(ctx context.Context, id corde.Snowflake) (*T, error) {
	s, err := r.client.Get(ctx, redisKey(id, r.prefix, r.suffix)).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get channel char: %w", err)
	}

	var data T
	if err := cbor.Unmarshal([]byte(s), &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal char: %w", err)
	}

	return &data, nil
}

func (r RedisStore[T]) Delete(ctx context.Context, id corde.Snowflake) error {
	return r.client.Del(ctx, redisKey(id, r.prefix, r.suffix)).Err()
}

func redisKey(id corde.Snowflake, prefix, suffix string) string {
	b := strings.Builder{}
	if prefix != "" {
		b.WriteString(prefix)
		b.WriteByte(':')
	}

	b.WriteString(id.String())

	if suffix != "" {
		b.WriteByte(':')
		b.WriteString(suffix)
	}

	return b.String()
}
