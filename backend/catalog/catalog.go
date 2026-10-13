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

// Reader reads the character catalog.
type Reader interface {
	GetCharacterByID(ctx context.Context, charID CharacterID) (Character, error)
	SearchGlobalCharacters(ctx context.Context, term string) ([]Character, error)
	GetActiveIDs(ctx context.Context) ([]CharacterID, error)
	// KnownCharacterIDs returns the subset of the given IDs present in the
	// catalog. It does not filter on is_active: the flag reflects the
	// hydrator's last pass and can be stale for characters a live tracker
	// fetch just returned.
	KnownCharacterIDs(ctx context.Context, ids []CharacterID) ([]CharacterID, error)
}

// Writer hydrates the character catalog; the sync service and explicit
// maintenance tools are its only writers.
type Writer interface {
	UpsertCharacter(ctx context.Context, char Character) error
	MarkCharactersInactive(ctx context.Context, charIDs []CharacterID) error
}

// Store is the full catalog surface.
type Store interface {
	Reader
	Writer
}
