package inmemory

import (
	"context"

	"github.com/karitham/waifubot/storage/interfaces"
)

type Store struct {
	users        *UserStore
	characters   *CharacterStore
	wishlists    *WishlistStore
	guilds       *GuildStore
	commands     *CommandStore
	drops        *DropStore
	interactions *InteractionStore
}

func NewStore() *Store {
	return &Store{
		users:        NewUserStore(),
		characters:   NewCharacterStore(),
		wishlists:    NewWishlistStore(),
		guilds:       NewGuildStore(),
		commands:     NewCommandStore(),
		drops:        NewDropStore(),
		interactions: NewInteractionStore(),
	}
}

func (s *Store) UserStore() interfaces.UserRepository {
	return s.users
}

func (s *Store) CharacterStore() interfaces.CharacterRepository {
	return s.characters
}

func (s *Store) WishlistStore() interfaces.WishlistRepository {
	return s.wishlists
}

func (s *Store) GuildStore() interfaces.GuildRepository {
	return s.guilds
}

func (s *Store) CommandStore() interfaces.CommandRepository {
	return s.commands
}

func (s *Store) DropStore() interfaces.DropRepository {
	return s.drops
}

func (s *Store) InteractionStore() interfaces.InteractionRepository {
	return s.interactions
}

func (s *Store) Tx(ctx context.Context) (interfaces.Store, error) {
	return s, nil
}

func (s *Store) Commit(ctx context.Context) error {
	return nil
}

func (s *Store) Rollback(ctx context.Context) error {
	return nil
}
