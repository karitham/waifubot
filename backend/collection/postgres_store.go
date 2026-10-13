package collection

import (
	"context"

	"github.com/karitham/waifubot/catalog"
	"github.com/karitham/waifubot/guild"
)

// guildReader embeds guild.Reader without colliding with catalog.Reader,
// which both promote under the name Reader in the composed surfaces.
type guildReader = guild.Reader

// composedStore is the read surface assembled from repository adapters.
type composedStore struct {
	UserReader
	CollectionReader
	guildReader
	catalog.Reader
}

// composedTx adds the transaction-scoped writers and control to the read
// surface.
type composedTx struct {
	composedStore
	UserWriter
	CollectionWriter
	DropWriter
	guild.Writer
	control TxControl
}

func (t composedTx) Commit(ctx context.Context) error   { return t.control.Commit(ctx) }
func (t composedTx) Rollback(ctx context.Context) error { return t.control.Rollback(ctx) }

// NewStore composes the reader adapters into a query Store.
func NewStore(
	user UserReader,
	coll CollectionReader,
	guild guild.Reader,
	cat catalog.Reader,
) Store {
	return composedStore{
		UserReader:       user,
		CollectionReader: coll,
		guildReader:      guild,
		Reader:           cat,
	}
}

// NewTx composes the reader and writer adapters over one transaction into a
// Tx. The control commits and rolls back that transaction.
func NewTx(
	control TxControl,
	store Store,
	user UserWriter,
	coll CollectionWriter,
	drop DropWriter,
	guild guild.Writer,
) Tx {
	return composedTx{
		composedStore: composedStore{
			UserReader:       store,
			CollectionReader: store,
			guildReader:      store,
			Reader:           store,
		},
		UserWriter:       user,
		CollectionWriter: coll,
		DropWriter:       drop,
		Writer:           guild,
		control:          control,
	}
}
