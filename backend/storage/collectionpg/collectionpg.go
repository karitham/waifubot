package collectionpg

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/karitham/waifubot/collection"
	"github.com/karitham/waifubot/storage/collectionstore"
	"github.com/karitham/waifubot/storage/wishliststore"
)

type Pg struct {
	C collectionstore.Querier
	W wishliststore.Querier
}

func New(c collectionstore.Querier, w wishliststore.Querier) *Pg {
	return &Pg{C: c, W: w}
}

func (p *Pg) GetCollection(ctx context.Context, userID collection.UserID) ([]collection.OwnedCharacter, error) {
	rows, err := p.C.List(ctx, userID)
	if err != nil {
		return nil, err
	}
	chars := make([]collection.OwnedCharacter, len(rows))
	for i, r := range rows {
		chars[i] = collection.OwnedCharacter{
			Character: collection.Character{ID: r.ID, Name: r.Name, Image: r.Image, MediaTitle: r.MediaTitle},
			Date:      r.Date.Time,
			Source:    r.Source,
			UserID:    userID,
		}
	}
	return chars, nil
}

func (p *Pg) GetCollectionIDs(ctx context.Context, userID collection.UserID) ([]int64, error) {
	return p.C.ListIDs(ctx, userID)
}

func (p *Pg) GetOwnedCharacter(ctx context.Context, userID collection.UserID, charID int64) (collection.OwnedCharacter, error) {
	row, err := p.C.Get(ctx, collectionstore.GetParams{ID: charID, UserID: userID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return collection.OwnedCharacter{}, collection.ErrNotFound
		}
		return collection.OwnedCharacter{}, err
	}
	return collection.OwnedCharacter{
		Character: collection.Character{ID: row.ID, Name: row.Name, Image: row.Image, MediaTitle: row.MediaTitle},
		Date:      row.Date.Time,
		Source:    row.Source,
		UserID:    userID,
	}, nil
}

func (p *Pg) CharacterOwnedByUser(ctx context.Context, userID collection.UserID, charID int64) (bool, error) {
	_, err := p.GetOwnedCharacter(ctx, userID, charID)
	if err != nil {
		if errors.Is(err, collection.ErrNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (p *Pg) AddToCollection(ctx context.Context, userID collection.UserID, char collection.Character, source string, acquiredAt time.Time) error {
	_, err := p.C.Insert(ctx, collectionstore.InsertParams{
		UserID:      userID,
		CharacterID: char.ID,
		Source:      source,
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

func (p *Pg) RemoveFromCollection(ctx context.Context, userID collection.UserID, charID int64) error {
	_, err := p.C.Delete(ctx, collectionstore.DeleteParams{UserID: userID, CharacterID: charID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return collection.ErrNotFound
		}
		return err
	}
	return nil
}

func (p *Pg) GiveCharacter(ctx context.Context, from, to collection.UserID, charID int64) (collection.OwnedCharacter, error) {
	col, err := p.C.Give(ctx, collectionstore.GiveParams{UserID: to, CharacterID: charID, UserID_2: from})
	if err != nil {
		return collection.OwnedCharacter{}, err
	}

	charRow, err := p.C.GetByID(ctx, charID)
	if err != nil {
		return collection.OwnedCharacter{}, nil
	}

	return collection.OwnedCharacter{
		Character: collection.Character{ID: charRow.ID, Name: charRow.Name, Image: charRow.Image, MediaTitle: charRow.MediaTitle},
		Date:      col.AcquiredAt.Time,
		Source:    col.Source,
		UserID:    to,
	}, nil
}

func (p *Pg) CountCollection(ctx context.Context, userID collection.UserID) (int64, error) {
	return p.C.Count(ctx, userID)
}

func (p *Pg) RemoveFromWishlist(ctx context.Context, userID collection.UserID, charID int64) error {
	_ = p.W.RemoveCharactersFromWishlist(ctx, wishliststore.RemoveCharactersFromWishlistParams{
		UserID:  userID,
		Column2: []int64{charID},
	})
	return nil
}
