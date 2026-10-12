package collection

import (
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
