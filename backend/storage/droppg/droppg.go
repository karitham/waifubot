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

func (p *Pg) GetDropForUpdate(ctx context.Context, channelID collection.ChannelID) (catalog.Character, error) {
	c, err := p.Q.GetDropForUpdate(ctx, uint64(channelID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return catalog.Character{}, collection.ErrNotFound
		}
		return catalog.Character{}, err
	}
	return catalog.Character{
		ID:         catalog.CharacterID(c.ID),
		Name:       c.Name,
		Image:      c.Image,
		MediaTitle: c.MediaTitle,
		Popularity: int(c.Favorites),
	}, nil
}

func (p *Pg) DeleteDrop(ctx context.Context, channelID collection.ChannelID) error {
	return p.Q.DeleteDrop(ctx, uint64(channelID))
}
