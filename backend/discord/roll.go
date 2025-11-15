package discord

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/Karitham/corde"
)

type RandomCharer interface {
	RandomChar(ctx context.Context, notIn ...int64) (MediaCharacter, error)
}

type MediaCharacter struct {
	ID          int64
	Name        string
	ImageURL    string
	URL         string
	Description string
	MediaTitle  string
}

type Character struct {
	Date   time.Time       `json:"date"`
	Image  string          `json:"image"`
	Name   string          `json:"name"`
	Type   string          `json:"type"`
	UserID corde.Snowflake `json:"user_id"`
	ID     int64           `json:"id"`
}

type User struct {
	Date       time.Time       `json:"date"`
	Quote      string          `json:"quote"`
	Favorite   uint64          `json:"favorite"`
	UserID     corde.Snowflake `json:"user_id"`
	AnilistURL string          `json:"anilist_url,omitempty"`
	Tokens     int32           `json:"tokens"`
}

func (b *Bot) roll(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[corde.SlashCommandInteractionData]) {
	var char MediaCharacter

	if err := b.Store.Tx(ctx, func(s Store) error {
		user, err := s.User(ctx, i.Member.User.ID)
		if err != nil {
			return err
		}

		var updateUser func() error
		switch {
		case time.Now().After(user.Date.Add(b.RollCooldown)):
			updateUser = func() error {
				if err = s.SetUserDate(ctx, i.Member.User.ID, time.Now()); err != nil {
					slog.ErrorContext(ctx, "error with db service", "error", err)
					w.Respond(rspErr("An error occurred dialing the database, please try again later"))
					return err
				}
				return nil
			}
		case user.Tokens >= b.TokensNeeded:
			updateUser = func() error {
				if _, err = s.ConsumeDropTokens(ctx, i.Member.User.ID, b.TokensNeeded); err != nil {
					slog.ErrorContext(ctx, "error with db service", "error", err)
					w.Respond(rspErr("An error occurred dialing the database, please try again later"))
					return err
				}
				return nil
			}
		default:
			w.Respond(newErrf("You need %d tokens to roll, you have %d, or you can wait %s until next free roll.",
				b.TokensNeeded,
				user.Tokens,
				time.Until(user.Date.Add(b.RollCooldown)).Round(time.Second),
			))
			return errors.New("not enough tokens or time")
		}

		charsIDs, err := s.CharsIDs(ctx, i.Member.User.ID)
		if err != nil {
			slog.ErrorContext(ctx, "error with db service", "error", err)
			w.Respond(rspErr("An error occurred dialing the database, please try again later"))
			return err
		}

		c, err := b.AnimeService.RandomChar(ctx, charsIDs...)
		if err != nil {
			slog.ErrorContext(ctx, "error with anime service", "error", err)
			w.Respond(rspErr("An error getting a random character occurred, please try again later"))
			return err
		}
		char = c

		if err := s.PutChar(
			ctx,
			i.Member.User.ID,
			Character{
				Date:   time.Now(),
				Image:  c.ImageURL,
				Name:   c.Name,
				Type:   "ROLL",
				UserID: i.Member.User.ID,
				ID:     int64(c.ID),
			}); err != nil {
			slog.ErrorContext(ctx, "error with db service", "error", err)
			w.Respond(rspErr("An error occurred dialing the database, please try again later"))
			return err
		}

		return updateUser()
	}); err != nil {
		return
	}

	w.Respond(corde.NewEmbed().
		Title(char.Name).
		URL(char.URL).
		Footer(corde.Footer{IconURL: AnilistIconURL, Text: "View on Anilist"}).
		Thumbnail(corde.Image{URL: char.ImageURL}).
		Descriptionf("You rolled %s!\nID: %d\nFrom: %s", char.Name, char.ID, char.MediaTitle),
	)
}
