package catalog

import (
	"context"
	"errors"
	"time"
)

var ErrNotFound = errors.New("not found")

type Character struct {
	ID         int64
	Name       string
	Image      string
	MediaTitle string
	Favorites  int
	UpdatedAt  time.Time // for cursor tracking
	IsActive   bool
}

// Drop is a Character that appeared in a channel drop.
type Drop = Character

// Store provides character catalog operations.
// Wraps collectionstore.Querier for character CRUD and guildstore.Querier for guild ownership.
type Store interface {
	UpsertCharacter(ctx context.Context, char Character) error
	GetCharacterByID(ctx context.Context, charID int64) (Character, error)
	SearchCharacters(ctx context.Context, userID uint64, term string) ([]Character, error)
	SearchGlobalCharacters(ctx context.Context, term string) ([]Character, error)
	GetCharacterHoldersInGuild(ctx context.Context, guildID uint64, charID int64) ([]uint64, error)
	GetActiveIDs(ctx context.Context) ([]int64, error)
	MarkCharactersInactive(ctx context.Context, ids []int64) error
}
