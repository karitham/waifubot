package collectionpg

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/karitham/waifubot/catalog"
	"github.com/karitham/waifubot/collection"
	"github.com/karitham/waifubot/storage"
)

type Pg struct {
	q storage.Querier
}

func New(q storage.Querier) *Pg {
	return &Pg{q: q}
}

func (p *Pg) GetCollection(ctx context.Context, userID collection.UserID) ([]collection.OwnedCharacter, error) {
	rows, err := p.q.List(ctx, uint64(userID))
	if err != nil {
		return nil, err
	}
	chars := make([]collection.OwnedCharacter, len(rows))
	for i, r := range rows {
		chars[i] = collection.OwnedCharacter{
			ID: catalog.CharacterID(r.ID), Name: r.Name, Image: r.Image, MediaTitle: r.MediaTitle, Popularity: int(r.Favorites),
			Date:   r.Date.Time,
			Source: collection.Source(r.Source),
			UserID: userID,
		}
	}
	return chars, nil
}

func (p *Pg) GetCollectionIDs(ctx context.Context, userID collection.UserID) ([]catalog.CharacterID, error) {
	ids, err := p.q.ListIDs(ctx, uint64(userID))
	if err != nil {
		return nil, err
	}

	charIDs := make([]catalog.CharacterID, len(ids))
	for i, id := range ids {
		charIDs[i] = catalog.CharacterID(id)
	}
	return charIDs, nil
}

func (p *Pg) GetOwnedCharacter(ctx context.Context, userID collection.UserID, charID catalog.CharacterID) (collection.OwnedCharacter, error) {
	row, err := p.q.GetCollection(ctx, storage.GetCollectionParams{ID: int64(charID), UserID: uint64(userID)})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return collection.OwnedCharacter{}, collection.ErrNotFound
		}
		return collection.OwnedCharacter{}, err
	}
	return collection.OwnedCharacter{
		ID: catalog.CharacterID(row.ID), Name: row.Name, Image: row.Image, MediaTitle: row.MediaTitle, Popularity: int(row.Favorites),
		Date:   row.Date.Time,
		Source: collection.Source(row.Source),
		UserID: userID,
	}, nil
}

func (p *Pg) AddToCollection(ctx context.Context, userID collection.UserID, charID catalog.CharacterID, source collection.Source, acquiredAt time.Time) error {
	_, err := p.q.Insert(ctx, storage.InsertParams{
		UserID:      uint64(userID),
		CharacterID: int64(charID),
		Source:      string(source),
		AcquiredAt:  pgtype.Timestamp{Time: acquiredAt.UTC(), Valid: true},
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return collection.ErrAlreadyOwned
		}
		return err
	}
	return nil
}

func (p *Pg) RemoveFromCollection(ctx context.Context, userID collection.UserID, charID catalog.CharacterID) error {
	_, err := p.q.Delete(ctx, storage.DeleteParams{UserID: uint64(userID), CharacterID: int64(charID)})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return collection.ErrNotFound
		}
		return err
	}

	return p.clearFavorite(ctx, userID, charID)
}

func (p *Pg) GiveCharacter(ctx context.Context, from, to collection.UserID, charID catalog.CharacterID) (collection.OwnedCharacter, error) {
	col, err := p.q.Give(ctx, storage.GiveParams{
		UserID:      uint64(to),
		CharacterID: int64(charID),
		UserID_2:    uint64(from),
		Source:      string(collection.SourceTrade),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return collection.OwnedCharacter{}, collection.ErrNotFound
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return collection.OwnedCharacter{}, collection.ErrAlreadyOwned
		}
		return collection.OwnedCharacter{}, err
	}

	if err := p.clearFavorite(ctx, from, charID); err != nil {
		return collection.OwnedCharacter{}, err
	}

	charRow, err := p.q.GetByID(ctx, int64(charID))
	if err != nil {
		return collection.OwnedCharacter{}, err
	}

	return collection.OwnedCharacter{
		ID: catalog.CharacterID(charRow.ID), Name: charRow.Name, Image: charRow.Image, MediaTitle: charRow.MediaTitle, Popularity: int(charRow.Favorites),
		Date:   col.AcquiredAt.Time,
		Source: collection.Source(col.Source),
		UserID: to,
	}, nil
}

// clearFavorite removes the character as the user's favorite when it is.
func (p *Pg) clearFavorite(ctx context.Context, userID collection.UserID, charID catalog.CharacterID) error {
	return p.q.ClearFavorite(ctx, storage.ClearFavoriteParams{
		UserID:   uint64(userID),
		Favorite: pgtype.Int8{Int64: int64(charID), Valid: true},
	})
}

func (p *Pg) SetFavorite(ctx context.Context, userID collection.UserID, charID catalog.CharacterID) error {
	var favorite pgtype.Int8
	if charID != 0 {
		favorite = pgtype.Int8{Int64: int64(charID), Valid: true}
	}

	rows, err := p.q.SetFavoriteIfOwned(ctx, storage.SetFavoriteIfOwnedParams{
		UserID:      uint64(userID),
		CharacterID: int64(charID),
		Favorite:    favorite,
	})
	if err != nil {
		return err
	}
	if rows > 0 {
		return nil
	}

	// No row updated: either the user is missing or the character is not
	// in their collection.
	if _, err := p.q.GetUser(ctx, uint64(userID)); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return collection.ErrNotFound
		}
		return err
	}
	return fmt.Errorf("%w %d", collection.ErrUserDoesNotOwnCharacter, charID)
}

func (p *Pg) CountCollection(ctx context.Context, userID collection.UserID) (int64, error) {
	return p.q.Count(ctx, uint64(userID))
}

func (p *Pg) SearchCharacters(ctx context.Context, userID collection.UserID, term string) ([]catalog.Character, error) {
	rows, err := p.q.SearchCharacters(ctx, storage.SearchCharactersParams{
		UserID: uint64(userID),
		Term:   term,
		Lim:    25,
		Off:    0,
	})
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

func (p *Pg) RemoveFromWishlist(ctx context.Context, userID collection.UserID, charID catalog.CharacterID) error {
	return p.q.RemoveCharactersFromWishlist(ctx, storage.RemoveCharactersFromWishlistParams{
		UserID:  uint64(userID),
		Column2: []int64{int64(charID)},
	})
}

func (p *Pg) RemoveAllFromWishlist(ctx context.Context, userID collection.UserID) error {
	return p.q.RemoveAllFromWishlist(ctx, uint64(userID))
}

func (p *Pg) AddCharactersToWishlist(ctx context.Context, userID collection.UserID, characterIDs []catalog.CharacterID) (int64, error) {
	ids := make([]int64, len(characterIDs))
	for i, id := range characterIDs {
		ids[i] = int64(id)
	}

	return p.q.AddCharactersToWishlist(ctx, storage.AddCharactersToWishlistParams{
		UserID:  uint64(userID),
		Column2: ids,
	})
}

func (p *Pg) RandomCharNotOwned(ctx context.Context, userID collection.UserID, weightExponent float64) (catalog.Character, error) {
	c, err := p.q.RandomCharNotOwned(ctx, storage.RandomCharNotOwnedParams{
		UserID:         uint64(userID),
		WeightExponent: weightExponent,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return catalog.Character{}, collection.ErrNotFound
		}
		return catalog.Character{}, err
	}
	return catalog.Character{
		ID:         catalog.CharacterID(c.ID),
		Name:       c.Name,
		Image:      c.Image,
		MediaTitle: c.MediaTitle,
		Popularity: int(c.Favorites),
	}, nil
}

func (p *Pg) RandomActiveChar(ctx context.Context, weightExponent float64) (catalog.Character, error) {
	c, err := p.q.RandomActiveChar(ctx, weightExponent)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return catalog.Character{}, collection.ErrNotFound
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
		IsActive:   c.IsActive,
	}, nil
}
