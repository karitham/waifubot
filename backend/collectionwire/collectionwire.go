// Package collectionwire composes the storage adapters into a collection.DB.
// Both the production binaries and the integration tests wire through here,
// so the transaction factory cannot silently diverge from what is tested.
package collectionwire

import (
	"context"

	"github.com/karitham/waifubot/catalog"
	"github.com/karitham/waifubot/collection"
	"github.com/karitham/waifubot/storage"
	"github.com/karitham/waifubot/storage/catalogpg"
	"github.com/karitham/waifubot/storage/collectionpg"
	"github.com/karitham/waifubot/storage/droppg"
	"github.com/karitham/waifubot/storage/guildpg"
	"github.com/karitham/waifubot/storage/userpg"
)

// db implements collection.DB over a storage.Store.
type db struct {
	collection.Store
	s storage.Store
}

// Begin starts a transaction and returns a Tx whose adapters and committer
// both come from the same transaction.
func (d db) Begin(ctx context.Context) (collection.Tx, error) {
	scoped, err := d.s.Tx(ctx)
	if err != nil {
		return nil, err
	}
	return tx{Store: QueryStore(scoped), control: scoped}, nil
}

// tx pairs tx-scoped adapters with the transaction that scopes them.
type tx struct {
	collection.Store
	control storage.Store
}

// Commit commits the transaction.
func (t tx) Commit(ctx context.Context) error {
	return t.control.Commit(ctx)
}

// Rollback rolls back the transaction.
func (t tx) Rollback(ctx context.Context) error {
	return t.control.Rollback(ctx)
}

// NewDB composes the storage adapters into a collection.DB.
func NewDB(s storage.Store) collection.DB {
	return db{Store: QueryStore(s), s: s}
}

// QueryStore composes the storage adapters into a query-only collection.Store.
func QueryStore(s storage.Store) collection.Store {
	return collection.NewStore(
		userpg.New(s.UserStore()),
		collectionpg.New(s.CollectionStore(), s.WishlistStore(), s.UserStore()),
		droppg.New(s.DropStore()),
		guildpg.New(s.GuildStore()),
		catalogpg.New(s.CollectionStore()),
	)
}

// NewCatalog composes the character catalog adapter.
func NewCatalog(s storage.Store) catalog.Store {
	return catalogpg.New(s.CollectionStore())
}
