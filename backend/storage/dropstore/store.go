package dropstore

import (
	"context"

	"github.com/karitham/waifubot/catalog"
	"github.com/karitham/waifubot/collection"
)

type PostgresStore struct {
	q Querier
}

func NewPostgresStore(q Querier) *PostgresStore {
	return &PostgresStore{
		q: q,
	}
}

// Set stores the drop for a channel. The character must already be in the
// catalog; characters enter it only through the hydrator, and the Redis
// migration materializes them explicitly. The channel_drops FK enforces it.
func (p *PostgresStore) Set(ctx context.Context, id collection.ChannelID, data catalog.Character) error {
	return p.q.SetDrop(ctx, SetDropParams{
		ChannelID:   uint64(id),
		CharacterID: int64(data.ID),
	})
}
