package commandpg

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/karitham/waifubot/storage"
)

type Pg struct {
	Q storage.Querier
}

func New(q storage.Querier) *Pg {
	return &Pg{Q: q}
}

func (p *Pg) GetCommandHash(ctx context.Context) (string, error) {
	hash, err := p.Q.GetCommandHash(ctx)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", nil
		}
		return "", err
	}
	return hash, nil
}

func (p *Pg) SetCommandHash(ctx context.Context, hash string) error {
	return p.Q.SetCommandHash(ctx, hash)
}

func (p *Pg) UpdateCommandHash(ctx context.Context, hash string) error {
	return p.Q.UpdateCommandHash(ctx, hash)
}
