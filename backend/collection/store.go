package collection

import (
	"context"
	"errors"
	"time"

	"github.com/karitham/waifubot/catalog"
	"github.com/karitham/waifubot/drops"
	"github.com/karitham/waifubot/guild"
)

var (
	ErrNotFound                = catalog.ErrNotFound
	ErrAlreadyOwned            = errors.New("character already in collection")
	ErrUserDoesNotOwnCharacter = errors.New("user does not own character")
)

// RollWeightExponent controls the favorites bias for standard rolls (1.0 = linear).
const RollWeightExponent = 1.0

// UserID identifies a collection user.
type UserID uint64

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

// UserReader reads user profile data.
type UserReader interface {
	GetUser(ctx context.Context, userID UserID) (User, error)
	GetUserByAnilist(ctx context.Context, anilistURL string) (User, error)
	GetUserByDiscordUsername(ctx context.Context, username string) (User, error)
}

// UserWriter holds the user writes that run inside withUsers, with the user
// materialized and locked.
type UserWriter interface {
	// CreateUser creates the user when it is absent. It is idempotent and
	// the only path that creates users.
	CreateUser(ctx context.Context, userID UserID) error
	// LockUser locks the user's row for the rest of the transaction,
	// returning ErrNotFound when the user does not exist.
	LockUser(ctx context.Context, userID UserID) error
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

// CollectionReader reads owned characters and wishlist membership.
type CollectionReader interface {
	GetCollection(ctx context.Context, userID UserID) ([]OwnedCharacter, error)
	GetCollectionIDs(ctx context.Context, userID UserID) ([]catalog.CharacterID, error)
	GetOwnedCharacter(ctx context.Context, userID UserID, charID catalog.CharacterID) (OwnedCharacter, error)
	CountCollection(ctx context.Context, userID UserID) (int64, error)
	// SearchCharacters searches the user's collection by name or ID.
	SearchCharacters(ctx context.Context, userID UserID, term string) ([]catalog.Character, error)
	// RandomCharNotOwned returns a random active character not owned by the user,
	// weighted by favorites^weightExponent. Does NOT filter the default AniList image
	// (rolls/direct rolls use this path and the image isn't publicly embedded).
	RandomCharNotOwned(ctx context.Context, userID UserID, weightExponent float64) (catalog.Character, error)
}

// WishlistReader reads wishlist membership and discovery views.
type WishlistReader interface {
	GetUserCharacterWishlist(ctx context.Context, userID UserID) ([]WishlistEntry, error)
	// Discovery
	GetWishlistHolders(ctx context.Context, characterIDs []catalog.CharacterID, userID UserID, guildID guild.ID) ([]WishlistUserSet, error)
	GetWantedCharacters(ctx context.Context, userID UserID, guildID guild.ID) ([]WishlistUserSet, error)
	CompareWithUser(ctx context.Context, userID1, userID2 UserID) (WishlistComparison, error)
	GetUsersWantingCharacter(ctx context.Context, charID catalog.CharacterID, guildID guild.ID, excludeUserID UserID) ([]guild.MemberID, error)
}

// CollectionWriter holds the collection and wishlist writes; all run inside
// a transaction.
type CollectionWriter interface {
	AddToCollection(ctx context.Context, userID UserID, charID catalog.CharacterID, source Source, acquiredAt time.Time) error
	RemoveFromCollection(ctx context.Context, userID UserID, charID catalog.CharacterID) error
	GiveCharacter(ctx context.Context, from, to UserID, charID catalog.CharacterID) (OwnedCharacter, error)
	// SetFavorite atomically sets the user's favorite character, requiring
	// ownership when setting; a zero charID clears it.
	SetFavorite(ctx context.Context, userID UserID, charID catalog.CharacterID) error
	// RemoveFromWishlist removes a character from the user's wishlist.
	RemoveFromWishlist(ctx context.Context, userID UserID, charID catalog.CharacterID) error
	// RemoveAllFromWishlist clears the user's wishlist.
	RemoveAllFromWishlist(ctx context.Context, userID UserID) error
	// AddCharactersToWishlist is called inside wishlist-add transactions.
	// It reports how many characters were actually added.
	AddCharactersToWishlist(ctx context.Context, userID UserID, characterIDs []catalog.CharacterID) (int64, error)
}

// Store is the non-transactional surface: reads over users, collections,
// wishlists, guilds, and the catalog.
type Store interface {
	UserReader
	CollectionReader
	WishlistReader
	guild.Reader
	catalog.Reader
}

// Tx is a Store scoped to a database transaction; it adds the writes that
// must run inside one, plus transaction control.
type Tx interface {
	Store
	UserWriter
	CollectionWriter
	drops.TxWriter
	guild.Writer
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

// DB is the pool-level entry point: the query surface plus transaction
// creation.
type DB interface {
	Store
	Begin(ctx context.Context) (Tx, error)
}
