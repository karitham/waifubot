package interfaces

import (
	"context"

	"github.com/Karitham/corde"
)

type InteractionRepository interface {
	Increment(ctx context.Context, channelID corde.Snowflake) error
	Get(ctx context.Context, channelID corde.Snowflake) (int64, error)
	Reset(ctx context.Context, channelID corde.Snowflake) error
}
