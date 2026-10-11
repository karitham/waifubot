package dropstore

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/karitham/waifubot/catalog"
	"github.com/karitham/waifubot/collection"
)

// Store persists the character currently dropped in a channel.
type Store interface {
	Delete(ctx context.Context, id collection.ChannelID) error
	Get(ctx context.Context, id collection.ChannelID) (*catalog.Character, error)
	Set(ctx context.Context, id collection.ChannelID, data catalog.Character) error
}

type PostgresStore struct {
	q Querier
}

func NewPostgresStore(q Querier) *PostgresStore {
	return &PostgresStore{
		q: q,
	}
}

func (p *PostgresStore) Set(ctx context.Context, id collection.ChannelID, data catalog.Character) error {
	err := p.q.UpsertCharacter(ctx, UpsertCharacterParams{
		ID:         int64(data.ID),
		Name:       data.Name,
		Image:      data.Image,
		MediaTitle: data.MediaTitle,
		Favorites:  int32(data.Favorites),
	})
	if err != nil {
		return fmt.Errorf("failed to upsert character: %w", err)
	}

	return p.q.SetDrop(ctx, SetDropParams{
		ChannelID:   uint64(id),
		CharacterID: int64(data.ID),
	})
}

func (p *PostgresStore) Get(ctx context.Context, id collection.ChannelID) (*catalog.Character, error) {
	row, err := p.q.GetDrop(ctx, uint64(id))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("no drop found")
		}
		return nil, fmt.Errorf("failed to get channel char: %w", err)
	}

	return &catalog.Character{
		ID:         catalog.CharacterID(row.ID),
		Name:       row.Name,
		Image:      row.Image,
		MediaTitle: row.MediaTitle,
		Favorites:  int(row.Favorites),
	}, nil
}

func (p *PostgresStore) Delete(ctx context.Context, id collection.ChannelID) error {
	return p.q.DeleteDrop(ctx, uint64(id))
}
