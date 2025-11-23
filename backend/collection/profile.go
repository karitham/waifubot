package collection

//go:generate mockgen -source=profile.go -destination=profile_mock.go -package=collection -mock_names=Store=MockProfileStore,AnimeService=MockAnimeService
//go:generate mockgen -source=../storage/collectionstore/querier.go -destination=collectionstore_mock.go -package=collection -mock_names=Querier=MockCollectionQuerier
//go:generate mockgen -source=../storage/userstore/querier.go -destination=userstore_mock.go -package=collection -mock_names=Querier=MockUserQuerier
//go:generate mockgen -source=../storage/wishliststore/querier.go -destination=wishliststore_mock.go -package=collection -mock_names=Querier=MockWishlistQuerier

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/Karitham/corde"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/karitham/waifubot/storage"
	"github.com/karitham/waifubot/storage/collectionstore"
	"github.com/karitham/waifubot/storage/guildstore"
	"github.com/karitham/waifubot/storage/userstore"
	"github.com/karitham/waifubot/storage/wishliststore"
)

// Store defines the interface for database operations
type Store interface {
	Tx(ctx context.Context) (storage.Store, error)
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
	CollectionStore() collectionstore.Querier
	UserStore() userstore.Querier
	GuildStore() guildstore.Querier
	WishlistStore() wishliststore.Querier
}

// AnimeService defines the interface for anime operations
type AnimeService interface {
	RandomChar(ctx context.Context, notIn ...int64) (MediaCharacter, error)
	Anime(ctx context.Context, name string) ([]Media, error)
	Manga(ctx context.Context, name string) ([]Media, error)
	User(ctx context.Context, name string) ([]TrackerUser, error)
	Character(ctx context.Context, name string) ([]MediaCharacter, error)
}

// Config holds configuration values
type Config struct {
	RollCooldown time.Duration
	TokensNeeded int32
}

// MediaCharacter represents a character from the anime service
type MediaCharacter struct {
	ID          int64
	Name        string
	ImageURL    string
	URL         string
	Description string
	MediaTitle  string
}

// Character represents a character in a user's collection
type Character struct {
	Date   time.Time       `json:"date"`
	Image  string          `json:"image"`
	Name   string          `json:"name"`
	Type   string          `json:"type"`
	UserID corde.Snowflake `json:"user_id"`
	ID     int64           `json:"id"`
}

// User represents a user profile
type User struct {
	Date       time.Time       `json:"date"`
	Quote      string          `json:"quote"`
	Favorite   uint64          `json:"favorite"`
	UserID     corde.Snowflake `json:"user_id"`
	AnilistURL string          `json:"anilist_url,omitempty"`
	Tokens     int32           `json:"tokens"`
}

// Profile represents a complete user profile with character count and favorite
type Profile struct {
	User
	CharacterCount int
	Favorite       Character
}

// Media represents an anime or manga
type Media struct {
	Title           string
	URL             string
	CoverImageURL   string
	BannerImageURL  string
	CoverImageColor uint32
	Description     string
}

// TrackerUser represents an anime tracker user
type TrackerUser struct {
	Name     string
	URL      string
	ImageURL string
	About    string
}

// UserProfile retrieves a user's profile
func UserProfile(ctx context.Context, store Store, userID corde.Snowflake) (Profile, error) {
	// Get user
	u, err := store.UserStore().Get(ctx, uint64(userID))
	if err != nil {
		return Profile{}, err
	}

	// Get favorite if any
	var favorite Character
	if u.Favorite.Valid {
		favRow, err := store.CollectionStore().Get(ctx, collectionstore.GetParams{ID: u.Favorite.Int64, UserID: uint64(userID)})
		if err != nil {
			// Error getting favorite, continue without it
		} else {
			favorite = Character{
				Date:   favRow.Date.Time,
				Image:  favRow.Image,
				Name:   favRow.Name,
				Type:   favRow.Source,
				UserID: userID,
				ID:     int64(favRow.ID),
			}
		}
	}

	// Count chars
	count, err := store.CollectionStore().Count(ctx, uint64(userID))
	if err != nil {
		return Profile{}, err
	}

	return Profile{
		User: User{
			Date:       u.Date.Time,
			Quote:      u.Quote,
			UserID:     corde.Snowflake(u.UserID),
			Favorite:   uint64(u.Favorite.Int64),
			Tokens:     u.Tokens,
			AnilistURL: u.AnilistUrl,
		},
		Favorite:       favorite,
		CharacterCount: int(count),
	}, nil
}

// SetFavorite sets a user's favorite character
func SetFavorite(ctx context.Context, store Store, userID corde.Snowflake, charID int64) error {
	return store.UserStore().UpdateFavorite(ctx, userstore.UpdateFavoriteParams{
		Favorite: pgtype.Int8{Int64: charID, Valid: true},
		UserID:   uint64(userID),
	})
}

// SetAnilistURL sets a user's anilist URL
func SetAnilistURL(ctx context.Context, store Store, userID corde.Snowflake, anilistURL string) error {
	parsedURL, err := url.Parse(anilistURL)
	if err != nil {
		return fmt.Errorf("invalid URL")
	}

	if parsedURL.Host != "anilist.co" {
		return fmt.Errorf("invalid Anilist URL")
	}

	if !strings.HasPrefix(parsedURL.Path, "/user/") {
		return fmt.Errorf("invalid Anilist URL")
	}

	return store.UserStore().UpdateAnilistURL(ctx, userstore.UpdateAnilistURLParams{
		AnilistUrl: anilistURL,
		UserID:     uint64(userID),
	})
}

// SetQuote sets a user's quote
func SetQuote(ctx context.Context, store Store, userID corde.Snowflake, quote string) error {
	if len(quote) > 1024 {
		return fmt.Errorf("quote is too long")
	}

	return store.UserStore().UpdateQuote(ctx, userstore.UpdateQuoteParams{
		Quote:  quote,
		UserID: uint64(userID),
	})
}

// SearchCharacters searches for characters in a user's collection for autocomplete
func SearchCharacters(ctx context.Context, store Store, userID corde.Snowflake, term string) ([]collectionstore.Character, error) {
	rows, err := store.CollectionStore().SearchCharacters(ctx, collectionstore.SearchCharactersParams{
		UserID: uint64(userID),
		Term:   term,
		Lim:    25,
		Off:    0,
	})
	if err != nil {
		return nil, err
	}
	if len(rows) > 25 {
		rows = rows[:25]
	}

	// Convert SearchCharactersRow to Character
	chars := make([]collectionstore.Character, len(rows))
	for i, row := range rows {
		chars[i] = collectionstore.Character{
			ID:    row.ID,
			Name:  row.Name,
			Image: row.Image,
		}
	}

	return chars, nil
}
