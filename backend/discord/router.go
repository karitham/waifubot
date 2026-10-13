package discord

import (
	"context"
	"log/slog"

	"github.com/Karitham/corde"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/karitham/waifubot/collection"
	"github.com/karitham/waifubot/drops"
	"github.com/karitham/waifubot/guild"
	"github.com/karitham/waifubot/tracker"
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

// CommandStore handles discord command hash persistence.
type CommandStore interface {
	GetCommandHash(ctx context.Context) (string, error)
	SetCommandHash(ctx context.Context, hash string) error
	UpdateCommandHash(ctx context.Context, hash string) error
}

// Config is the bot's Discord identity and command tuning, set from CLI flags.
type Config struct {
	AppID             corde.Snowflake
	GuildID           *corde.Snowflake
	BotToken          string
	PublicKey         string
	RollConfig        collection.Config
	InteractionNeeded int64
}

// Deps are the adapters the command handlers run against.
type Deps struct {
	Store        collection.DB
	Drops        drops.DB
	CommandStore CommandStore
	AnimeService tracker.Service
	GuildIndexer *guild.Indexer
}

// New wires the bot: it registers the command definitions when their hash
// changed, builds every command handler, and returns the mux serving them.
func New(ctx context.Context, cfg Config, deps Deps) (*corde.Mux, error) {
	prometheus.MustRegister(commandCounter, commandDuration)

	mux := corde.NewMux(cfg.PublicKey, cfg.AppID, cfg.BotToken)
	if err := migrateCommands(ctx, mux, cfg.GuildID, deps.CommandStore); err != nil {
		return nil, err
	}

	mux.OnNotFound = removeUnknownCommands(mux, cfg.GuildID)

	trigger := &dropTrigger{
		store:     deps.Drops,
		mux:       mux,
		threshold: cfg.InteractionNeeded,
		guildID:   cfg.GuildID,
	}

	t := trace[corde.SlashCommandInteractionData]
	i := onInteraction(trigger.record)
	idx := indexMiddleware[corde.SlashCommandInteractionData](deps.GuildIndexer)

	infoHandler := &InfoHandler{}
	claimHandler := &ClaimHandler{store: deps.Store}
	listHandler := &ListHandler{store: deps.Store}
	giveHandler := &GiveHandler{store: deps.Store}
	verifyHandler := &VerifyHandler{store: deps.Store, guildIndexer: deps.GuildIndexer}
	profileHandler := &ProfileHandler{store: deps.Store}
	searchHandler := &SearchHandler{
		animeService:  deps.AnimeService,
		onInteraction: i,
		guildIndexer:  deps.GuildIndexer,
	}
	holdersHandler := &HoldersHandler{store: deps.Store, guildIndexer: deps.GuildIndexer}
	rollHandler := &RollHandler{
		store:  deps.Store,
		config: cfg.RollConfig,
	}
	tokenHandler := &TokenHandler{
		store:        deps.Store,
		animeService: deps.AnimeService,
		config:       cfg.RollConfig,
	}
	wishlistHandler := &WishlistHandler{
		store:        deps.Store,
		animeService: deps.AnimeService,
		guildIndexer: deps.GuildIndexer,
	}

	mux.SlashCommand("info", wrap(wrapCtx(infoHandler.Info), t))
	mux.SlashCommand("claim", wrap(wrapCtx(claimHandler.Claim), t))
	mux.SlashCommand("list", wrap(wrapCtx(listHandler.List), t, i, idx))
	mux.Route("give", giveHandler.Register)
	mux.Route("verify", verifyHandler.Register)
	mux.Route("profile", profileHandler.Register)
	mux.Route("search", searchHandler.Register)
	mux.Route("holders", holdersHandler.Register)
	mux.SlashCommand("roll", wrap(wrapCtx(rollHandler.Roll), t, i, idx))
	mux.Route("token", tokenHandler.Register)
	mux.Route("wishlist", wishlistHandler.Register)

	return mux, nil
}

// removeUnknownCommands answers commands Discord sends that the bot no longer
// defines and deletes them.
func removeUnknownCommands(mux *corde.Mux, guildID *corde.Snowflake) func(context.Context, corde.ResponseWriter, *corde.Interaction[corde.JsonRaw]) {
	return func(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[corde.JsonRaw]) {
		slog.Error("Unknown command", "command", i.Route, "type", int(i.Type))
		w.Respond(corde.NewResp().Content("I don't know what that means, you shouldn't be able to do that").Ephemeral())

		var opt []func(*corde.CommandsOpt)
		if guildID != nil {
			opt = append(opt, corde.GuildOpt(*guildID))
		}

		_ = mux.DeleteCommand(i.ID, opt...)
	}
}
