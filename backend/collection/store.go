package collection

import (
	"context"
	"errors"
	"time"

	"github.com/karitham/waifubot/catalog"
)

var (
	ErrNotFound                = errors.New("not found")
	ErrAlreadyOwned            = errors.New("character already in collection")
	ErrUserDoesNotOwnCharacter = errors.New("user does not own character")
)

type UserID = uint64

// Re-export catalog types so callers don't need a separate import.
type Character = catalog.Character
type Drop = catalog.Drop

// User holds user profile data.
type User struct {
	UserID          UserID
	Date            time.Time
	Tokens          int32
	Quote           string
	Favorite        int64
	AnilistURL      string
	DiscordUsername string
	DiscordAvatar   string
	LastUpdated     time.Time
}

type IndexingStatus int

const (
	IndexingPending IndexingStatus = iota
	IndexingInProgress
	IndexingCompleted
)

type GuildIndexStatus struct {
	Status    IndexingStatus
	UpdatedAt time.Time
}

type OwnedCharacter struct {
	Character
	Date   time.Time
	Source string
	UserID UserID
}

// UserRepository handles user CRUD and profile fields.
type UserRepository interface {
	GetUser(ctx context.Context, userID UserID) (User, error)
	CreateUser(ctx context.Context, userID UserID) error
	GetUserByAnilist(ctx context.Context, anilistURL string) (User, error)
	GetUserByDiscordUsername(ctx context.Context, username string) (User, error)
	UpdateLastRoll(ctx context.Context, userID UserID, when time.Time) error
	SpendTokens(ctx context.Context, userID UserID, amount int32) (User, error)
	AddTokens(ctx context.Context, userID UserID, amount int32) (User, error)
	UpdateFavorite(ctx context.Context, userID UserID, charID int64) error
	UpdateQuote(ctx context.Context, userID UserID, quote string) error
	UpdateAnilistURL(ctx context.Context, userID UserID, url string) error
	UpdateDiscordInfo(ctx context.Context, userID UserID, username, avatar string, lastUpdated time.Time) error
}

// CollectionRepository handles owned character operations.
type CollectionRepository interface {
	GetCollection(ctx context.Context, userID UserID) ([]OwnedCharacter, error)
	GetCollectionIDs(ctx context.Context, userID UserID) ([]int64, error)
	GetOwnedCharacter(ctx context.Context, userID UserID, charID int64) (OwnedCharacter, error)
	CharacterOwnedByUser(ctx context.Context, userID UserID, charID int64) (bool, error)
	AddToCollection(ctx context.Context, userID UserID, char Character, source string, acquiredAt time.Time) error
	RemoveFromCollection(ctx context.Context, userID UserID, charID int64) error
	GiveCharacter(ctx context.Context, from, to UserID, charID int64) (OwnedCharacter, error)
	CountCollection(ctx context.Context, userID UserID) (int64, error)
	// RemoveFromWishlist is called inside roll/claim/give transactions.
	RemoveFromWishlist(ctx context.Context, userID UserID, charID int64) error
}

// DropRepository handles channel drop operations for the claim flow.
type DropRepository interface {
	GetDropForUpdate(ctx context.Context, channelID uint64) (Drop, error)
	DeleteDrop(ctx context.Context, channelID uint64) error
}

// GuildQuerier provides guild indexing operations.
type GuildQuerier interface {
	IsGuildIndexed(ctx context.Context, guildID uint64) (GuildIndexStatus, error)
	StartIndexingJob(ctx context.Context, guildID uint64) error
	CompleteIndexingJob(ctx context.Context, guildID uint64) error
	UpsertGuildMembers(ctx context.Context, guildID uint64, memberIDs []uint64, indexedAt time.Time) error
	DeleteGuildMembersNotIn(ctx context.Context, guildID uint64, memberIDs []uint64) error
}

// Store is the transactional composition of all repositories.
// Each domain function (Roll, Claim, Give, Exchange) takes this.
type Store interface {
	UserRepository
	CollectionRepository
	DropRepository
	GuildQuerier
	catalog.Store

	WithTx(ctx context.Context) (Store, error)
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}
