// Package collectionwire composes the storage adapters into a collection.DB.
// Both the production binaries and the integration tests wire through here,
// so the transaction factory cannot silently diverge from what is tested.
package collectionwire

import (
	"context"

	"github.com/karitham/waifubot/catalog"
	"github.com/karitham/waifubot/collection"
	"github.com/karitham/waifubot/guild"
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
	return QueryTx(scoped), nil
}

// NewDB composes the storage adapters into a collection.DB.
func NewDB(s storage.Store) collection.DB {
	return db{Store: QueryStore(s), s: s}
}

// QueryStore composes the storage adapters into a read-only collection.Store.
func QueryStore(s storage.Store) collection.Store {
	return collection.NewStore(
		userpg.New(s.UserStore()),
		collectionpg.New(s.CollectionStore(), s.WishlistStore(), s.UserStore()),
		guildpg.New(s.GuildStore()),
		catalogpg.New(s.CollectionStore()),
	)
}

// QueryTx composes the storage adapters into a transaction-scoped
// collection.Tx. The storage transaction controls commit and rollback.
func QueryTx(s storage.Store) collection.Tx {
	user := userpg.New(s.UserStore())
	coll := collectionpg.New(s.CollectionStore(), s.WishlistStore(), s.UserStore())
	guildStore := guildpg.New(s.GuildStore())
	return collection.NewTx(
		s,
		collection.NewStore(user, coll, guildStore, catalogpg.New(s.CollectionStore())),
		user,
		coll,
		droppg.New(s.DropStore()),
		guildStore,
	)
}

// NewGuildDB composes the guild transaction surface over a collection.DB.
func NewGuildDB(db collection.DB) guild.DB {
	return guildDB{DB: db}
}

// guildDB adapts a collection.DB to the guild transaction surface. Go has no
// covariant returns, so Begin needs its own method to satisfy guild.DB.
type guildDB struct {
	collection.DB
}

// Begin starts a transaction and returns it as a guild.TxQuerier.
func (g guildDB) Begin(ctx context.Context) (guild.TxQuerier, error) {
	tx, err := g.DB.Begin(ctx)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

// NewCatalog composes the character catalog adapter.
func NewCatalog(s storage.Store) catalog.Store {
	return catalogpg.New(s.CollectionStore())
}
