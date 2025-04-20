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
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/tracelog"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/karitham/waifubot/db/characters"
	"github.com/karitham/waifubot/db/users"
	"github.com/karitham/waifubot/discord"
)

type TXer interface {
	Begin(context.Context) (pgx.Tx, error)
	users.DBTX
}

type Store struct {
	UserStore      *users.Queries
	CharacterStore *characters.Queries
	db             TXer
}

func NewStore(ctx context.Context, url string) (*Store, error) {
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, fmt.Errorf("couldn't parse connect url: %w", err)
	}

	cfg.ConnConfig.Tracer = &tracelog.TraceLog{
		Logger:   newLogger(log.Logger),
		LogLevel: tracelog.LogLevelDebug,
	}

	conn, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("couldn't connect to db: %w", err)
	}

	if err := conn.Ping(ctx); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Store{
		UserStore:      users.New(conn),
		CharacterStore: characters.New(conn),
		db:             conn,
	}, nil
}

func (s *Store) withTx(tx pgx.Tx) *Store {
	return &Store{
		UserStore:      s.UserStore.WithTx(tx),
		CharacterStore: s.CharacterStore.WithTx(tx),
		db:             tx,
	}
}

type Logger struct {
	logger zerolog.Logger
}

func newLogger(logger zerolog.Logger) *Logger {
	return &Logger{
		logger: logger.With().Str("module", "pgx").Logger(),
	}
}

func (pl *Logger) Log(ctx context.Context, level tracelog.LogLevel, msg string, data map[string]any) {
	var zlevel zerolog.Level
	switch level {
	case tracelog.LogLevelNone:
		zlevel = zerolog.NoLevel
	case tracelog.LogLevelError:
		zlevel = zerolog.ErrorLevel
	case tracelog.LogLevelWarn:
		zlevel = zerolog.WarnLevel
	case tracelog.LogLevelInfo:
		zlevel = zerolog.InfoLevel
	case tracelog.LogLevelDebug:
		zlevel = zerolog.DebugLevel
	default:
		zlevel = zerolog.DebugLevel
	}

	pgxlog := pl.logger.With().Fields(data).Logger()
	pgxlog.WithLevel(zlevel).Msg(msg)
}

func (s *Store) Tx(ctx context.Context, fn func(store discord.Store) error) error {
	return s.asTx(ctx, func(q *Store) error {
		return fn(q)
	})
}

func (s *Store) asTx(ctx context.Context, fn func(q *Store) error) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	if err := fn(s.withTx(tx)); err != nil {
		if rerr := tx.Rollback(context.Background()); rerr != nil {
			return fmt.Errorf("transaction function failed: %w (rollback also failed: %v)", err, rerr)
		}

		return err
	}

	return tx.Commit(ctx)
}

func userToDiscordUser(u users.User) discord.User {
	return discord.User{
		Date:       u.Date.Time,
		Quote:      u.Quote,
		UserID:     corde.Snowflake(u.UserID),
		Favorite:   uint64(u.Favorite.Int64),
		Tokens:     u.Tokens,
		AnilistURL: u.AnilistUrl,
	}
}

func charToDiscordChar(c characters.Character) discord.Character {
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
	err := s.CharacterStore.Insert(ctx, characters.InsertParams{
		ID:     c.ID,
		UserID: uint64(c.UserID),
		Image:  c.Image,
		Name:   strings.Join(strings.Fields(c.Name), " "),
		Type:   c.Type,
	})
	if err != nil {
		return fmt.Errorf("failed to insert char %d for user %d: %w", c.ID, userID, err)
	}

	return nil
}

func ensureUserExists(ctx context.Context, userID corde.Snowflake, us *users.Queries) error {
	_, err := us.Get(ctx, uint64(userID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			if createErr := us.Create(ctx, uint64(userID)); createErr != nil {
				return fmt.Errorf("failed to create user %d for put char: %w", userID, createErr)
			}
		} else {
			return fmt.Errorf("failed to get user %d for put char: %w", userID, err)
		}
	}

	return nil
}

func (s *Store) updateUser(ctx context.Context, userID corde.Snowflake, opts ...func(*squirrel.UpdateBuilder)) error {
	if err := ensureUserExists(ctx, userID, s.UserStore); err != nil {
		return err
	}

	builder := squirrel.Update("users").
		Where(squirrel.Eq{"user_id": userID}).
		PlaceholderFormat(squirrel.Dollar)

	for _, opt := range opts {
		opt(&builder)
	}

	sqlStr, args, err := builder.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build update query for user %d: %w", userID, err)
	}

	if _, err := s.db.Exec(ctx, sqlStr, args...); err != nil {
		return fmt.Errorf("failed to execute update for user %d: %w", userID, err)
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
	dbchs, err := s.CharacterStore.List(ctx, uint64(userID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []discord.Character{}, nil
		}

		return nil, fmt.Errorf("failed to get chars for user %d: %w", userID, err)
	}

	chars := make([]discord.Character, 0, len(dbchs))
	for _, c := range dbchs {
		chars = append(chars, charToDiscordChar(c))
	}

	return chars, nil
}

func (s *Store) CharsIDs(ctx context.Context, userID corde.Snowflake) ([]int64, error) {
	ids, err := s.CharacterStore.ListIDs(ctx, uint64(userID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
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
	if err := ensureUserExists(ctx, userID, s.UserStore); err != nil {
		return discord.User{}, err
	}

	u, err := s.UserStore.Get(ctx, uint64(userID))
	return userToDiscordUser(u), err
}

func (s *Store) CharsStartingWith(ctx context.Context, userID corde.Snowflake, prefix string) ([]discord.Character, error) {
	if err := ensureUserExists(ctx, userID, s.UserStore); err != nil {
		return nil, err
	}

	dbchs, err := s.CharacterStore.ListFilterIDPrefix(ctx, characters.ListFilterIDPrefixParams{
		UserID:   uint64(userID),
		Lim:      25,
		IDPrefix: "%" + prefix,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []discord.Character{}, nil
		}

		return nil, fmt.Errorf("failed to get chars starting with '%s' for user %d: %w", prefix, userID, err)
	}

	chars := make([]discord.Character, 0, len(dbchs))
	for _, c := range dbchs {
		chars = append(chars, charToDiscordChar(c))
	}

	return chars, nil
}

func (s *Store) ProfileOverview(ctx context.Context, userID corde.Snowflake) (discord.Profile, error) {
	if err := ensureUserExists(ctx, userID, s.UserStore); err != nil {
		return discord.Profile{}, err
	}

	profile := discord.Profile{}

	user, err := s.UserStore.Get(ctx, uint64(userID))
	if err != nil {
		return discord.Profile{}, fmt.Errorf("couldn't fetch user: %w", err)
	}

	if user.Favorite.Valid {
		fav, err := s.CharacterStore.Get(ctx, characters.GetParams{
			ID:     user.Favorite.Int64,
			UserID: user.UserID,
		})
		if err != nil {
			return discord.Profile{}, fmt.Errorf("couldn't fetch favorite: %w", err)
		}

		profile.Favorite = charToDiscordChar(fav)
	}

	count, err := s.CharacterStore.Count(ctx, user.UserID)
	if err != nil {
		return discord.Profile{}, fmt.Errorf("couldn't count chars: %w", err)
	}

	profile.CharacterCount = int(count)
	profile.User = userToDiscordUser(user)

	return profile, nil
}

func (s *Store) GiveUserChar(ctx context.Context, dst corde.Snowflake, src corde.Snowflake, charID int64) error {
	_, err := s.CharacterStore.Give(ctx, characters.GiveParams{
		UserID:   uint64(dst),
		UserID_2: uint64(src),
		ID:       charID,
	})
	if err != nil {
		return fmt.Errorf("failed to give char %d from %d to %d: %w", charID, src, dst, err)
	}

	return nil
}

func (s *Store) VerifyChar(ctx context.Context, userID corde.Snowflake, charID int64) (discord.Character, error) {
	c, err := s.CharacterStore.Get(ctx, characters.GetParams{
		ID:     charID,
		UserID: uint64(userID),
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return discord.Character{}, fmt.Errorf("character %d not found or does not belong to user %d", charID, userID)
		}
		return discord.Character{}, fmt.Errorf("failed to verify char %d for user %d: %w", charID, userID, err)
	}

	return charToDiscordChar(c), nil
}

func (s *Store) ConsumeDropTokens(ctx context.Context, userID corde.Snowflake, count int32) (discord.User, error) {
	if count <= 0 {
		return discord.User{}, errors.New("token count to consume must be positive")
	}

	u, err := s.UserStore.ConsumeTokens(ctx, users.ConsumeTokensParams{
		Tokens: count,
		UserID: uint64(userID),
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			currentUser, userErr := s.User(ctx, userID)
			if userErr == nil {
				return discord.User{}, fmt.Errorf("insufficient tokens for user %d: requires %d, has %d", userID, count, currentUser.Tokens)
			}

			return discord.User{}, fmt.Errorf("insufficient tokens for user %d (requires %d): %w", userID, count, err)
		}

		return discord.User{}, fmt.Errorf("failed to consume tokens for user %d: %w", userID, err)
	}

	return userToDiscordUser(u), nil
}

func (s *Store) AddDropToken(ctx context.Context, userID corde.Snowflake) error {
	if err := ensureUserExists(ctx, userID, s.UserStore); err != nil {
		return err
	}

	return s.UserStore.IncTokens(ctx, uint64(userID))
}

func (s *Store) DeleteChar(ctx context.Context, userID corde.Snowflake, charID int64) (discord.Character, error) {
	c, err := s.CharacterStore.Delete(ctx, characters.DeleteParams{
		UserID: uint64(userID),
		ID:     charID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return discord.Character{}, fmt.Errorf("character %d not found or does not belong to user %d", charID, userID)
		}

		return discord.Character{}, fmt.Errorf("failed to delete char %d for user %d: %w", charID, userID, err)
	}

	return charToDiscordChar(c), nil
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

func mapUser(user users.User, list []characters.Character) *APIProfile {
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
	p, err := s.UserStore.Get(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user %d not found for profile API", userID)
		}

		return nil, fmt.Errorf("failed to get profile data for user %d: %w", userID, err)
	}

	chars, err := s.CharacterStore.List(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			chars = []characters.Character{}
		} else {
			return nil, fmt.Errorf("failed to get character list for user %d: %w", userID, err)
		}
	}

	return mapUser(p, chars), nil
}

func (s *Store) UserByAnilistURL(ctx context.Context, anilistURL string) (discord.User, error) {
	u, err := s.UserStore.GetByAnilist(ctx, anilistURL)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return discord.User{}, fmt.Errorf("no user found with Anilist URL: %s", anilistURL)
		}

		return discord.User{}, fmt.Errorf("failed to get user by Anilist URL: %w", err)
	}

	return userToDiscordUser(u), nil
}
