package pg

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/karitham/waifubot/catalog"
	"github.com/karitham/waifubot/drops"
	"github.com/karitham/waifubot/storage"
)

func (b *base) IncrementActivity(ctx context.Context, channelID drops.ChannelID) (int64, error) {
	return b.q.IncrementActivity(ctx, uint64(channelID))
}

func (b *base) ResetActivity(ctx context.Context, channelID drops.ChannelID) error {
	return b.q.ResetActivity(ctx, uint64(channelID))
}

func (b *base) RandomActiveChar(ctx context.Context, weightExponent float64) (catalog.Character, error) {
	c, err := b.q.RandomActiveChar(ctx, weightExponent)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return catalog.Character{}, catalog.ErrNotFound
		}
		return catalog.Character{}, err
	}
	return catalog.Character{
		ID:         catalog.CharacterID(c.ID),
		Name:       c.Name,
		Image:      c.Image,
		MediaTitle: c.MediaTitle,
		Popularity: int(c.Favorites),
		UpdatedAt:  c.UpdatedAt.Time,
		IsActive:   c.IsActive,
	}, nil
}

func (b *base) SetDrop(ctx context.Context, drop drops.Drop) error {
	return b.q.SetDrop(ctx, storage.SetDropParams{
		ChannelID:   uint64(drop.Channel),
		CharacterID: int64(drop.Character.ID),
	})
}

func (t *Tx) GetDropForUpdate(ctx context.Context, channelID drops.ChannelID) (drops.Drop, error) {
	c, err := t.q.GetDropForUpdate(ctx, uint64(channelID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return drops.Drop{}, drops.ErrNoDropInChannel
		}
		return drops.Drop{}, err
	}
	return drops.Drop{
		Channel: channelID,
		Character: catalog.Character{
			ID:         catalog.CharacterID(c.ID),
			Name:       c.Name,
			Image:      c.Image,
			MediaTitle: c.MediaTitle,
			Popularity: int(c.Favorites),
		},
	}, nil
}

func (t *Tx) DeleteDrop(ctx context.Context, channelID drops.ChannelID) error {
	return t.q.DeleteDrop(ctx, uint64(channelID))
}
