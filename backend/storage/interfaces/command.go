package interfaces

import (
	"context"
)

type CommandRepository interface {
	GetCommandHash(ctx context.Context) (string, error)
	SetCommandHash(ctx context.Context, hash string) error
	UpdateCommandHash(ctx context.Context, hash string) error
}
