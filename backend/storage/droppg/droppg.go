package droppg

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/karitham/waifubot/catalog"
	"github.com/karitham/waifubot/collection"
	"github.com/karitham/waifubot/storage/dropstore"
)

type Pg struct {
	Q dropstore.Querier
}

func New(q dropstore.Querier) *Pg {
	return &Pg{Q: q}
}

func (p *Pg) GetDropForUpdate(ctx context.Context, channelID uint64) (catalog.Drop, error) {
	c, err := p.Q.GetDropForUpdate(ctx, channelID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return catalog.Drop{}, collection.ErrNotFound
		}
		return catalog.Drop{}, err
	}
	return catalog.Drop{ID: c.ID, Name: c.Name, Image: c.Image, MediaTitle: c.MediaTitle}, nil
}

func (p *Pg) DeleteDrop(ctx context.Context, channelID uint64) error {
	return p.Q.DeleteDrop(ctx, channelID)
}
