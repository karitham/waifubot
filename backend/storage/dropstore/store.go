package dropstore

import (
	"context"
	"fmt"

	"github.com/Karitham/corde"
	"github.com/jackc/pgx/v5"
)

type Drop struct {
	ID         int64
	Name       string
	ImageURL   string
	MediaTitle string
}

type Store interface {
	Delete(ctx context.Context, id corde.Snowflake) error
	Get(ctx context.Context, id corde.Snowflake) (*Drop, error)
	Set(ctx context.Context, id corde.Snowflake, data Drop) error
}

type PostgresStore struct {
	q Querier
}

func NewPostgresStore(q Querier) *PostgresStore {
	return &PostgresStore{
		q: q,
	}
}

func (p *PostgresStore) Set(ctx context.Context, id corde.Snowflake, data Drop) error {
	err := p.q.UpsertCharacter(ctx, UpsertCharacterParams{
		ID:         data.ID,
		Name:       data.Name,
		Image:      data.ImageURL,
		MediaTitle: data.MediaTitle,
	})
	if err != nil {
		return fmt.Errorf("failed to upsert character: %w", err)
	}

	return p.q.SetDrop(ctx, SetDropParams{
		ChannelID:   uint64(id),
		CharacterID: data.ID,
	})
}

func (p *PostgresStore) Get(ctx context.Context, id corde.Snowflake) (*Drop, error) {
	row, err := p.q.GetDrop(ctx, uint64(id))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("no drop found")
		}
		return nil, fmt.Errorf("failed to get channel char: %w", err)
	}

	return &Drop{
		ID:         row.ID,
		Name:       row.Name,
		ImageURL:   row.Image,
		MediaTitle: row.MediaTitle,
	}, nil
}

func (p *PostgresStore) Delete(ctx context.Context, id corde.Snowflake) error {
	return p.q.DeleteDrop(ctx, uint64(id))
}
