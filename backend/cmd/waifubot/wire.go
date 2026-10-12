package main

import (
	"github.com/karitham/waifubot/catalog"
	"github.com/karitham/waifubot/collection"
	"github.com/karitham/waifubot/collectionwire"
	"github.com/karitham/waifubot/storage"
)

// newCollectionStore wires adapters into a collection.DB.
func newCollectionStore(s storage.Store) collection.DB {
	return collectionwire.NewDB(s)
}

// newCatalogStore creates a catalog.Store from the underlying storage.
func newCatalogStore(s storage.Store) catalog.Store {
	return collectionwire.NewCatalog(s)
}
