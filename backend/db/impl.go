package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Karitham/corde"
	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"

	"github.com/karitham/waifubot/discord"
)

//go:generate sqlc generate

type Store struct {
	*Queries
	conn *pgx.Conn
}

// NewDB initialises the connection with the db
func NewDB(ctx context.Context, url string) (*Store, error) {
	conn, err := pgx.Connect(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("couldn't connect to db: %w", err)
	}

	if err := conn.Ping(ctx); err != nil {
		return nil, err
	}

	return &Store{
		Queries: New(conn),
		conn:    conn,
	}, nil
}

// Tx executes a function in a transaction.
func (q *Store) Tx(ctx context.Context, fn func(s discord.Store) error) error {
	return q.asTx(ctx, func(s *Store) error {
		return fn(s)
	})
}

func (q *Store) asTx(ctx context.Context, fn func(q *Store) error) error {
	// already in tx
	if q.conn == nil {
		return fn(q)
	}

	tx, err := q.conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}

	queriesErr := fn(&Store{
		Queries: q.WithTx(tx),
	})
	if queriesErr != nil {
		if err = tx.Rollback(ctx); err != nil {
			return queriesErr
		}
		return queriesErr
	}

	return tx.Commit(ctx)
}

// PutChar a char in the database
func (q *Store) PutChar(ctx context.Context, userID corde.Snowflake, c discord.Character) error {
	return q.asTx(ctx, func(q *Store) error {
		_, err := q.getUser(ctx, uint64(userID))
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				if err := q.createUser(ctx, uint64(userID)); err != nil {
					return err
				}
			} else {
				return err
			}
		}

		p := insertCharParams{
			ID:     c.ID,
			UserID: uint64(c.UserID),
			Image:  c.Image,
			Name:   strings.Join(strings.Fields(c.Name), " "),
			Type:   c.Type,
		}

		return q.insertChar(ctx, p)
	})
}

func (q *Store) SetUserAnilistURL(ctx context.Context, userID corde.Snowflake, url string) error {
	return q.updateUser(ctx, userID, withAnilistURL(url))
}

// Chars returns the user's characters
func (q *Store) Chars(ctx context.Context, userID corde.Snowflake) ([]discord.Character, error) {
	dbchs, err := q.getChars(ctx, uint64(userID))
	if err != nil {
		return nil, err
	}

	chars := make([]discord.Character, 0, len(dbchs))
	for _, c := range dbchs {
		chars = append(chars, discord.Character{
			Date:   c.Date.Time,
			Image:  c.Image,
			Name:   c.Name,
			Type:   c.Type,
			UserID: corde.Snowflake(c.UserID),
			ID:     c.ID,
		})
	}

	return chars, nil
}

// CharsIDs returns the user's character's ID
func (q *Store) CharsIDs(ctx context.Context, userID corde.Snowflake) ([]int64, error) {
	return q.getCharsID(ctx, uint64(userID))
}

// User returns a user
func (q *Store) User(ctx context.Context, userID corde.Snowflake) (discord.User, error) {
	u, err := q.getUser(ctx, uint64(userID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			if err := q.createUser(ctx, uint64(userID)); err != nil {
				return discord.User{}, err
			}
		} else {
			return discord.User{}, err
		}
	}

	return discord.User{
		Date:     u.Date.Time,
		Quote:    u.Quote,
		UserID:   corde.Snowflake(u.UserID),
		Favorite: uint64(u.Favorite.Int64),
		Tokens:   u.Tokens,
	}, nil
}

// updateUser updates a user's properties
func (q *Store) updateUser(ctx context.Context, userID corde.Snowflake, opts ...func(*squirrel.UpdateBuilder)) error {
	return q.asTx(ctx, func(q *Store) error {
		_, err := q.getUser(ctx, uint64(userID))
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				if err := q.createUser(ctx, uint64(userID)); err != nil {
					return err
				}
			} else {
				return err
			}
		}

		builder := squirrel.Update("users").Where(squirrel.Eq{
			"user_id": userID,
		}).PlaceholderFormat(squirrel.Dollar)

		for _, opt := range opts {
			opt(&builder)
		}

		str, args, err := builder.ToSql()
		if err != nil {
			return err
		}

		if _, err := q.db.Exec(ctx, str, args...); err != nil {
			return err
		}

		return nil
	})
}

// withFavorite sets user favorite
func withFav(f int64) func(*squirrel.UpdateBuilder) {
	return func(s *squirrel.UpdateBuilder) {
		*s = s.Set("favorite", f)
	}
}

// withQuote sets user quote
func withQuote(q string) func(*squirrel.UpdateBuilder) {
	return func(s *squirrel.UpdateBuilder) {
		*s = s.Set("quote", q)
	}
}

// withDate sets the date
func withDate(d time.Time) func(*squirrel.UpdateBuilder) {
	return func(s *squirrel.UpdateBuilder) {
		*s = s.Set("date", d.UTC())
	}
}

// withAnilistURL sets the anilist url
func withAnilistURL(url string) func(*squirrel.UpdateBuilder) {
	return func(s *squirrel.UpdateBuilder) {
		*s = s.Set("anilist_url", url)
	}
}

// withToken sets the token
func withToken(t string) func(*squirrel.UpdateBuilder) {
	return func(s *squirrel.UpdateBuilder) {
		*s = s.Set("token", t)
	}
}

// SetUserDate sets the user's date
func (q *Store) SetUserDate(ctx context.Context, userID corde.Snowflake, d time.Time) error {
	return q.updateUser(ctx, userID, withDate(d))
}

// SetUserToken sets the user's token
func (q *Store) SetUserToken(ctx context.Context, userID corde.Snowflake, token string) error {
	return q.updateUser(ctx, userID, withToken(token))
}

// SetUserFavorite sets the user's favorite
func (q *Store) SetUserFavorite(ctx context.Context, userID corde.Snowflake, c int64) error {
	return q.updateUser(ctx, userID, withFav(c))
}

// SetUserQuote sets the user's quote
func (q *Store) SetUserQuote(ctx context.Context, userID corde.Snowflake, quote string) error {
	return q.updateUser(ctx, userID, withQuote(quote))
}

// CharsStartingWith returns characters starting with the given string
func (q *Store) CharsStartingWith(ctx context.Context, userID corde.Snowflake, s string) ([]discord.Character, error) {
	_, err := q.getUser(ctx, uint64(userID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			if err := q.createUser(ctx, uint64(userID)); err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	dbchs, err := q.getCharsWhoseIDStartWith(ctx, getCharsWhoseIDStartWithParams{
		UserID:  uint64(userID),
		Lim:     50,
		Off:     0,
		LikeStr: s + "%",
	})
	if err != nil {
		return nil, err
	}

	chars := make([]discord.Character, 0, len(dbchs))
	for _, c := range dbchs {
		chars = append(chars, discord.Character{
			Date:   c.Date.Time,
			Image:  c.Image,
			Name:   c.Name,
			Type:   c.Type,
			UserID: corde.Snowflake(c.UserID),
			ID:     c.ID,
		})
	}

	return chars, nil
}

// Profile returns the user's profile
func (q *Store) ProfileOverview(ctx context.Context, userID corde.Snowflake) (discord.Profile, error) {
	_, err := q.getUser(ctx, uint64(userID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			if err := q.createUser(ctx, uint64(userID)); err != nil {
				return discord.Profile{}, err
			}
		} else {
			return discord.Profile{}, err
		}
	}

	p, err := q.getProfileOverview(ctx, uint64(userID))
	if err != nil {
		return discord.Profile{}, err
	}

	return discord.Profile{
		User: discord.User{
			Date:       p.UserDate.Time,
			Quote:      p.UserQuote,
			UserID:     corde.Snowflake(p.UserID),
			Tokens:     p.UserTokens,
			AnilistURL: p.UserAnilistUrl,
			Favorite:   uint64(p.FavoriteID.Int64),
		},
		CharacterCount: int(p.Count),
		Favorite: discord.Character{
			Image:  p.FavoriteImage.String,
			Name:   p.FavoriteName.String,
			UserID: userID,
			ID:     p.FavoriteID.Int64,
		},
	}, nil
}

func (q *Store) GiveUserChar(ctx context.Context, dst corde.Snowflake, src corde.Snowflake, charID int64) error {
	_, err := q.giveChar(ctx, giveCharParams{
		Given: uint64(dst),
		ID:    charID,
		Giver: uint64(src),
	})
	return err
}

func (q *Store) VerifyChar(ctx context.Context, userID corde.Snowflake, charID int64) (discord.Character, error) {
	c, err := q.getChar(ctx, getCharParams{
		ID:     charID,
		UserID: uint64(userID),
	})
	if err != nil {
		return discord.Character{}, err
	}

	return discord.Character{
		Date:   c.Date.Time,
		Image:  c.Image,
		Name:   c.Name,
		Type:   c.Type,
		UserID: corde.Snowflake(c.UserID),
		ID:     c.ID,
	}, nil
}

func (q *Store) ConsumeDropTokens(ctx context.Context, userID corde.Snowflake, count int32) (discord.User, error) {
	u, err := q.consumeDropTokens(ctx, consumeDropTokensParams{
		Tokens: count,
		UserID: uint64(userID),
	})
	if err != nil {
		return discord.User{}, err
	}
	return discord.User{
		Date:     u.Date.Time,
		Quote:    u.Quote,
		Favorite: uint64(u.Favorite.Int64),
		UserID:   userID,
		Tokens:   u.Tokens,
	}, nil
}

func (q *Store) AddDropToken(ctx context.Context, userID corde.Snowflake) error {
	return q.addDropToken(ctx, uint64(userID))
}

func (q *Store) DeleteChar(ctx context.Context, userID corde.Snowflake, charID int64) (discord.Character, error) {
	c, err := q.deleteChar(ctx, deleteCharParams{
		UserID: uint64(userID),
		ID:     charID,
	})
	if err != nil {
		return discord.Character{}, err
	}

	return discord.Character{
		Date:   c.Date.Time,
		Image:  c.Image,
		Name:   c.Name,
		Type:   c.Type,
		UserID: corde.Snowflake(c.UserID),
		ID:     c.ID,
	}, nil
}

type Profile struct {
	ID         uint64 `json:"id,string"`
	Quote      string `json:"quote,omitempty"`
	Tokens     int32  `json:"tokens,omitempty"`
	AnilistURL string `json:"anilist_url,omitempty"`
	Favorite   Char   `json:"favorite,omitzero"`
	Waifus     []Char `json:"waifus,omitempty"`
}

type Char struct {
	Date  time.Time `json:"date"`
	Name  string    `json:"name"`
	Image string    `json:"image"`
	Type  string    `json:"type"`
	ID    int64     `json:"id"`
}

func (q *Store) Profile(ctx context.Context, userID uint64) (*Profile, error) {
	p, err := q.getProfile(ctx, userID)
	if err != nil {
		return nil, err
	}

	chars, err := q.getList(ctx, userID)
	if err != nil {
		return nil, err
	}

	return mapUser(p, chars), nil
}

func (q *Store) UserByAnilistURL(ctx context.Context, anilistURL string) (*User, error) {
	u, err := q.getUserByAnilist(ctx, anilistURL)
	if err != nil {
		return nil, err
	}

	return &User{
		UserID:     u.UserID,
		Quote:      u.Quote,
		Date:       u.RollDate,
		Favorite:   u.Favorite,
		Tokens:     u.Tokens,
		AnilistUrl: u.AnilistUrl,
	}, nil
}

func mapUser(user getProfileRow, list []Character) *Profile {
	p := &Profile{
		ID:         user.UserID,
		Quote:      user.Quote,
		Tokens:     user.Tokens,
		AnilistURL: user.AnilistUrl,
		Waifus:     make([]Char, len(list)),
	}

	for i, u := range list {
		if user.Favorite.Int64 == u.ID {
			p.Favorite = Char{
				ID:    u.ID,
				Name:  u.Name,
				Image: u.Image,
				Type:  u.Type,
				Date:  u.Date.Time,
			}
		}

		p.Waifus[i] = Char{
			ID:    u.ID,
			Name:  u.Name,
			Image: u.Image,
			Type:  u.Type,
			Date:  u.Date.Time,
		}
	}

	return p
}
