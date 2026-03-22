package collection

import (
	"context"

	"github.com/jackc/pgx/v5"

	"github.com/karitham/waifubot/catalog"
)

// TxFn creates a transactional Store from a pgx transaction.
// Registered by the wiring layer (cmd/) so collection doesn't import storage/*pg.
type TxFn func(tx pgx.Tx) Store

// pooler is the minimal interface for beginning transactions.
type pooler interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}

// PostgresStore composes per-domain adapters into the transactional Store interface.
// Non-transactional instances hold pool-level adapters and a TxFn.
// Transactional instances hold tx-level adapters and a pgx.Tx for Commit/Rollback.
type PostgresStore struct {
	UserRepository
	CollectionRepository
	DropRepository
	GuildQuerier
	catalog.Store

	db   pooler // connection pool (non-tx) or pgx.Tx (tx)
	txFn TxFn   // creates tx-scoped stores
}

// NewPostgresStore creates a non-transactional Store from individual adapters.
func NewPostgresStore(
	user UserRepository,
	coll CollectionRepository,
	drop DropRepository,
	guild GuildQuerier,
	cat catalog.Store,
	db pooler,
	txFn TxFn,
) *PostgresStore {
	return &PostgresStore{
		UserRepository:       user,
		CollectionRepository: coll,
		DropRepository:       drop,
		GuildQuerier:         guild,
		Store:                cat,
		db:                   db,
		txFn:                 txFn,
	}
}

// WithTx starts a new transaction and returns a Store scoped to it.
func (s *PostgresStore) WithTx(ctx context.Context) (Store, error) {
	tx, err := s.pool().Begin(ctx)
	if err != nil {
		return nil, err
	}
	return s.txFn(tx), nil
}

// Commit commits the transaction. Only valid on transactional instances.
func (s *PostgresStore) Commit(ctx context.Context) error {
	return s.db.(pgx.Tx).Commit(ctx)
}

// Rollback rolls back the transaction. Only valid on transactional instances.
func (s *PostgresStore) Rollback(ctx context.Context) error {
	return s.db.(pgx.Tx).Rollback(ctx)
}

// pool returns the underlying connection pool.
func (s *PostgresStore) pool() pooler {
	return s.db
}
