package pg

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/karitham/waifubot/catalog"
	"github.com/karitham/waifubot/collection"
	"github.com/karitham/waifubot/storage"
)

func (t *Tx) GetDropForUpdate(ctx context.Context, channelID collection.ChannelID) (catalog.Character, error) {
	c, err := t.q.GetDropForUpdate(ctx, uint64(channelID))
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

func (t *Tx) DeleteDrop(ctx context.Context, channelID collection.ChannelID) error {
	return t.q.DeleteDrop(ctx, uint64(channelID))
}

func (t *Tx) SetDrop(ctx context.Context, channelID collection.ChannelID, char catalog.Character) error {
	return t.q.SetDrop(ctx, storage.SetDropParams{
		ChannelID:   uint64(channelID),
		CharacterID: int64(char.ID),
	})
}
