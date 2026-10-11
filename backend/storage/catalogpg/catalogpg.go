package catalogpg

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/karitham/waifubot/catalog"
	"github.com/karitham/waifubot/storage/collectionstore"
)

type Pg struct {
	C collectionstore.Querier
}

func New(c collectionstore.Querier) *Pg {
	return &Pg{C: c}
}

func (p *Pg) UpsertCharacter(ctx context.Context, char catalog.Character) error {
	_, err := p.C.UpsertCharacter(ctx, collectionstore.UpsertCharacterParams{
		ID:         int64(char.ID),
		Name:       char.Name,
		Image:      char.Image,
		MediaTitle: char.MediaTitle,
		Favorites:  int32(char.Favorites),
	})
	return err
}

func (p *Pg) GetCharacterByID(ctx context.Context, charID catalog.CharacterID) (catalog.Character, error) {
	c, err := p.C.GetByID(ctx, int64(charID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return catalog.Character{}, catalog.ErrNotFound
		}
		return catalog.Character{}, err
	}
	return catalog.Character{
		ID:         catalog.CharacterID(c.ID),
		Name:       c.Name,
		Image:      c.Image,
		MediaTitle: c.MediaTitle,
		Favorites:  int(c.Favorites),
		UpdatedAt:  c.UpdatedAt.Time,
	}, nil
}

func (p *Pg) SearchGlobalCharacters(ctx context.Context, term string) ([]catalog.Character, error) {
	rows, err := p.C.SearchGlobalCharacters(ctx, collectionstore.SearchGlobalCharactersParams{Term: term, Lim: 25})
	if err != nil {
		return nil, err
	}
	chars := make([]catalog.Character, len(rows))
	for i, r := range rows {
		chars[i] = catalog.Character{
			ID:         catalog.CharacterID(r.ID),
			Name:       r.Name,
			Image:      r.Image,
			MediaTitle: r.MediaTitle,
			Favorites:  int(r.Favorites),
		}
	}
	return chars, nil
}

func (p *Pg) GetActiveIDs(ctx context.Context) ([]catalog.CharacterID, error) {
	ids, err := p.C.GetActiveIDs(ctx)
	if err != nil {
		return nil, err
	}

	charIDs := make([]catalog.CharacterID, len(ids))
	for i, id := range ids {
		charIDs[i] = catalog.CharacterID(id)
	}
	return charIDs, nil
}

func (p *Pg) MarkCharactersInactive(ctx context.Context, charIDs []catalog.CharacterID) error {
	ids := make([]int64, len(charIDs))
	for i, id := range charIDs {
		ids[i] = int64(id)
	}
	return p.C.MarkCharactersInactive(ctx, ids)
}
