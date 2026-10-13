package pg

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/karitham/waifubot/catalog"
	"github.com/karitham/waifubot/guild"
	"github.com/karitham/waifubot/storage"
)

func (b *base) IsGuildIndexed(ctx context.Context, guildID guild.ID) (guild.IndexStatus, error) {
	row, err := b.q.IsGuildIndexed(ctx, uint64(guildID))
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
func toIndexingStatus(status storage.IndexingStatus) guild.IndexingStatus {
	switch status {
	case storage.IndexingStatusCompleted:
		return guild.IndexingCompleted
	case storage.IndexingStatusInProgress:
		return guild.IndexingInProgress
	default:
		return guild.IndexingPending
	}
}

func (t *Tx) StartIndexingJob(ctx context.Context, guildID guild.ID) error {
	return t.q.StartIndexingJob(ctx, uint64(guildID))
}

func (t *Tx) CompleteIndexingJob(ctx context.Context, guildID guild.ID) error {
	return t.q.CompleteIndexingJob(ctx, uint64(guildID))
}

func (t *Tx) UpsertGuildMembers(ctx context.Context, guildID guild.ID, memberIDs []guild.MemberID, indexedAt time.Time) error {
	ids := make([]int64, len(memberIDs))
	for i, id := range memberIDs {
		ids[i] = int64(id)
	}
	return t.q.UpsertGuildMembers(ctx, storage.UpsertGuildMembersParams{
		GuildID:   uint64(guildID),
		Column2:   ids,
		IndexedAt: pgtype.Timestamp{Time: indexedAt.UTC(), Valid: true},
	})
}

func (t *Tx) DeleteGuildMembersNotIn(ctx context.Context, guildID guild.ID, memberIDs []guild.MemberID) error {
	ids := make([]int64, len(memberIDs))
	for i, id := range memberIDs {
		ids[i] = int64(id)
	}
	return t.q.DeleteGuildMembersNotIn(ctx, storage.DeleteGuildMembersNotInParams{
		GuildID: uint64(guildID),
		Column2: ids,
	})
}

func (b *base) GetCharacterHoldersInGuild(ctx context.Context, guildID guild.ID, charID catalog.CharacterID) ([]guild.MemberID, error) {
	holders, err := b.q.UsersOwningCharInGuild(ctx, storage.UsersOwningCharInGuildParams{
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
