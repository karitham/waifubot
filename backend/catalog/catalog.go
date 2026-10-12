package catalog

import (
	"context"
	"errors"
	"time"
)

var ErrNotFound = errors.New("not found")

// CharacterID identifies a character in the local catalog.
type CharacterID int64

type Character struct {
	ID         CharacterID
	Name       string
	Image      string
	MediaTitle string
	Popularity int
	UpdatedAt  time.Time // for cursor tracking
	IsActive   bool
}

// Store provides character catalog operations.
type Store interface {
	UpsertCharacter(ctx context.Context, char Character) error
	GetCharacterByID(ctx context.Context, charID CharacterID) (Character, error)
	SearchGlobalCharacters(ctx context.Context, term string) ([]Character, error)
	GetActiveIDs(ctx context.Context) ([]CharacterID, error)
	MarkCharactersInactive(ctx context.Context, charIDs []CharacterID) error
}
