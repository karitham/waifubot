package storage

//go:generate mockgen -source=store.go -destination=mocks/store_mock.go -package=mocks -mock_names=Store=MockStorageStore,TXer=MockTXer
//go:generate mockgen -source=userstore/querier.go -destination=mocks/userstore_mock.go -package=mocks -mock_names=Querier=MockUserQuerier

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/tracelog"

	"github.com/karitham/waifubot/storage/collectionstore"
	"github.com/karitham/waifubot/storage/guildstore"
	"github.com/karitham/waifubot/storage/userstore"
	"github.com/karitham/waifubot/storage/wishliststore"
)

type TXer interface {
	Begin(context.Context) (pgx.Tx, error)
	userstore.DBTX
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

type Store interface {
	UserStore() userstore.Querier
	CollectionStore() collectionstore.Querier
	GuildStore() guildstore.Querier
	WishlistStore() wishliststore.Querier
	Tx(ctx context.Context) (Store, error)
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

type DBStore struct {
	userStore       *userstore.Queries
	collectionStore *collectionstore.Queries
	guildStore      *guildstore.Queries
	wishlistStore   *wishliststore.Queries
	db              TXer
	tx              pgx.Tx
}

func NewStore(ctx context.Context, url string) (*DBStore, error) {
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, fmt.Errorf("couldn't parse connect url: %w", err)
	}

	cfg.ConnConfig.Tracer = &tracelog.TraceLog{
		Logger:   newLogger(slog.Default()),
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

	return &DBStore{
		userStore:       userstore.New(conn),
		collectionStore: collectionstore.New(conn),
		guildStore:      guildstore.New(conn),
		wishlistStore:   wishliststore.New(conn),
		db:              conn,
	}, nil
}

func (s *DBStore) withTx(tx pgx.Tx) *DBStore {
	return &DBStore{
		userStore:       s.userStore.WithTx(tx),
		collectionStore: s.collectionStore.WithTx(tx),
		guildStore:      s.guildStore.WithTx(tx),
		wishlistStore:   s.wishlistStore.WithTx(tx),
		db:              tx,
		tx:              tx,
	}
}

type Logger struct {
	logger *slog.Logger
}

func newLogger(logger *slog.Logger) *Logger {
	return &Logger{
		logger: logger.With("module", "pgx"),
	}
}

func (l *Logger) Log(ctx context.Context, level tracelog.LogLevel, msg string, data map[string]any) {
	logger := l.logger
	attrs := make([]slog.Attr, 0, len(data))
	for k, v := range data {
		attrs = append(attrs, slog.Any(k, v))
	}

	m := strings.ReplaceAll(msg, "\n", " ")
	formattedMsg := strings.ReplaceAll(m, "\t", " ")

	logger.LogAttrs(ctx, translateLevel(level), formattedMsg, attrs...)
}

func translateLevel(level tracelog.LogLevel) slog.Level {
	switch level {
	case tracelog.LogLevelNone:
		return slog.Level(-8) // LevelDebug-1, equivalent to NoLevel
	case tracelog.LogLevelError:
		return slog.LevelError
	case tracelog.LogLevelWarn:
		return slog.LevelWarn
	case tracelog.LogLevelInfo:
		return slog.LevelInfo
	case tracelog.LogLevelDebug, tracelog.LogLevelTrace:
		return slog.LevelDebug
	default:
		return slog.LevelError
	}
}

func (s *DBStore) UserStore() userstore.Querier {
	return s.userStore
}

func (s *DBStore) CollectionStore() collectionstore.Querier {
	return s.collectionStore
}

func (s *DBStore) GuildStore() guildstore.Querier {
	return s.guildStore
}

func (s *DBStore) WishlistStore() wishliststore.Querier {
	return s.wishlistStore
}

func (s *DBStore) Tx(ctx context.Context) (Store, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}

	return s.withTx(tx), nil
}

func (s *DBStore) Commit(ctx context.Context) error {
	return s.tx.Commit(ctx)
}

func (s *DBStore) Rollback(ctx context.Context) error {
	return s.tx.Rollback(ctx)
}
