package interfaces

import (
	"context"
	"time"
)

type Character struct {
	ID         int64
	Name       string
	Image      string
	MediaTitle string
}

type Collection struct {
	UserID      uint64
	CharacterID int64
	Source      string
	AcquiredAt  time.Time
}

type ListOptions struct {
	Limit     int
	Offset    int
	Search    string
	OrderBy   string
	Direction string
}

type CharacterRepository interface {
	Count(ctx context.Context, userID uint64) (int64, error)
	Get(ctx context.Context, userID uint64, charID int64) (Character, error)
	GetByID(ctx context.Context, charID int64) (Character, error)
	Give(ctx context.Context, userID uint64, charID int64, source string) (Collection, error)
	Insert(ctx context.Context, userID uint64, charID int64, source string) (Collection, error)
	Delete(ctx context.Context, userID uint64, charID int64) (Collection, error)
	List(ctx context.Context, userID uint64) ([]Collection, error)
	ListIDs(ctx context.Context, userID uint64) ([]int64, error)
	ListPaginated(ctx context.Context, userID uint64, opts ListOptions) ([]Collection, error)
	SearchCharacters(ctx context.Context, userID uint64, search string, limit int) ([]Character, error)
	SearchGlobalCharacters(ctx context.Context, search string, limit int) ([]Character, error)
	UpsertCharacter(ctx context.Context, char Character) (Character, error)
	UpdateImageName(ctx context.Context, charID int64, name, image string) (Character, error)
	UsersOwningCharFiltered(ctx context.Context, charID int64, userIDs []uint64) ([]uint64, error)
}
