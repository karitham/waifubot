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
	"github.com/karitham/waifubot/storage/collectionstore"
	"github.com/karitham/waifubot/storage/userstore"
	"github.com/karitham/waifubot/storage/wishliststore"
)

type Pg struct {
	C collectionstore.Querier
	W wishliststore.Querier
	U userstore.Querier
}

func New(c collectionstore.Querier, w wishliststore.Querier, u userstore.Querier) *Pg {
	return &Pg{C: c, W: w, U: u}
}

func (p *Pg) GetCollection(ctx context.Context, userID collection.UserID) ([]collection.OwnedCharacter, error) {
	rows, err := p.C.List(ctx, uint64(userID))
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
	ids, err := p.C.ListIDs(ctx, uint64(userID))
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
	row, err := p.C.Get(ctx, collectionstore.GetParams{ID: int64(charID), UserID: uint64(userID)})
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
	_, err := p.C.Insert(ctx, collectionstore.InsertParams{
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
	_, err := p.C.Delete(ctx, collectionstore.DeleteParams{UserID: uint64(userID), CharacterID: int64(charID)})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return collection.ErrNotFound
		}
		return err
	}

	return p.clearFavorite(ctx, userID, charID)
}

func (p *Pg) GiveCharacter(ctx context.Context, from, to collection.UserID, charID catalog.CharacterID) (collection.OwnedCharacter, error) {
	col, err := p.C.Give(ctx, collectionstore.GiveParams{
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

	charRow, err := p.C.GetByID(ctx, int64(charID))
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
	return p.U.ClearFavorite(ctx, userstore.ClearFavoriteParams{
		UserID:   uint64(userID),
		Favorite: pgtype.Int8{Int64: int64(charID), Valid: true},
	})
}

func (p *Pg) SetFavorite(ctx context.Context, userID collection.UserID, charID catalog.CharacterID) error {
	var favorite pgtype.Int8
	if charID != 0 {
		favorite = pgtype.Int8{Int64: int64(charID), Valid: true}
	}

	rows, err := p.C.SetFavoriteIfOwned(ctx, collectionstore.SetFavoriteIfOwnedParams{
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
	if _, err := p.U.Get(ctx, uint64(userID)); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return collection.ErrNotFound
		}
		return err
	}
	return fmt.Errorf("%w %d", collection.ErrUserDoesNotOwnCharacter, charID)
}

func (p *Pg) CountCollection(ctx context.Context, userID collection.UserID) (int64, error) {
	return p.C.Count(ctx, uint64(userID))
}

func (p *Pg) SearchCharacters(ctx context.Context, userID collection.UserID, term string) ([]catalog.Character, error) {
	rows, err := p.C.SearchCharacters(ctx, collectionstore.SearchCharactersParams{
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
	return p.W.RemoveCharactersFromWishlist(ctx, wishliststore.RemoveCharactersFromWishlistParams{
		UserID:  uint64(userID),
		Column2: []int64{int64(charID)},
	})
}

func (p *Pg) AddCharactersToWishlist(ctx context.Context, userID collection.UserID, characterIDs []catalog.CharacterID) error {
	ids := make([]int64, len(characterIDs))
	for i, id := range characterIDs {
		ids[i] = int64(id)
	}

	return p.W.AddCharactersToWishlist(ctx, wishliststore.AddCharactersToWishlistParams{
		UserID:  uint64(userID),
		Column2: ids,
	})
}

func (p *Pg) RandomCharNotOwned(ctx context.Context, userID collection.UserID, weightExponent float64) (catalog.Character, error) {
	c, err := p.C.RandomCharNotOwned(ctx, collectionstore.RandomCharNotOwnedParams{
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
	c, err := p.C.RandomActiveChar(ctx, weightExponent)
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
