package pg

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

func (b *base) GetCommandHash(ctx context.Context) (string, error) {
	hash, err := b.q.GetCommandHash(ctx)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", nil
		}
		return "", err
	}
	return hash, nil
}

func (b *base) SetCommandHash(ctx context.Context, hash string) error {
	return b.q.SetCommandHash(ctx, hash)
}

func (b *base) UpdateCommandHash(ctx context.Context, hash string) error {
	return b.q.UpdateCommandHash(ctx, hash)
}
