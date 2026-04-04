package discord

import (
	"context"
	"log/slog"
	"time"

	"github.com/Karitham/corde"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/karitham/waifubot/collection"
	"github.com/karitham/waifubot/guild"
	"github.com/karitham/waifubot/storage/dropstore"
	"github.com/karitham/waifubot/storage/interactionstore"
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
	RandomChar(ctx context.Context, notIn ...int64) (collection.MediaCharacter, error)
	Anime(ctx context.Context, name string) ([]collection.Media, error)
	Manga(ctx context.Context, name string) ([]collection.Media, error)
	User(ctx context.Context, name string) ([]collection.TrackerUser, error)
	Character(ctx context.Context, name string) ([]collection.MediaCharacter, error)
	SearchMedia(ctx context.Context, search string) ([]collection.Media, error)
	GetMediaCharacters(ctx context.Context, mediaId int64) ([]collection.MediaCharacter, error)
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
	Store             collection.Store
	Catalog           catalog.Store
	CommandStore      CommandStore
	WishlistStore     wishlist.Store
	AnimeService      TrackingService
	DropStore         dropstore.Store
	InterStore        interactionstore.Store
	GuildIndexer      *guild.Indexer
	GuildOps          guild.GuildQuerier
	guildTxFn         func(context.Context) (guild.TxQuerier, error)
	AppID             corde.Snowflake
	GuildID           *corde.Snowflake
	BotToken          string
	PublicKey         string
	RollCooldown      time.Duration
	InteractionNeeded int64
	SeriesRollCost    int32
}

// New constructs a Router with all dependencies and runs command migration.
func New(r *Router) *Router {
	prometheus.MustRegister(commandCounter, commandDuration)

	// Wire up guild transaction factory — collection.Store already satisfies guild.TxQuerier
	r.guildTxFn = func(ctx context.Context) (guild.TxQuerier, error) {
		return r.Store.WithTx(ctx)
	}

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
	idx := indexMiddleware[corde.SlashCommandInteractionData](r.GuildIndexer, r.guildTxFn)

	// Construct handlers
	infoHandler := &InfoHandler{}
	claimHandler := &ClaimHandler{store: r.Store}
	listHandler := &ListHandler{store: r.Store}
	giveHandler := &GiveHandler{store: r.Store}
	verifyHandler := &VerifyHandler{store: r.Store, guildIndexer: r.GuildIndexer, guildTxFn: r.guildTxFn}
	profileHandler := &ProfileHandler{store: r.Store}
	searchHandler := &SearchHandler{
		animeService:  r.AnimeService,
		interStore:    r.InterStore,
		onInteraction: r.onInteraction,
		guildIndexer:  r.GuildIndexer,
		guildTxFn:     r.guildTxFn,
	}
	holdersHandler := &HoldersHandler{guildOps: r.GuildOps, catalog: r.Catalog, guildIndexer: r.GuildIndexer, guildTxFn: r.guildTxFn}
	rollHandler := &RollHandler{
		rollService: collection.NewRollService(r.Store, r.AnimeService, collection.RollConfig{RollCooldown: r.RollCooldown}),
		wishlist:    r.WishlistStore,
	}
	tokenHandler := &TokenHandler{
		store:        r.Store,
		animeService: r.AnimeService,
		rollService:  collection.NewRollService(r.Store, r.AnimeService, collection.RollConfig{RollCooldown: r.RollCooldown}),
		config:       collection.Config{RollCooldown: r.RollCooldown, SeriesRollCost: r.SeriesRollCost},
	}
	wishlistHandler := &WishlistHandler{
		wishlist:     r.WishlistStore,
		store:        r.Store,
		animeService: r.AnimeService,
		catalog:      r.Catalog,
		guildIndexer: r.GuildIndexer,
		guildTxFn:    r.guildTxFn,
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
