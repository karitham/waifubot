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

// NewDB composes the storage adapters into a collection.DB. The begin
// function derives every adapter from the new transaction; the integration
// suite asserts a write inside it is rolled back.
func NewDB(s storage.Store) collection.DB {
	return collection.NewPostgresStore(QueryStore(s), func(ctx context.Context) (collection.Tx, error) {
		scoped, err := s.Tx(ctx)
		if err != nil {
			return nil, err
		}
		return collection.NewTx(QueryStore(scoped), scoped), nil
	})
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
