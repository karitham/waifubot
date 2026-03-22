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

// newCollectionStore wires adapters into a collection.Store.
func newCollectionStore(s storage.Store) collection.Store {
	return newStoreFromStorage(s)
}

func newStoreFromStorage(s storage.Store) *collection.PostgresStore {
	catQ := s.CollectionStore()

	txFn := func(tx pgx.Tx) collection.Store {
		return collection.NewPostgresStore(
			userpg.New(s.UserStore()),
			collectionpg.New(catQ, s.WishlistStore()),
			droppg.New(s.DropStore()),
			guildpg.New(s.GuildStore()),
			catalogpg.New(catQ, s.GuildStore()),
			tx,
			nil,
		)
	}

	return collection.NewPostgresStore(
		userpg.New(s.UserStore()),
		collectionpg.New(catQ, s.WishlistStore()),
		droppg.New(s.DropStore()),
		guildpg.New(s.GuildStore()),
		catalogpg.New(catQ, s.GuildStore()),
		s.DB(),
		txFn,
	)
}

// newCatalogStore creates a catalog.Store from the underlying storage.
func newCatalogStore(s storage.Store) catalog.Store {
	return catalogpg.New(s.CollectionStore(), s.GuildStore())
}
