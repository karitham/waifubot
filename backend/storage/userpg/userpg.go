package userpg

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/karitham/waifubot/collection"
	"github.com/karitham/waifubot/storage/userstore"
)

type Pg struct {
	Q userstore.Querier
}

func New(q userstore.Querier) *Pg {
	return &Pg{Q: q}
}

func (p *Pg) GetUser(ctx context.Context, userID collection.UserID) (collection.User, error) {
	u, err := p.Q.Get(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return collection.User{}, collection.ErrNotFound
		}
		return collection.User{}, err
	}
	return toUser(u), nil
}

func (p *Pg) CreateUser(ctx context.Context, userID collection.UserID) error {
	return p.Q.Create(ctx, userID)
}

func (p *Pg) GetUserByAnilist(ctx context.Context, anilistURL string) (collection.User, error) {
	u, err := p.Q.GetByAnilist(ctx, anilistURL)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return collection.User{}, collection.ErrNotFound
		}
		return collection.User{}, err
	}
	return toUser(u), nil
}

func (p *Pg) GetUserByDiscordUsername(ctx context.Context, username string) (collection.User, error) {
	u, err := p.Q.GetByDiscordUsername(ctx, username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return collection.User{}, collection.ErrNotFound
		}
		return collection.User{}, err
	}
	return toUser(u), nil
}

func (p *Pg) UpdateLastRoll(ctx context.Context, userID collection.UserID, when time.Time) error {
	return p.Q.UpdateDate(ctx, userstore.UpdateDateParams{
		Date:   pgtype.Timestamp{Time: when.UTC(), Valid: true},
		UserID: userID,
	})
}

func (p *Pg) SpendTokens(ctx context.Context, userID collection.UserID, amount int32) (collection.User, error) {
	u, err := p.Q.SpendTokens(ctx, userstore.SpendTokensParams{
		Tokens: amount,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return collection.User{}, collection.ErrInsufficientTokens
		}
		return collection.User{}, err
	}
	return toUser(u), nil
}

func (p *Pg) AddTokens(ctx context.Context, userID collection.UserID, amount int32) (collection.User, error) {
	u, err := p.Q.UpdateTokens(ctx, userstore.UpdateTokensParams{
		Tokens: amount,
		UserID: userID,
	})
	if err != nil {
		return collection.User{}, err
	}
	return toUser(u), nil
}

func (p *Pg) UpdateFavorite(ctx context.Context, userID collection.UserID, charID int64) error {
	return p.Q.UpdateFavorite(ctx, userstore.UpdateFavoriteParams{
		Favorite: pgtype.Int8{Int64: charID, Valid: true},
		UserID:   userID,
	})
}

func (p *Pg) UpdateQuote(ctx context.Context, userID collection.UserID, quote string) error {
	return p.Q.UpdateQuote(ctx, userstore.UpdateQuoteParams{
		Quote:  quote,
		UserID: userID,
	})
}

func (p *Pg) UpdateAnilistURL(ctx context.Context, userID collection.UserID, url string) error {
	return p.Q.UpdateAnilistURL(ctx, userstore.UpdateAnilistURLParams{
		AnilistUrl: url,
		UserID:     userID,
	})
}

func (p *Pg) UpdateDiscordInfo(ctx context.Context, userID collection.UserID, username, avatar string, lastUpdated time.Time) error {
	return p.Q.UpdateDiscordInfo(ctx, userstore.UpdateDiscordInfoParams{
		DiscordUsername: username,
		DiscordAvatar:   avatar,
		LastUpdated:     pgtype.Timestamp{Time: lastUpdated.UTC(), Valid: true},
		UserID:          userID,
	})
}

func toUser(u userstore.User) collection.User {
	return collection.User{
		UserID:          u.UserID,
		Date:            u.Date.Time,
		Tokens:          u.Tokens,
		Quote:           u.Quote,
		Favorite:        u.Favorite.Int64,
		AnilistURL:      u.AnilistUrl,
		DiscordUsername: u.DiscordUsername,
		DiscordAvatar:   u.DiscordAvatar,
		LastUpdated:     u.LastUpdated.Time,
	}
}
