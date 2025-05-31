package discord

import (
	"context"
	"time"

	"github.com/Karitham/corde"
	"github.com/rs/zerolog/log"
)

const (
	AnilistColor   = 0x02a9ff
	AnilistIconURL = "https://anilist.co/img/icons/favicon-32x32.png"
)

// Store is the database
type Store interface {
	PutChar(context.Context, corde.Snowflake, Character) error
	Chars(context.Context, corde.Snowflake) ([]Character, error)
	VerifyChar(context.Context, corde.Snowflake, int64) (Character, error)
	CharsIDs(ctx context.Context, userID corde.Snowflake) ([]int64, error)
	DeleteChar(context.Context, corde.Snowflake, int64) (Character, error)
	CharsStartingWith(context.Context, corde.Snowflake, string) ([]Character, error)
	User(context.Context, corde.Snowflake) (User, error)
	ProfileOverview(context.Context, corde.Snowflake) (Profile, error)
	SetUserDate(context.Context, corde.Snowflake, time.Time) error
	SetUserFavorite(context.Context, corde.Snowflake, int64) error
	SetUserQuote(context.Context, corde.Snowflake, string) error
	SetUserAnilistURL(context.Context, corde.Snowflake, string) error
	GiveUserChar(ctx context.Context, dst corde.Snowflake, src corde.Snowflake, charID int64) error
	AddDropToken(context.Context, corde.Snowflake) error
	ConsumeDropTokens(context.Context, corde.Snowflake, int32) (User, error)
	Tx(ctx context.Context, fn func(s Store) error) error
}

// Interacter
type Interacter interface {
	GetInteractionCount(ctx context.Context, channelID corde.Snowflake) (int64, error)
	ResetInteractionCount(ctx context.Context, channelID corde.Snowflake) error
	IncrementInteractionCount(ctx context.Context, channelID corde.Snowflake) error

	SetChannelChar(ctx context.Context, channelID corde.Snowflake, char MediaCharacter) error
	GetChannelChar(ctx context.Context, channelID corde.Snowflake) (MediaCharacter, error)
	RemoveChannelChar(ctx context.Context, channelID corde.Snowflake) error
}

// TrackingService is the interface for the anilist service
type TrackingService interface {
	RandomCharer
	AnimeSearcher
	CharSearcher
	MangaSearcher
	UserSearcher
}

// Bot holds the bot state
type Bot struct {
	mux               *corde.Mux
	Store             Store
	AnimeService      TrackingService
	Inter             Interacter
	AppID             corde.Snowflake
	GuildID           *corde.Snowflake
	BotToken          string
	PublicKey         string
	RollCooldown      time.Duration
	InteractionNeeded int64
	TokensNeeded      int32
}

// New runs the bot
func New(b *Bot) *corde.Mux {
	b.mux = corde.NewMux(b.PublicKey, b.AppID, b.BotToken)
	b.mux.OnNotFound = b.RemoveUnknownCommands

	t := trace[corde.SlashCommandInteractionData]
	i := interact(b.Inter, onInteraction[corde.SlashCommandInteractionData](b))

	b.mux.Route("give", b.give)
	b.mux.Route("search", b.search)
	b.mux.Route("profile", b.profile)
	b.mux.Route("verify", b.verify)
	b.mux.Route("exchange", b.exchange)
	b.mux.SlashCommand("list", wrap(b.list, t, i))
	b.mux.SlashCommand("roll", wrap(b.roll, t, i))
	b.mux.SlashCommand("info", wrap(b.info, t))
	b.mux.SlashCommand("claim", wrap(b.claim, t))

	return b.mux
}

func onInteraction[T corde.InteractionDataConstraint](b *Bot) func(ctx context.Context, count int64, i *corde.Interaction[T]) {
	return func(ctx context.Context, count int64, i *corde.Interaction[T]) {
		if count < b.InteractionNeeded {
			return
		}

		if b.GuildID != nil && *b.GuildID != i.GuildID {
			return
		}

		b.Inter.ResetInteractionCount(ctx, i.ChannelID)
		b.drop(ctx, i.ChannelID)
	}
}

// interaction middleware
func interact[T corde.InteractionDataConstraint](inter Interacter, interact func(ctx context.Context, count int64, i *corde.Interaction[T])) func(func(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[T])) func(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[T]) {
	return func(next func(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[T])) func(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[T]) {
		return func(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[T]) {
			go func() {
				ctx := context.Background()

				err := inter.IncrementInteractionCount(ctx, i.ChannelID)
				if err != nil {
					log.Debug().Err(err).Msg("failed to increment interaction count")
				}

				count, err := inter.GetInteractionCount(ctx, i.ChannelID)
				if err != nil {
					log.Err(err).Msg("failed to get interaction count")
					return
				}

				interact(ctx, count, i)
			}()

			next(ctx, w, i)
		}
	}
}

func wrap[T corde.InteractionDataConstraint](
	next func(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[T]),
	fns ...func(func(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[T])) func(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[T]),
) func(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[T]) {
	// apply middleware in reverse order
	for i := len(fns) - 1; i >= 0; i-- {
		next = fns[i](next)
	}
	return next
}

func (b *Bot) RemoveUnknownCommands(ctx context.Context, r corde.ResponseWriter, i *corde.Interaction[corde.JsonRaw]) {
	log.Error().Str("command", i.Route).Int("type", int(i.Type)).Msg("Unknown command")
	r.Respond(corde.NewResp().Content("I don't know what that means, you shouldn't be able to do that").Ephemeral())

	var opt []func(*corde.CommandsOpt)
	if b.GuildID != nil {
		opt = append(opt, corde.GuildOpt(*b.GuildID))
	}

	b.mux.DeleteCommand(i.ID, opt...)
}
