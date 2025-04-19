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

type Store struct {
	*Queries
	conn *pgx.Conn
}

func NewStore(ctx context.Context, url string) (*Store, error) {
	conn, err := pgx.Connect(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("couldn't connect to db: %w", err)
	}

	if err := conn.Ping(ctx); err != nil {
		conn.Close(ctx)
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Store{
		Queries: New(conn),
		conn:    conn,
	}, nil
}

func (s *Store) Close(ctx context.Context) error {
	if s.conn != nil {
		return s.conn.Close(ctx)
	}

	return nil
}

func (s *Store) Tx(ctx context.Context, fn func(store discord.Store) error) error {
	return s.asTx(ctx, func(q *Store) error {
		return fn(q)
	})
}

func (s *Store) asTx(ctx context.Context, fn func(q *Store) error) error {
	if s.conn == nil {
		return fn(s)
	}

	tx, err := s.conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	txStore := &Store{
		Queries: s.WithTx(tx),
		conn:    nil,
	}

	queriesErr := fn(txStore)
	if queriesErr != nil {
		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
			return fmt.Errorf("transaction function failed: %w (rollback also failed: %v)", queriesErr, rollbackErr)
		}
		return queriesErr
	}

	return tx.Commit(ctx)
}

func mapGetUserRowToDiscordUser(u User) discord.User {
	return discord.User{
		Date:       u.Date.Time,
		Quote:      u.Quote,
		UserID:     corde.Snowflake(u.UserID),
		Favorite:   uint64(u.Favorite.Int64),
		Tokens:     u.Tokens,
		AnilistURL: u.AnilistUrl,
	}
}

func mapCharacterRowToDiscordCharacter(c Character) discord.Character {
	return discord.Character{
		Date:   c.Date.Time,
		Image:  c.Image,
		Name:   c.Name,
		Type:   c.Type,
		UserID: corde.Snowflake(c.UserID),
		ID:     c.ID,
	}
}

func (s *Store) PutChar(ctx context.Context, userID corde.Snowflake, c discord.Character) error {
	uid := uint64(userID)
	_, err := s.getUser(ctx, uid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) || errors.Is(err, sql.ErrNoRows) {
			if createErr := s.createUser(ctx, uid); createErr != nil {
				return fmt.Errorf("failed to create user %d for put char: %w", uid, createErr)
			}
		} else {
			return fmt.Errorf("failed to get user %d for put char: %w", uid, err)
		}
	}

	p := insertCharParams{
		ID:     c.ID,
		UserID: uint64(c.UserID),
		Image:  c.Image,
		Name:   strings.Join(strings.Fields(c.Name), " "),
		Type:   c.Type,
	}

	err = s.insertChar(ctx, p)
	if err != nil {
		return fmt.Errorf("failed to insert char %d for user %d: %w", c.ID, userID, err)
	}
	return nil
}

func (s *Store) updateUser(ctx context.Context, userID corde.Snowflake, opts ...func(*squirrel.UpdateBuilder)) error {
	uid := uint64(userID)

	_, err := s.getUser(ctx, uid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) || errors.Is(err, sql.ErrNoRows) {
			if createErr := s.createUser(ctx, uid); createErr != nil {
				return fmt.Errorf("failed to create user %d for update: %w", uid, createErr)
			}
		} else {
			return fmt.Errorf("failed to get user %d for update: %w", uid, err)
		}
	}

	builder := squirrel.Update("users").
		Where(squirrel.Eq{"user_id": uid}).
		PlaceholderFormat(squirrel.Dollar)

	for _, opt := range opts {
		opt(&builder)
	}

	sqlStr, args, err := builder.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build update query for user %d: %w", uid, err)
	}

	if _, err := s.db.Exec(ctx, sqlStr, args...); err != nil {
		return fmt.Errorf("failed to execute update for user %d: %w", uid, err)
	}

	return nil
}

func withFav(f int64) func(*squirrel.UpdateBuilder) {
	return func(s *squirrel.UpdateBuilder) {
		*s = s.Set("favorite", f)
	}
}

func withQuote(q string) func(*squirrel.UpdateBuilder) {
	return func(s *squirrel.UpdateBuilder) {
		*s = s.Set("quote", q)
	}
}

func withDate(d time.Time) func(*squirrel.UpdateBuilder) {
	return func(s *squirrel.UpdateBuilder) {
		*s = s.Set("date", d.UTC())
	}
}

func withAnilistURL(url string) func(*squirrel.UpdateBuilder) {
	return func(s *squirrel.UpdateBuilder) {
		*s = s.Set("anilist_url", url)
	}
}

func (s *Store) SetUserAnilistURL(ctx context.Context, userID corde.Snowflake, url string) error {
	return s.updateUser(ctx, userID, withAnilistURL(url))
}

func (s *Store) SetUserDate(ctx context.Context, userID corde.Snowflake, d time.Time) error {
	return s.updateUser(ctx, userID, withDate(d))
}

func (s *Store) SetUserFavorite(ctx context.Context, userID corde.Snowflake, c int64) error {
	return s.updateUser(ctx, userID, withFav(c))
}

func (s *Store) SetUserQuote(ctx context.Context, userID corde.Snowflake, quote string) error {
	return s.updateUser(ctx, userID, withQuote(quote))
}

func (s *Store) Chars(ctx context.Context, userID corde.Snowflake) ([]discord.Character, error) {
	dbchs, err := s.getChars(ctx, uint64(userID))
	if err != nil {

		if errors.Is(err, pgx.ErrNoRows) || errors.Is(err, sql.ErrNoRows) {
			return []discord.Character{}, nil
		}
		return nil, fmt.Errorf("failed to get chars for user %d: %w", userID, err)
	}

	chars := make([]discord.Character, 0, len(dbchs))
	for _, c := range dbchs {
		chars = append(chars, mapCharacterRowToDiscordCharacter(c))
	}

	return chars, nil
}

func (s *Store) CharsIDs(ctx context.Context, userID corde.Snowflake) ([]int64, error) {
	ids, err := s.getCharsID(ctx, uint64(userID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) || errors.Is(err, sql.ErrNoRows) {
			return []int64{}, nil
		}
		return nil, fmt.Errorf("failed to get char IDs for user %d: %w", userID, err)
	}
	if ids == nil {
		return []int64{}, nil
	}
	return ids, nil
}

func (s *Store) User(ctx context.Context, userID corde.Snowflake) (discord.User, error) {
	uid := uint64(userID)

	u, err := s.getUser(ctx, uid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) || errors.Is(err, sql.ErrNoRows) {

			if createErr := s.createUser(ctx, uid); createErr != nil {
				return discord.User{}, fmt.Errorf("failed to create user %d for get user: %w", uid, createErr)
			}

			u, err = s.getUser(ctx, uid)
			if err != nil {
				return discord.User{}, fmt.Errorf("failed to get newly created user %d: %w", uid, err)
			}
		} else {
			return discord.User{}, fmt.Errorf("failed to get user %d: %w", uid, err)
		}
	}

	return mapGetUserRowToDiscordUser(u), nil
}

func (s *Store) CharsStartingWith(ctx context.Context, userID corde.Snowflake, prefix string) ([]discord.Character, error) {
	uid := uint64(userID)

	_, err := s.getUser(ctx, uid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) || errors.Is(err, sql.ErrNoRows) {
			if createErr := s.createUser(ctx, uid); createErr != nil {
				return nil, fmt.Errorf("failed to create user %d for chars starting with: %w", uid, createErr)
			}

			return []discord.Character{}, nil
		} else {
			return nil, fmt.Errorf("failed to get user %d for chars starting with: %w", uid, err)
		}
	}

	params := getCharsWhoseIDStartWithParams{
		UserID:  uid,
		Lim:     25,
		Off:     0,
		LikeStr: prefix + "%",
	}
	dbchs, err := s.getCharsWhoseIDStartWith(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) || errors.Is(err, sql.ErrNoRows) {
			return []discord.Character{}, nil
		}
		return nil, fmt.Errorf("failed to get chars starting with '%s' for user %d: %w", prefix, userID, err)
	}

	chars := make([]discord.Character, 0, len(dbchs))
	for _, c := range dbchs {
		chars = append(chars, mapCharacterRowToDiscordCharacter(c))
	}

	return chars, nil
}

func (s *Store) ProfileOverview(ctx context.Context, userID corde.Snowflake) (discord.Profile, error) {
	uid := uint64(userID)

	_, err := s.getUser(ctx, uid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) || errors.Is(err, sql.ErrNoRows) {
			if createErr := s.createUser(ctx, uid); createErr != nil {
				return discord.Profile{}, fmt.Errorf("failed to create user %d for profile overview: %w", uid, createErr)
			}
		} else {
			return discord.Profile{}, fmt.Errorf("failed to get user %d for profile overview: %w", uid, err)
		}
	}

	p, err := s.getProfileOverview(ctx, uid)
	if err != nil {

		if errors.Is(err, pgx.ErrNoRows) || errors.Is(err, sql.ErrNoRows) {

			u, _ := s.User(ctx, userID)
			return discord.Profile{User: u}, nil
		}
		return discord.Profile{}, fmt.Errorf("failed to get profile overview for user %d: %w", userID, err)
	}

	profile := discord.Profile{
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
			ID:     p.FavoriteID.Int64,
			Image:  p.FavoriteImage.String,
			Name:   p.FavoriteName.String,
			UserID: userID,
		},
	}
	return profile, nil
}

func (s *Store) GiveUserChar(ctx context.Context, dst corde.Snowflake, src corde.Snowflake, charID int64) error {
	params := giveCharParams{
		Given: uint64(dst),
		ID:    charID,
		Giver: uint64(src),
	}

	_, err := s.giveChar(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to give char %d from %d to %d: %w", charID, src, dst, err)
	}

	return nil
}

func (s *Store) VerifyChar(ctx context.Context, userID corde.Snowflake, charID int64) (discord.Character, error) {
	params := getCharParams{
		ID:     charID,
		UserID: uint64(userID),
	}
	c, err := s.getChar(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) || errors.Is(err, sql.ErrNoRows) {
			return discord.Character{}, fmt.Errorf("character %d not found or does not belong to user %d", charID, userID)
		}
		return discord.Character{}, fmt.Errorf("failed to verify char %d for user %d: %w", charID, userID, err)
	}

	return mapCharacterRowToDiscordCharacter(c), nil
}

func (s *Store) ConsumeDropTokens(ctx context.Context, userID corde.Snowflake, count int32) (discord.User, error) {
	if count <= 0 {
		return discord.User{}, errors.New("token count to consume must be positive")
	}

	params := consumeDropTokensParams{
		Tokens: count,
		UserID: uint64(userID),
	}

	u, err := s.consumeDropTokens(ctx, params)
	if err != nil {

		if errors.Is(err, pgx.ErrNoRows) || errors.Is(err, sql.ErrNoRows) {

			currentUser, userErr := s.User(ctx, userID)
			if userErr == nil {
				return discord.User{}, fmt.Errorf("insufficient tokens for user %d: requires %d, has %d", userID, count, currentUser.Tokens)
			}

			return discord.User{}, fmt.Errorf("insufficient tokens for user %d (requires %d): %w", userID, count, err)
		}
		return discord.User{}, fmt.Errorf("failed to consume tokens for user %d: %w", userID, err)
	}

	return mapGetUserRowToDiscordUser(u), nil
}

func (s *Store) AddDropToken(ctx context.Context, userID corde.Snowflake) error {
	uid := uint64(userID)
	_, err := s.getUser(ctx, uid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) || errors.Is(err, sql.ErrNoRows) {
			if createErr := s.createUser(ctx, uid); createErr != nil {
				return fmt.Errorf("failed to create user %d for add token: %w", uid, createErr)
			}
		} else {
			return fmt.Errorf("failed to get user %d for add token: %w", uid, err)
		}
	}

	err = s.addDropToken(ctx, uint64(userID))
	if err != nil {
		return fmt.Errorf("failed to add token for user %d: %w", userID, err)
	}
	return nil
}

func (s *Store) DeleteChar(ctx context.Context, userID corde.Snowflake, charID int64) (discord.Character, error) {
	params := deleteCharParams{
		UserID: uint64(userID),
		ID:     charID,
	}

	c, err := s.deleteChar(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) || errors.Is(err, sql.ErrNoRows) {
			return discord.Character{}, fmt.Errorf("character %d not found or does not belong to user %d", charID, userID)
		}
		return discord.Character{}, fmt.Errorf("failed to delete char %d for user %d: %w", charID, userID, err)
	}

	return mapCharacterRowToDiscordCharacter(c), nil
}

type APIProfile struct {
	ID         uint64    `json:"id,string"`
	Quote      string    `json:"quote,omitempty"`
	Tokens     int32     `json:"tokens,omitempty"`
	AnilistURL string    `json:"anilist_url,omitempty"`
	Favorite   APIChar   `json:"favorite,omitzero"`
	Waifus     []APIChar `json:"waifus,omitempty"`
}

type APIChar struct {
	Date  time.Time `json:"date"`
	Name  string    `json:"name"`
	Image string    `json:"image"`
	Type  string    `json:"type"`
	ID    int64     `json:"id"`
}

func mapUser(user getProfileRow, list []Character) *APIProfile {
	p := &APIProfile{
		ID:         user.UserID,
		Quote:      user.Quote,
		Tokens:     user.Tokens,
		AnilistURL: user.AnilistUrl,
		Waifus:     make([]APIChar, 0, len(list)),
	}

	for _, u := range list {
		apiChar := APIChar{
			ID:    u.ID,
			Name:  u.Name,
			Image: u.Image,
			Type:  u.Type,
			Date:  u.Date.Time,
		}
		p.Waifus = append(p.Waifus, apiChar)

		if user.Favorite.Valid && user.Favorite.Int64 == u.ID {
			p.Favorite = apiChar
		}
	}

	return p
}

func (s *Store) Profile(ctx context.Context, userID uint64) (*APIProfile, error) {
	p, err := s.getProfile(ctx, userID)
	if err != nil {

		if errors.Is(err, pgx.ErrNoRows) || errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user %d not found for profile API", userID)
		}
		return nil, fmt.Errorf("failed to get profile data for user %d: %w", userID, err)
	}

	chars, err := s.getList(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) || errors.Is(err, sql.ErrNoRows) {
			chars = []Character{}
		} else {
			return nil, fmt.Errorf("failed to get character list for user %d: %w", userID, err)
		}
	}

	return mapUser(p, chars), nil
}

func (s *Store) UserByAnilistURL(ctx context.Context, anilistURL string) (discord.User, error) {
	u, err := s.getUserByAnilist(ctx, anilistURL)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) || errors.Is(err, sql.ErrNoRows) {
			return discord.User{}, fmt.Errorf("no user found with Anilist URL: %s", anilistURL)
		}
		return discord.User{}, fmt.Errorf("failed to get user by Anilist URL: %w", err)
	}

	return mapGetUserRowToDiscordUser(u), nil
}
