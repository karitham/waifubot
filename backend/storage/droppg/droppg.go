package droppg

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/karitham/waifubot/catalog"
	"github.com/karitham/waifubot/collection"
	"github.com/karitham/waifubot/storage"
)

type Pg struct {
	Q storage.Querier
}

func New(q storage.Querier) *Pg {
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

func (p *Pg) SetDrop(ctx context.Context, channelID collection.ChannelID, char catalog.Character) error {
	return p.Q.SetDrop(ctx, storage.SetDropParams{
		ChannelID:   uint64(channelID),
		CharacterID: int64(char.ID),
	})
}
