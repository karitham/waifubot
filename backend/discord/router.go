package discord

import (
	"context"
	"log/slog"
	"time"

	"github.com/Karitham/corde"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/karitham/waifubot/collection"
	"github.com/karitham/waifubot/guild"
	"github.com/karitham/waifubot/storage/interactionstore"
	"github.com/karitham/waifubot/tracker"
	"github.com/karitham/waifubot/wishlist"

	"github.com/karitham/waifubot/catalog"
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
)

// TrackingService is the interface for the anilist service.
type TrackingService interface {
	Anime(ctx context.Context, name string) ([]tracker.Media, error)
	Manga(ctx context.Context, name string) ([]tracker.Media, error)
	User(ctx context.Context, name string) ([]tracker.User, error)
	Character(ctx context.Context, name string) ([]tracker.Character, error)
	SearchMedia(ctx context.Context, search string) ([]tracker.Media, error)
	GetMediaCharacters(ctx context.Context, mediaID tracker.MediaID) ([]tracker.Character, error)
}

// CommandStore handles discord command hash persistence.
type CommandStore interface {
	GetCommandHash(ctx context.Context) (string, error)
	SetCommandHash(ctx context.Context, hash string) error
	UpdateCommandHash(ctx context.Context, hash string) error
}

// Router holds the infrastructure needed to wire command handlers.
type Router struct {
	mux               *corde.Mux
	Store             collection.DB
	Catalog           catalog.Reader
	CommandStore      CommandStore
	WishlistStore     wishlist.Store
	AnimeService      TrackingService
	InterStore        interactionstore.Store
	GuildIndexer      *guild.Indexer
	GuildOps          guild.Querier
	AppID             corde.Snowflake
	GuildID           *corde.Snowflake
	BotToken          string
	PublicKey         string
	RollCooldown      time.Duration
	InteractionNeeded int64
	SeriesRollCost    collection.TokenAmount
}

// New constructs a Router with all dependencies and runs command migration.
func New(r *Router) *Router {
	prometheus.MustRegister(commandCounter, commandDuration)

	r.MustMigrateCommands()

	return r
}

// Register sets up the mux, middleware, and all command routes.
// Returns the configured *corde.Mux.
func (r *Router) Register() *corde.Mux {
	r.mux = corde.NewMux(r.PublicKey, r.AppID, r.BotToken)
	r.mux.OnNotFound = r.RemoveUnknownCommands

	t := trace[corde.SlashCommandInteractionData]
	i := interact(r.InterStore, r.onInteraction)
	idx := indexMiddleware[corde.SlashCommandInteractionData](r.GuildIndexer)

	// Construct handlers
	infoHandler := &InfoHandler{}
	claimHandler := &ClaimHandler{store: r.Store}
	listHandler := &ListHandler{store: r.Store}
	giveHandler := &GiveHandler{store: r.Store}
	verifyHandler := &VerifyHandler{store: r.Store, guildIndexer: r.GuildIndexer}
	profileHandler := &ProfileHandler{store: r.Store}
	searchHandler := &SearchHandler{
		animeService:  r.AnimeService,
		interStore:    r.InterStore,
		onInteraction: r.onInteraction,
		guildIndexer:  r.GuildIndexer,
	}
	holdersHandler := &HoldersHandler{guildOps: r.GuildOps, catalog: r.Catalog, guildIndexer: r.GuildIndexer}
	rollConfig := collection.Config{RollCooldown: r.RollCooldown, SeriesRollCost: r.SeriesRollCost}
	rollHandler := &RollHandler{
		store:    r.Store,
		config:   rollConfig,
		wishlist: r.WishlistStore,
	}
	tokenHandler := &TokenHandler{
		store:        r.Store,
		animeService: r.AnimeService,
		config:       rollConfig,
	}
	wishlistHandler := &WishlistHandler{
		wishlist:     r.WishlistStore,
		store:        r.Store,
		animeService: r.AnimeService,
		catalog:      r.Catalog,
		guildIndexer: r.GuildIndexer,
	}

	// Register routes
	r.mux.SlashCommand("info", wrap(wrapCtx(infoHandler.Info), t))
	r.mux.SlashCommand("claim", wrap(wrapCtx(claimHandler.Claim), t))
	r.mux.SlashCommand("list", wrap(wrapCtx(listHandler.List), t, i, idx))
	r.mux.Route("give", giveHandler.Register)
	r.mux.Route("verify", verifyHandler.Register)
	r.mux.Route("profile", profileHandler.Register)
	r.mux.Route("search", searchHandler.Register)
	r.mux.Route("holders", holdersHandler.Register)
	r.mux.SlashCommand("roll", wrap(wrapCtx(rollHandler.Roll), t, i, idx))
	r.mux.Route("token", tokenHandler.Register)
	r.mux.Route("wishlist", wishlistHandler.Register)

	return r.mux
}

func (r *Router) onInteraction(ctx context.Context, count int64, i *corde.Interaction[corde.SlashCommandInteractionData]) {
	if count < r.InteractionNeeded {
		return
	}

	if r.GuildID != nil && *r.GuildID != i.GuildID {
		return
	}

	_ = r.InterStore.Reset(ctx, i.ChannelID)
	r.drop(ctx, i.ChannelID)
}

func (r *Router) RemoveUnknownCommands(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[corde.JsonRaw]) {
	slog.Error("Unknown command", "command", i.Route, "type", int(i.Type))
	w.Respond(corde.NewResp().Content("I don't know what that means, you shouldn't be able to do that").Ephemeral())

	var opt []func(*corde.CommandsOpt)
	if r.GuildID != nil {
		opt = append(opt, corde.GuildOpt(*r.GuildID))
	}

	_ = r.mux.DeleteCommand(i.ID, opt...)
}
