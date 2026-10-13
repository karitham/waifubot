package storage

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/tracelog"
)

type TXer interface {
	Begin(context.Context) (pgx.Tx, error)
	DBTX
}

// Store is the database handle: queries, the raw connection, and transaction
// creation. A Store returned by Tx is scoped to that transaction.
type Store interface {
	Querier() Querier
	DB() TXer
	Tx(ctx context.Context) (Store, error)
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

type DBStore struct {
	q  *Queries
	db TXer
	tx pgx.Tx
}

func NewStore(ctx context.Context, url string) (Store, error) {
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

	return &DBStore{q: New(conn), db: conn}, nil
}

func (s *DBStore) withTx(tx pgx.Tx) *DBStore {
	return &DBStore{
		q:  s.q.WithTx(tx),
		db: tx,
		tx: tx,
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

func (s *DBStore) Querier() Querier {
	return s.q
}

func (s *DBStore) DB() TXer {
	return s.db
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
