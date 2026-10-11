package collection

import (
	"context"
	"errors"
	"time"

	"github.com/karitham/waifubot/catalog"
	"github.com/karitham/waifubot/guild"
)

var (
	ErrNotFound                = catalog.ErrNotFound
	ErrAlreadyOwned            = errors.New("character already in collection")
	ErrUserDoesNotOwnCharacter = errors.New("user does not own character")
)

// DefaultAnilistCharImage is the fallback image URL AniList returns when a
// character has no custom artwork. Characters with this image are excluded
// from drops (the image is embedded publicly in the drop message).
const DefaultAnilistCharImage = "https://s4.anilist.co/file/anilistcdn/character/large/default.jpg"

// DropWeightExponent controls how much favorites skew drop probability in the
// weighted-random formula. Higher values amplify the advantage of high-favorites
// (rarer) characters. 0 = uniform random, 1 = linear, 2 = quadratic.
const DropWeightExponent = 2.0

// RollWeightExponent controls the favorites bias for standard rolls (1.0 = linear).
const RollWeightExponent = 1.0

// UserID identifies a collection user.
type UserID uint64

// ChannelID identifies a channel that can hold a drop.
type ChannelID uint64

// User holds user profile data.
type User struct {
	UserID          UserID
	Date            time.Time
	Tokens          int32
	Quote           string
	Favorite        catalog.CharacterID
	AnilistURL      string
	DiscordUsername string
	DiscordAvatar   string
	LastUpdated     time.Time
}

type OwnedCharacter struct {
	catalog.Character
	Date   time.Time
	Source Source
	UserID UserID
}

// UserRepository handles user CRUD and profile fields.
type UserRepository interface {
	GetUser(ctx context.Context, userID UserID) (User, error)
	CreateUser(ctx context.Context, userID UserID) error
	GetUserByAnilist(ctx context.Context, anilistURL string) (User, error)
	GetUserByDiscordUsername(ctx context.Context, username string) (User, error)
	// RecordRoll atomically records a roll for the user, but only when the
	// cooldown has elapsed; it is the gate that serializes concurrent free
	// rolls. When the cooldown is still active it returns ErrRollCooldown
	// with the time the user may roll again.
	RecordRoll(ctx context.Context, userID UserID, now time.Time, cooldown time.Duration) error
	SpendTokens(ctx context.Context, userID UserID, amount int32) (User, error)
	AddTokens(ctx context.Context, userID UserID, amount int32) (User, error)
	UpdateFavorite(ctx context.Context, userID UserID, charID catalog.CharacterID) error
	UpdateQuote(ctx context.Context, userID UserID, quote string) error
	UpdateAnilistURL(ctx context.Context, userID UserID, url string) error
	UpdateDiscordInfo(ctx context.Context, userID UserID, username, avatar string, lastUpdated time.Time) error
}

// CollectionRepository handles owned character operations.
type CollectionRepository interface {
	GetCollection(ctx context.Context, userID UserID) ([]OwnedCharacter, error)
	GetCollectionIDs(ctx context.Context, userID UserID) ([]catalog.CharacterID, error)
	GetOwnedCharacter(ctx context.Context, userID UserID, charID catalog.CharacterID) (OwnedCharacter, error)
	AddToCollection(ctx context.Context, userID UserID, charID catalog.CharacterID, source Source, acquiredAt time.Time) error
	RemoveFromCollection(ctx context.Context, userID UserID, charID catalog.CharacterID) error
	GiveCharacter(ctx context.Context, from, to UserID, charID catalog.CharacterID) (OwnedCharacter, error)
	CountCollection(ctx context.Context, userID UserID) (int64, error)
	// SearchCharacters searches the user's collection by name or ID.
	SearchCharacters(ctx context.Context, userID UserID, term string) ([]catalog.Character, error)
	// RemoveFromWishlist is called inside roll/claim/give transactions.
	RemoveFromWishlist(ctx context.Context, userID UserID, charID catalog.CharacterID) error
	// RandomCharNotOwned returns a random active character not owned by the user,
	// weighted by favorites^weightExponent. Does NOT filter the default AniList image
	// (rolls/direct rolls use this path and the image isn't publicly embedded).
	RandomCharNotOwned(ctx context.Context, userID UserID, weightExponent float64) (catalog.Character, error)
	// RandomActiveChar returns a random active character for a channel drop,
	// weighted by favorites^weightExponent. Excludes characters with the default
	// AniList image (DefaultAnilistCharImage) since the image is embedded publicly.
	RandomActiveChar(ctx context.Context, weightExponent float64) (catalog.Character, error)
}

// DropRepository handles channel drop operations for the claim flow.
type DropRepository interface {
	GetDropForUpdate(ctx context.Context, channelID ChannelID) (catalog.Drop, error)
	DeleteDrop(ctx context.Context, channelID ChannelID) error
}

// Store is the transactional composition of all repositories.
// Each domain function (Roll, Claim, Give, Exchange) takes this.
type Store interface {
	UserRepository
	CollectionRepository
	DropRepository
	guild.Querier
	catalog.Store

	WithTx(ctx context.Context) (Store, error)
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}
