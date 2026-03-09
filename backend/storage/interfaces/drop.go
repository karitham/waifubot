package interfaces

import (
	"context"

	"github.com/Karitham/corde"
)

type Drop struct {
	ID         int64
	Name       string
	ImageURL   string
	MediaTitle string
}

type DropRepository interface {
	Delete(ctx context.Context, id corde.Snowflake) error
	Get(ctx context.Context, id corde.Snowflake) (*Drop, error)
	Set(ctx context.Context, id corde.Snowflake, data Drop) error
}
