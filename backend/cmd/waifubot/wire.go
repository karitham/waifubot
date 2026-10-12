package main

import (
	"github.com/jackc/pgx/v5"

	"github.com/karitham/waifubot/catalog"
	"github.com/karitham/waifubot/collection"
	"github.com/karitham/waifubot/storage"
	"github.com/karitham/waifubot/storage/catalogpg"
	"github.com/karitham/waifubot/storage/collectionpg"
	"github.com/karitham/waifubot/storage/droppg"
	"github.com/karitham/waifubot/storage/guildpg"
	"github.com/karitham/waifubot/storage/userpg"
)

// newCollectionStore wires adapters into a collection.DB.
func newCollectionStore(s storage.Store) collection.DB {
	return newStoreFromStorage(s)
}

func newStoreFromStorage(s storage.Store) *collection.PostgresStore {
	// txFn must build every adapter over the transaction; using the
	// pool-level adapters here would make the "transaction" a no-op.
	txFn := func(tx pgx.Tx) collection.Store {
		scoped := s.WithTx(tx)
		return collection.NewStore(
			userpg.New(scoped.UserStore()),
			collectionpg.New(scoped.CollectionStore(), scoped.WishlistStore()),
			droppg.New(scoped.DropStore()),
			guildpg.New(scoped.GuildStore()),
			catalogpg.New(scoped.CollectionStore()),
		)
	}

	return collection.NewPostgresStore(
		collection.NewStore(
			userpg.New(s.UserStore()),
			collectionpg.New(s.CollectionStore(), s.WishlistStore()),
			droppg.New(s.DropStore()),
			guildpg.New(s.GuildStore()),
			catalogpg.New(s.CollectionStore()),
		),
		s.DB(),
		txFn,
	)
}

// newCatalogStore creates a catalog.Store from the underlying storage.
func newCatalogStore(s storage.Store) catalog.Store {
	return catalogpg.New(s.CollectionStore())
}
