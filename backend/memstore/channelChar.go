package memstore

import (
	"context"
	"fmt"

	"github.com/karitham/waifubot/discord"
	"github.com/Karitham/corde"
	"github.com/fxamacker/cbor/v2"
	"github.com/go-redis/redis/v8"
)

// Client is a cache client. It is currently implemented through redis.
type Client struct {
	client redis.UniversalClient
}

// New creates a new cache client.
func New(opts *redis.Options) Client {
	return Client{
		client: redis.NewClient(opts),
	}
}

// Close closes the client.
func (c Client) Close() error {
	return c.client.Close()
}

// SetChannelChar sets the currently dropped char in the channel
func (c Client) SetChannelChar(ctx context.Context, channelID corde.Snowflake, char discord.MediaCharacter) error {
	key := "channel:" + channelID.String() + ":char"

	b, err := cbor.Marshal(char)
	if err != nil {
		return fmt.Errorf("failed to marshal char: %w", err)
	}

	return c.client.Set(ctx, key, string(b), 0).Err()
}

// GetChannelChar gets the currently dropped char in the channel
func (c Client) GetChannelChar(ctx context.Context, channelID corde.Snowflake) (discord.MediaCharacter, error) {
	key := "channel:" + channelID.String() + ":char"

	s, err := c.client.Get(ctx, key).Result()
	if err != nil {
		return discord.MediaCharacter{}, fmt.Errorf("failed to get channel char: %w", err)
	}

	var char discord.MediaCharacter
	if err := cbor.Unmarshal([]byte(s), &char); err != nil {
		return discord.MediaCharacter{}, fmt.Errorf("failed to unmarshal char: %w", err)
	}

	return char, nil
}

// RemoveChannelChar removes the currently dropped char in the channel
func (c Client) RemoveChannelChar(ctx context.Context, channelID corde.Snowflake) error {
	key := "channel:" + channelID.String() + ":char"

	return c.client.Del(ctx, key).Err()
}
