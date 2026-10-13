package pg

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/karitham/waifubot/catalog"
	"github.com/karitham/waifubot/storage"
)

func (b *base) UpsertCharacter(ctx context.Context, char catalog.Character) error {
	_, err := b.q.UpsertCharacter(ctx, storage.UpsertCharacterParams{
		ID:         int64(char.ID),
		Name:       char.Name,
		Image:      char.Image,
		MediaTitle: char.MediaTitle,
		Favorites:  int32(char.Popularity),
	})
	return err
}

func (b *base) GetCharacterByID(ctx context.Context, charID catalog.CharacterID) (catalog.Character, error) {
	c, err := b.q.GetByID(ctx, int64(charID))
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
		Popularity: int(c.Favorites),
		UpdatedAt:  c.UpdatedAt.Time,
	}, nil
}

func (b *base) SearchGlobalCharacters(ctx context.Context, term string) ([]catalog.Character, error) {
	rows, err := b.q.SearchGlobalCharacters(ctx, storage.SearchGlobalCharactersParams{Term: term, Lim: 25})
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
			Popularity: int(r.Favorites),
		}
	}
	return chars, nil
}

func (b *base) GetActiveIDs(ctx context.Context) ([]catalog.CharacterID, error) {
	ids, err := b.q.GetActiveIDs(ctx)
	if err != nil {
		return nil, err
	}

	charIDs := make([]catalog.CharacterID, len(ids))
	for i, id := range ids {
		charIDs[i] = catalog.CharacterID(id)
	}
	return charIDs, nil
}

// KnownCharacterIDs returns the subset of the given IDs present in the catalog.
func (b *base) KnownCharacterIDs(ctx context.Context, ids []catalog.CharacterID) ([]catalog.CharacterID, error) {
	dbIDs := make([]int64, len(ids))
	for i, id := range ids {
		dbIDs[i] = int64(id)
	}

	rows, err := b.q.KnownCharacterIDs(ctx, dbIDs)
	if err != nil {
		return nil, err
	}

	out := make([]catalog.CharacterID, len(rows))
	for i, id := range rows {
		out[i] = catalog.CharacterID(id)
	}
	return out, nil
}

func (b *base) MarkCharactersInactive(ctx context.Context, charIDs []catalog.CharacterID) error {
	ids := make([]int64, len(charIDs))
	for i, id := range charIDs {
		ids[i] = int64(id)
	}
	return b.q.MarkCharactersInactive(ctx, ids)
}
