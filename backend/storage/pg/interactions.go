package pg

import (
	"context"

	"github.com/Karitham/corde"
)

func (b *base) Increment(ctx context.Context, channelID corde.Snowflake) error {
	return b.q.IncrementInteraction(ctx, uint64(channelID))
}

func (b *base) Get(ctx context.Context, channelID corde.Snowflake) (int64, error) {
	count, err := b.q.GetInteraction(ctx, uint64(channelID))
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (b *base) Reset(ctx context.Context, channelID corde.Snowflake) error {
	return b.q.ResetInteraction(ctx, uint64(channelID))
}
