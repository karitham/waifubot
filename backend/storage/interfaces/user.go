package interfaces

import (
	"context"
	"time"
)

type User struct {
	ID              uint64
	UserID          uint64
	Quote           string
	Date            time.Time
	Favorite        int64
	Tokens          int
	AnilistURL      string
	DiscordUsername string
	DiscordAvatar   string
	LastUpdated     time.Time
}

type UserRepository interface {
	Create(ctx context.Context, userID uint64) error
	Get(ctx context.Context, userID uint64) (User, error)
	GetByAnilist(ctx context.Context, url string) (User, error)
	GetByDiscordUsername(ctx context.Context, username string) (User, error)
	List(ctx context.Context, limit, offset int) ([]User, error)
	UpdateTokens(ctx context.Context, userID uint64, tokens int) (User, error)
	UpdateQuote(ctx context.Context, userID uint64, quote string) error
	UpdateFavorite(ctx context.Context, userID uint64, favorite int64) error
	UpdateAnilistURL(ctx context.Context, userID uint64, url string) error
	UpdateDiscordInfo(ctx context.Context, userID uint64, username, avatar string) error
	CountFiltered(ctx context.Context, search string) (int64, error)
}
