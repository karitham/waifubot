package guild

import (
	"context"

	"github.com/karitham/waifubot/catalog"
	"github.com/karitham/waifubot/collection"
)

// GuildQuerier is an alias for the interface defined in collection.
// Re-exported so guild consumers don't need to import collection directly.
type GuildQuerier = collection.GuildQuerier

// CharacterQuerier provides character lookups with guild ownership.
type CharacterQuerier interface {
	catalog.Store
	GetCharacterHoldersInGuild(ctx context.Context, guildID uint64, charID int64) ([]uint64, error)
}

// TxQuerier is a GuildQuerier inside a transaction.
type TxQuerier interface {
	GuildQuerier
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}
