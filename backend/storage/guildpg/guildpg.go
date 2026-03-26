package guildpg

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/karitham/waifubot/collection"
	"github.com/karitham/waifubot/storage/guildstore"
)

type Pg struct {
	Q guildstore.Querier
}

func New(q guildstore.Querier) *Pg {
	return &Pg{Q: q}
}

func (p *Pg) IsGuildIndexed(ctx context.Context, guildID uint64) (collection.GuildIndexStatus, error) {
	row, err := p.Q.IsGuildIndexed(ctx, guildID)
	if err != nil {
		return collection.GuildIndexStatus{}, err
	}
	return collection.GuildIndexStatus{
		Status:    collection.ConvertIndexingStatus(string(row.Status)),
		UpdatedAt: row.UpdatedAt.Time,
	}, nil
}

func (p *Pg) StartIndexingJob(ctx context.Context, guildID uint64) error {
	return p.Q.StartIndexingJob(ctx, guildID)
}

func (p *Pg) CompleteIndexingJob(ctx context.Context, guildID uint64) error {
	return p.Q.CompleteIndexingJob(ctx, guildID)
}

func (p *Pg) UpsertGuildMembers(ctx context.Context, guildID uint64, memberIDs []uint64, indexedAt time.Time) error {
	ids := make([]int64, len(memberIDs))
	for i, id := range memberIDs {
		ids[i] = int64(id)
	}
	return p.Q.UpsertGuildMembers(ctx, guildstore.UpsertGuildMembersParams{
		GuildID:   guildID,
		Column2:   ids,
		IndexedAt: pgtype.Timestamp{Time: indexedAt.UTC(), Valid: true},
	})
}

func (p *Pg) DeleteGuildMembersNotIn(ctx context.Context, guildID uint64, memberIDs []uint64) error {
	ids := make([]int64, len(memberIDs))
	for i, id := range memberIDs {
		ids[i] = int64(id)
	}
	return p.Q.DeleteGuildMembersNotIn(ctx, guildstore.DeleteGuildMembersNotInParams{
		GuildID: guildID,
		Column2: ids,
	})
}
