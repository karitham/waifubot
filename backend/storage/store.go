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

// Store is the pool-level database handle: queries, the raw connection,
// transaction creation, and pool shutdown.
type Store interface {
	Querier() Querier
	DB() TXer
	Tx(ctx context.Context) (TxStore, error)
	Close()
}

// TxStore is a database handle scoped to a transaction.
type TxStore interface {
	Querier() Querier
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

type DBStore struct {
	q    *Queries
	pool *pgxpool.Pool
}

type txStore struct {
	q  *Queries
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

	return &DBStore{q: New(conn), pool: conn}, nil
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
	return s.pool
}

func (s *DBStore) Close() {
	s.pool.Close()
}

func (s *DBStore) Tx(ctx context.Context) (TxStore, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}

	return &txStore{q: s.q.WithTx(tx), tx: tx}, nil
}

func (s *txStore) Querier() Querier {
	return s.q
}

func (s *txStore) Commit(ctx context.Context) error {
	return s.tx.Commit(ctx)
}

func (s *txStore) Rollback(ctx context.Context) error {
	return s.tx.Rollback(ctx)
}
