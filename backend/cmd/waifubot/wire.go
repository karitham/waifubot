package main

import (
	"github.com/karitham/waifubot/catalog"
	"github.com/karitham/waifubot/collection"
	"github.com/karitham/waifubot/collectionwire"
	"github.com/karitham/waifubot/guild"
	"github.com/karitham/waifubot/storage"
)

// newCollectionStore wires adapters into a collection.DB.
func newCollectionStore(s storage.Store) collection.DB {
	return collectionwire.NewDB(s)
}

// newGuildDB adapts the collection transaction surface for the guild indexer.
func newGuildDB(db collection.DB) guild.DB {
	return collectionwire.NewGuildDB(db)
}

// newCatalogStore creates a catalog.Store from the underlying storage.
func newCatalogStore(s storage.Store) catalog.Store {
	return collectionwire.NewCatalog(s)
}
