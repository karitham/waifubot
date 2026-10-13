// Package pg adapts the generated storage queries to the domain ports.
//
// Pool carries what is safe on a pool handle: reads and single-statement
// writes with no multi-statement invariant. Tx carries the writes that must
// run inside a transaction, plus control. GuildDB exists because guild.DB's
// Begin returns guild.Tx while collection.DB's returns collection.Tx, and Go
// has no covariant returns.
package pg

import (
	"context"

	"github.com/karitham/waifubot/collection"
	"github.com/karitham/waifubot/guild"
	"github.com/karitham/waifubot/storage"
)

// base carries the pool-safe translations: reads and single-statement
// writes.
type base struct {
	q storage.Querier
}

// Pool implements the pool-level ports: collection.DB, catalog.Store,
// guild.Reader, and the command and interaction stores.
type Pool struct {
	base
	s storage.Store
}

func NewPool(s storage.Store) *Pool {
	return &Pool{base: base{q: s.Querier()}, s: s}
}

// Begin starts a transaction and returns it as the collection surface.
func (p *Pool) Begin(ctx context.Context) (collection.Tx, error) {
	scoped, err := p.s.Tx(ctx)
	if err != nil {
		return nil, err
	}
	return NewTx(scoped), nil
}

// Tx implements the transaction-scoped ports: collection.Tx, guild.Tx, and
// catalog.Store.
type Tx struct {
	base
	s storage.TxStore
}

func NewTx(s storage.TxStore) *Tx {
	return &Tx{base: base{q: s.Querier()}, s: s}
}

func (t *Tx) Commit(ctx context.Context) error   { return t.s.Commit(ctx) }
func (t *Tx) Rollback(ctx context.Context) error { return t.s.Rollback(ctx) }

// GuildDB adapts a pool store to guild.DB.
type GuildDB struct {
	base
	s storage.Store
}

func NewGuildDB(s storage.Store) *GuildDB {
	return &GuildDB{base: base{q: s.Querier()}, s: s}
}

// Begin starts a transaction and returns it as the guild surface.
func (g *GuildDB) Begin(ctx context.Context) (guild.Tx, error) {
	scoped, err := g.s.Tx(ctx)
	if err != nil {
		return nil, err
	}
	return NewTx(scoped), nil
}
