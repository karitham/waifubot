package discord

import (
	"context"
	"log/slog"
	"time"

	"github.com/Karitham/corde"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/karitham/waifubot/collection"
	"github.com/karitham/waifubot/guild"
	"github.com/karitham/waifubot/storage"
	"github.com/karitham/waifubot/storage/dropstore"
	"github.com/karitham/waifubot/storage/interactionstore"
	"github.com/karitham/waifubot/wishlist"
)

const (
	AnilistColor   = 0x02a9ff
	AnilistIconURL = "https://anilist.co/img/icons/favicon-32x32.png"
)

var (
	commandCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "waifubot_command_total",
			Help: "Total number of command invocations",
		},
		[]string{"command"},
	)
	commandDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "waifubot_command_duration_seconds",
			Help:    "Duration of command execution",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"command"},
	)
	errorCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "waifubot_error_total",
			Help: "Total number of errors",
		},
		[]string{"command", "error_type"},
	)
)

// TrackingService is the interface for the anilist service
type TrackingService interface {
	RandomChar(ctx context.Context, notIn ...int64) (collection.MediaCharacter, error)
	Anime(ctx context.Context, name string) ([]collection.Media, error)
	Manga(ctx context.Context, name string) ([]collection.Media, error)
	User(ctx context.Context, name string) ([]collection.TrackerUser, error)
	Character(ctx context.Context, name string) ([]collection.MediaCharacter, error)
	SearchMedia(ctx context.Context, search string) ([]collection.Media, error)
	GetMediaCharacters(ctx context.Context, mediaId int64) ([]collection.MediaCharacter, error)
}

// Bot holds the bot state
type Bot struct {
	mux               *corde.Mux
	Store             storage.Store
	WishlistStore     wishlist.Store
	AnimeService      TrackingService
	DropStore         dropstore.Store[collection.MediaCharacter]
	InterStore        interactionstore.Store
	GuildIndexer      *guild.Indexer
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
	prometheus.MustRegister(commandCounter, commandDuration, errorCounter)

	b.mux = corde.NewMux(b.PublicKey, b.AppID, b.BotToken)
	b.mux.OnNotFound = b.RemoveUnknownCommands

	t := trace[corde.SlashCommandInteractionData]
	i := interact(b.InterStore, onInteraction[corde.SlashCommandInteractionData](b))
	idx := indexMiddleware[corde.SlashCommandInteractionData](b)

	b.mux.Route("give", b.give)
	b.mux.Route("search", b.search)
	b.mux.Route("profile", b.profile)
	b.mux.Route("verify", b.verify)
	b.mux.Route("exchange", b.exchange)
	b.mux.Route("holders", b.holders)
	b.mux.Route("wishlist", b.wishlist)
	b.mux.Route("token", b.token)
	b.mux.SlashCommand("list", wrap(b.list, t, i, idx))
	b.mux.SlashCommand("roll", wrap(b.roll, t, i, idx))
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

		_ = b.InterStore.Reset(ctx, i.ChannelID)
		b.drop(ctx, i.ChannelID)
	}
}

// interaction middleware
func interact[T corde.InteractionDataConstraint](inter interactionstore.Store, interact func(ctx context.Context, count int64, i *corde.Interaction[T])) func(func(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[T])) func(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[T]) {
	return func(next func(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[T])) func(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[T]) {
		return func(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[T]) {
			go func() {
				ctx := context.Background()

				err := inter.Increment(ctx, i.ChannelID)
				if err != nil {
					slog.Debug("failed to increment interaction count", "error", err)
				}

				count, err := inter.Get(ctx, i.ChannelID)
				if err != nil {
					slog.Error("failed to get interaction count", "error", err)
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
	slog.Error("Unknown command", "command", i.Route, "type", int(i.Type))
	r.Respond(corde.NewResp().Content("I don't know what that means, you shouldn't be able to do that").Ephemeral())

	var opt []func(*corde.CommandsOpt)
	if b.GuildID != nil {
		opt = append(opt, corde.GuildOpt(*b.GuildID))
	}

	_ = b.mux.DeleteCommand(i.ID, opt...)
}
