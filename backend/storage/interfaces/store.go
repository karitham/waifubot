package interfaces

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Store interface {
	UserStore() UserRepository
	CharacterStore() CharacterRepository
	WishlistStore() WishlistRepository
	GuildStore() GuildRepository
	CommandStore() CommandRepository
	DropStore() DropRepository
	InteractionStore() InteractionRepository
	Tx(ctx context.Context) (Store, error)
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

type TXer interface {
	Begin(ctx context.Context) (pgxpool.Pool, error)
}
