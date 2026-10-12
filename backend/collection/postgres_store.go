package collection

import (
	"context"

	"github.com/karitham/waifubot/catalog"
	"github.com/karitham/waifubot/guild"
)

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

// TxControl commits or rolls back a transaction.
type TxControl interface {
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

// NewTx wraps tx-scoped query adapters with commit and rollback control.
func NewTx(store Store, control TxControl) Tx {
	return txStore{Store: store, control: control}
}

type txStore struct {
	Store
	control TxControl
}

// Commit commits the transaction.
func (t txStore) Commit(ctx context.Context) error {
	return t.control.Commit(ctx)
}

// Rollback rolls back the transaction.
func (t txStore) Rollback(ctx context.Context) error {
	return t.control.Rollback(ctx)
}

// BeginFunc starts a transaction and returns a Tx scoped to it.
type BeginFunc func(ctx context.Context) (Tx, error)

// PostgresStore is the pool-level DB: repository adapters plus transaction
// creation. The wiring layer supplies BeginFunc so collection doesn't import
// storage packages.
type PostgresStore struct {
	Store
	begin BeginFunc
}

// NewPostgresStore creates a pool-level DB from a query Store and a begin
// function.
func NewPostgresStore(store Store, begin BeginFunc) *PostgresStore {
	return &PostgresStore{
		Store: store,
		begin: begin,
	}
}

// Begin starts a new transaction and returns a Tx scoped to it.
func (s *PostgresStore) Begin(ctx context.Context) (Tx, error) {
	return s.begin(ctx)
}
