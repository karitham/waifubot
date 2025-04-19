package memstore

import (
	"context"

	"github.com/Karitham/corde"
)

// IncrementInteractionCount increments the interaction count for the given channel/user combo
func (c Client) IncrementInteractionCount(ctx context.Context, channelID corde.Snowflake) error {
	key := "channel:" + channelID.String() + ":interactions"
	return c.client.Incr(ctx, key).Err()
}

// GetInteractionCount gets the interaction count for the given channel/user combo
func (c Client) GetInteractionCount(ctx context.Context, channelID corde.Snowflake) (int64, error) {
	key := "channel:" + channelID.String() + ":interactions"
	return c.client.Get(ctx, key).Int64()
}

// ResetInteractionCount resets the interaction count for the given channel/user combo
func (c Client) ResetInteractionCount(ctx context.Context, channelID corde.Snowflake) error {
	key := "channel:" + channelID.String() + ":interactions"
	return c.client.Del(ctx, key).Err()
}
