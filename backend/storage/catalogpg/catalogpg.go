package catalogpg

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/karitham/waifubot/catalog"
	"github.com/karitham/waifubot/collection"
	"github.com/karitham/waifubot/storage/collectionstore"
	"github.com/karitham/waifubot/storage/guildstore"
)

type Pg struct {
	C collectionstore.Querier
	G guildstore.Querier
}

func New(c collectionstore.Querier, g guildstore.Querier) *Pg {
	return &Pg{C: c, G: g}
}

func (p *Pg) UpsertCharacter(ctx context.Context, char catalog.Character) error {
	_, err := p.C.UpsertCharacter(ctx, collectionstore.UpsertCharacterParams{
		ID:         char.ID,
		Name:       char.Name,
		Image:      char.Image,
		MediaTitle: char.MediaTitle,
		Favorites:  int32(char.Favorites),
	})
	return err
}

func (p *Pg) GetCharacterByID(ctx context.Context, charID int64) (catalog.Character, error) {
	c, err := p.C.GetByID(ctx, charID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return catalog.Character{}, collection.ErrNotFound
		}
		return catalog.Character{}, err
	}
	return catalog.Character{ID: c.ID, Name: c.Name, Image: c.Image, MediaTitle: c.MediaTitle, Favorites: int(c.Favorites), UpdatedAt: c.UpdatedAt.Time}, nil
}

func (p *Pg) SearchCharacters(ctx context.Context, userID uint64, term string) ([]catalog.Character, error) {
	rows, err := p.C.SearchCharacters(ctx, collectionstore.SearchCharactersParams{
		UserID: userID,
		Term:   term,
		Lim:    25,
		Off:    0,
	})
	if err != nil {
		return nil, err
	}
	chars := make([]catalog.Character, len(rows))
	for i, r := range rows {
		chars[i] = catalog.Character{ID: r.ID, Name: r.Name, Image: r.Image, MediaTitle: r.MediaTitle, Favorites: int(r.Favorites)}
	}
	return chars, nil
}

func (p *Pg) SearchGlobalCharacters(ctx context.Context, term string) ([]catalog.Character, error) {
	rows, err := p.C.SearchGlobalCharacters(ctx, collectionstore.SearchGlobalCharactersParams{Term: term, Lim: 25})
	if err != nil {
		return nil, err
	}
	chars := make([]catalog.Character, len(rows))
	for i, r := range rows {
		chars[i] = catalog.Character{ID: r.ID, Name: r.Name, Image: r.Image, MediaTitle: r.MediaTitle, Favorites: int(r.Favorites)}
	}
	return chars, nil
}

func (p *Pg) IsGuildIndexed(ctx context.Context, guildID uint64) (collection.GuildIndexStatus, error) {
	row, err := p.G.IsGuildIndexed(ctx, guildID)
	if err != nil {
		return collection.GuildIndexStatus{}, err
	}
	return collection.GuildIndexStatus{
		Status:    collection.ConvertIndexingStatus(string(row.Status)),
		UpdatedAt: row.UpdatedAt.Time,
	}, nil
}

func (p *Pg) GetCharacterHoldersInGuild(ctx context.Context, guildID uint64, charID int64) ([]uint64, error) {
	return p.G.UsersOwningCharInGuild(ctx, guildstore.UsersOwningCharInGuildParams{
		CharacterID: charID,
		GuildID:     guildID,
	})
}

func (p *Pg) StartIndexingJob(ctx context.Context, guildID uint64) error {
	return p.G.StartIndexingJob(ctx, guildID)
}

func (p *Pg) CompleteIndexingJob(ctx context.Context, guildID uint64) error {
	return p.G.CompleteIndexingJob(ctx, guildID)
}

func (p *Pg) UpsertGuildMembers(ctx context.Context, guildID uint64, memberIDs []uint64, indexedAt time.Time) error {
	ids := make([]int64, len(memberIDs))
	for i, id := range memberIDs {
		ids[i] = int64(id)
	}
	return p.G.UpsertGuildMembers(ctx, guildstore.UpsertGuildMembersParams{
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
	return p.G.DeleteGuildMembersNotIn(ctx, guildstore.DeleteGuildMembersNotInParams{
		GuildID: guildID,
		Column2: ids,
	})
}

func (p *Pg) MarkCharactersInactive(ctx context.Context, ids []int64) error {
	return p.C.MarkCharactersInactive(ctx, ids)
}

func (p *Pg) GetActiveIDs(ctx context.Context) ([]int64, error) {
	return p.C.GetActiveIDs(ctx)
}
