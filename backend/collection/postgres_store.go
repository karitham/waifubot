package collection

import (
	"context"

	"github.com/jackc/pgx/v5"

	"github.com/karitham/waifubot/catalog"
	"github.com/karitham/waifubot/guild"
)

// TxFn creates tx-scoped query adapters from a pgx transaction.
// Registered by the wiring layer (cmd/) so collection doesn't import
// storage/*pg. It must build every adapter over the transaction, never
// over the pool, or the "transaction" is a no-op.
type TxFn func(tx pgx.Tx) Store

// pooler is the minimal interface for beginning transactions.
type pooler interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}

// composedStore is the query surface assembled from repository adapters.
type composedStore struct {
	UserRepository
	CollectionRepository
	DropRepository
	guild.Querier
	catalog.Store
}

// NewStore composes repository adapters into a query Store.
func NewStore(
	user UserRepository,
	coll CollectionRepository,
	drop DropRepository,
	guild guild.Querier,
	cat catalog.Store,
) Store {
	return composedStore{
		UserRepository:       user,
		CollectionRepository: coll,
		DropRepository:       drop,
		Querier:              guild,
		Store:                cat,
	}
}

// PostgresStore is the pool-level DB: repository adapters plus transaction
// creation.
type PostgresStore struct {
	Store
	pool pooler
	txFn TxFn
}

// NewPostgresStore creates a pool-level DB from a query Store and a tx
// adapter factory.
func NewPostgresStore(store Store, pool pooler, txFn TxFn) *PostgresStore {
	return &PostgresStore{
		Store: store,
		pool:  pool,
		txFn:  txFn,
	}
}

// WithTx starts a new transaction and returns a Store scoped to it.
func (s *PostgresStore) WithTx(ctx context.Context) (Tx, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	return NewTx(s.txFn(tx), tx), nil
}

// PostgresTx is a Store scoped to a pgx transaction.
type PostgresTx struct {
	Store
	tx pgx.Tx
}

// NewTx wraps tx-scoped adapters with commit and rollback control.
func NewTx(store Store, tx pgx.Tx) Tx {
	return &PostgresTx{Store: store, tx: tx}
}

// Commit commits the transaction.
func (t *PostgresTx) Commit(ctx context.Context) error {
	return t.tx.Commit(ctx)
}

// Rollback rolls back the transaction.
func (t *PostgresTx) Rollback(ctx context.Context) error {
	return t.tx.Rollback(ctx)
}
