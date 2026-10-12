package guildpg

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
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

func (p *Pg) IsGuildIndexed(ctx context.Context, guildID guild.ID) (guild.IndexStatus, error) {
	row, err := p.Q.IsGuildIndexed(ctx, uint64(guildID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// A guild with no job row is simply not indexed yet.
			return guild.IndexStatus{Status: guild.IndexingPending}, nil
		}
		return guild.IndexStatus{}, err
	}
	return guild.IndexStatus{
		Status:    toIndexingStatus(row.Status),
		UpdatedAt: row.UpdatedAt.Time,
	}, nil
}

// toIndexingStatus maps a stored indexing status to the domain type.
func toIndexingStatus(status guildstore.IndexingStatus) guild.IndexingStatus {
	switch status {
	case guildstore.IndexingStatusCompleted:
		return guild.IndexingCompleted
	case guildstore.IndexingStatusInProgress:
		return guild.IndexingInProgress
	default:
		return guild.IndexingPending
	}
}

func (p *Pg) StartIndexingJob(ctx context.Context, guildID guild.ID) error {
	return p.Q.StartIndexingJob(ctx, uint64(guildID))
}

func (p *Pg) CompleteIndexingJob(ctx context.Context, guildID guild.ID) error {
	return p.Q.CompleteIndexingJob(ctx, uint64(guildID))
}

func (p *Pg) UpsertGuildMembers(ctx context.Context, guildID guild.ID, memberIDs []guild.MemberID, indexedAt time.Time) error {
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

func (p *Pg) DeleteGuildMembersNotIn(ctx context.Context, guildID guild.ID, memberIDs []guild.MemberID) error {
	ids := make([]int64, len(memberIDs))
	for i, id := range memberIDs {
		ids[i] = int64(id)
	}
	return p.Q.DeleteGuildMembersNotIn(ctx, guildstore.DeleteGuildMembersNotInParams{
		GuildID: uint64(guildID),
		Column2: ids,
	})
}

func (p *Pg) GetCharacterHoldersInGuild(ctx context.Context, guildID guild.ID, charID catalog.CharacterID) ([]guild.MemberID, error) {
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
