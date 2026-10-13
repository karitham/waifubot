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
	UserID              UserID
	LastRoll            time.Time
	Tokens              int32
	Quote               string
	FavoriteCharacterID catalog.CharacterID
	AnilistURL          string
	DiscordUsername     string
	DiscordAvatar       string
	LastUpdated         time.Time
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
	// CreateUser creates the user when it is absent. It is idempotent and
	// the only path that creates users.
	CreateUser(ctx context.Context, userID UserID) error
	// LockUser locks the user's row for the rest of the transaction,
	// returning ErrNotFound when the user does not exist.
	LockUser(ctx context.Context, userID UserID) error
	GetUserByAnilist(ctx context.Context, anilistURL string) (User, error)
	GetUserByDiscordUsername(ctx context.Context, username string) (User, error)
	// RecordRoll atomically records a roll for the user, but only when the
	// cooldown has elapsed; it is the gate that serializes concurrent free
	// rolls. When the cooldown is still active it returns ErrRollCooldown
	// with the time the user may roll again.
	RecordRoll(ctx context.Context, userID UserID, now time.Time, cooldown time.Duration) error
	// SpendTokens deducts a positive amount from the user's balance.
	SpendTokens(ctx context.Context, userID UserID, amount TokenAmount) (User, error)
	// AddTokens credits a positive amount to the user's balance.
	AddTokens(ctx context.Context, userID UserID, amount TokenAmount) (User, error)
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
	// SetFavorite atomically sets the user's favorite character, requiring
	// ownership when setting; a zero charID clears it.
	SetFavorite(ctx context.Context, userID UserID, charID catalog.CharacterID) error
	// SearchCharacters searches the user's collection by name or ID.
	SearchCharacters(ctx context.Context, userID UserID, term string) ([]catalog.Character, error)
	// RemoveFromWishlist removes a character from the user's wishlist.
	RemoveFromWishlist(ctx context.Context, userID UserID, charID catalog.CharacterID) error
	// AddCharactersToWishlist is called inside wishlist-add transactions.
	// It reports how many characters were actually added.
	AddCharactersToWishlist(ctx context.Context, userID UserID, characterIDs []catalog.CharacterID) (int64, error)
	// RandomCharNotOwned returns a random active character not owned by the user,
	// weighted by favorites^weightExponent. Does NOT filter the default AniList image
	// (rolls/direct rolls use this path and the image isn't publicly embedded).
	RandomCharNotOwned(ctx context.Context, userID UserID, weightExponent float64) (catalog.Character, error)
	// RandomActiveChar returns a random active character for a channel drop,
	// weighted by favorites^weightExponent. Excludes characters with the default
	// AniList image (DefaultAnilistCharImage) since the image is embedded publicly.
	RandomActiveChar(ctx context.Context, weightExponent float64) (catalog.Character, error)
}

// DropRepository handles the channel drop lifecycle.
type DropRepository interface {
	// SetDrop stores the character as the channel's drop. The character must
	// already be in the catalog; the FK enforces it.
	SetDrop(ctx context.Context, channelID ChannelID, char catalog.Character) error
	GetDropForUpdate(ctx context.Context, channelID ChannelID) (catalog.Character, error)
	DeleteDrop(ctx context.Context, channelID ChannelID) error
}

// Store is the query surface over the domain repositories.
type Store interface {
	UserRepository
	CollectionRepository
	DropRepository
	guild.Querier
	catalog.Reader
}

// Tx is a Store scoped to a database transaction.
type Tx interface {
	Store
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

// DB is the pool-level entry point: the query surface plus transaction
// creation.
type DB interface {
	Store
	Begin(ctx context.Context) (Tx, error)
}
