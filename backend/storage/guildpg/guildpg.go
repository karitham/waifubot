package guildpg

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/karitham/waifubot/catalog"
	"github.com/karitham/waifubot/guild"
	"github.com/karitham/waifubot/storage/guildstore"
)

type Pg struct {
	Q guildstore.Querier
}

func New(q guildstore.Querier) *Pg {
	return &Pg{Q: q}
}

func (p *Pg) IsGuildIndexed(ctx context.Context, guildID guild.GuildID) (guild.GuildIndexStatus, error) {
	row, err := p.Q.IsGuildIndexed(ctx, uint64(guildID))
	if err != nil {
		return guild.GuildIndexStatus{}, err
	}
	return guild.GuildIndexStatus{
		Status:    guild.ConvertIndexingStatus(string(row.Status)),
		UpdatedAt: row.UpdatedAt.Time,
	}, nil
}

func (p *Pg) StartIndexingJob(ctx context.Context, guildID guild.GuildID) error {
	return p.Q.StartIndexingJob(ctx, uint64(guildID))
}

func (p *Pg) CompleteIndexingJob(ctx context.Context, guildID guild.GuildID) error {
	return p.Q.CompleteIndexingJob(ctx, uint64(guildID))
}

func (p *Pg) UpsertGuildMembers(ctx context.Context, guildID guild.GuildID, memberIDs []guild.MemberID, indexedAt time.Time) error {
	ids := make([]int64, len(memberIDs))
	for i, id := range memberIDs {
		ids[i] = int64(id)
	}
	return p.Q.UpsertGuildMembers(ctx, guildstore.UpsertGuildMembersParams{
		GuildID:   uint64(guildID),
		Column2:   ids,
		IndexedAt: pgtype.Timestamp{Time: indexedAt.UTC(), Valid: true},
	})
}

func (p *Pg) DeleteGuildMembersNotIn(ctx context.Context, guildID guild.GuildID, memberIDs []guild.MemberID) error {
	ids := make([]int64, len(memberIDs))
	for i, id := range memberIDs {
		ids[i] = int64(id)
	}
	return p.Q.DeleteGuildMembersNotIn(ctx, guildstore.DeleteGuildMembersNotInParams{
		GuildID: uint64(guildID),
		Column2: ids,
	})
}

func (p *Pg) GetCharacterHoldersInGuild(ctx context.Context, guildID guild.GuildID, charID catalog.CharacterID) ([]guild.MemberID, error) {
	holders, err := p.Q.UsersOwningCharInGuild(ctx, guildstore.UsersOwningCharInGuildParams{
		CharacterID: int64(charID),
		GuildID:     uint64(guildID),
	})
	if err != nil {
		return nil, err
	}

	memberIDs := make([]guild.MemberID, len(holders))
	for i, id := range holders {
		memberIDs[i] = guild.MemberID(id)
	}
	return memberIDs, nil
}
