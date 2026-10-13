package interactionstore

import (
	"context"

	"github.com/Karitham/corde"

	"github.com/karitham/waifubot/storage"
)

type Store interface {
	Increment(ctx context.Context, channelID corde.Snowflake) error
	Get(ctx context.Context, channelID corde.Snowflake) (int64, error)
	Reset(ctx context.Context, channelID corde.Snowflake) error
}

type PostgresStore struct {
	q storage.Querier
}

func NewPostgresStore(q storage.Querier) *PostgresStore {
	return &PostgresStore{q: q}
}

func (p *PostgresStore) Increment(ctx context.Context, channelID corde.Snowflake) error {
	return p.q.Increment(ctx, uint64(channelID))
}

func (p *PostgresStore) Get(ctx context.Context, channelID corde.Snowflake) (int64, error) {
	count, err := p.q.GetInteraction(ctx, uint64(channelID))
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (p *PostgresStore) Reset(ctx context.Context, channelID corde.Snowflake) error {
	return p.q.Reset(ctx, uint64(channelID))
}
