package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/Karitham/corde"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/urfave/cli/v2"

	"github.com/karitham/waifubot/anilist"
	"github.com/karitham/waifubot/discord"
	"github.com/karitham/waifubot/guild"
	"github.com/karitham/waifubot/rest"
	"github.com/karitham/waifubot/rest/api"
	"github.com/karitham/waifubot/services"
	"github.com/karitham/waifubot/storage"
	"github.com/karitham/waifubot/storage/dropstore"
	"github.com/karitham/waifubot/storage/interactionstore"
	"github.com/karitham/waifubot/wishlist"
)

var RunCommand = &cli.Command{
	Name:  "run",
	Usage: "Run the Discord bot and REST API server",
	Flags: []cli.Flag{
		botTokenFlag,
		&cli.Uint64Flag{
			Name:    "guild-id",
			EnvVars: []string{"GUILD_ID"},
		},
		appIDFlag,
		&cli.StringFlag{
			Name:     "public-key",
			EnvVars:  []string{"DISCORD_PUBLIC_KEY", "PUBLIC_KEY"},
			Required: true,
		},
		dbURLFlag,
		rollCooldownFlag,
		tokensNeededFlag,
		&cli.Int64Flag{
			Name:        "interaction-needed",
			EnvVars:     []string{"INTERACTION_NEEDED"},
			DefaultText: "25",
		},
		anilistMaxCharsFlag,
		&cli.StringFlag{
			Name:    "port",
			EnvVars: []string{"PORT"},
			Value:   "8080",
		},
		&cli.BoolFlag{
			Name:    "skip-migrate",
			Usage:   "Skip database migrations on startup",
			EnvVars: []string{"SKIP_MIGRATE"},
		},
		logLevelFlag,
		apiFlag,
	},
	Action: func(c *cli.Context) error {
		ctx := c.Context

		if !c.Bool("skip-migrate") {
			if err := storage.Migrate(c.String(dbURLFlag.Name)); err != nil {
				return fmt.Errorf("error running migrations: %w", err)
			}
		}

		store, err := storage.NewStore(ctx, c.String(dbURLFlag.Name))
		if err != nil {
			return fmt.Errorf("error connecting to db: %w", err)
		}

		var guildID *corde.Snowflake
		if gid := c.Uint64("guild-id"); gid != 0 {
			id := corde.Snowflake(gid)
			guildID = &id
		}

		interStore := interactionstore.NewPostgresStore(store.InteractionStore())
		dropStore := dropstore.NewPostgresStore(store.DropStore())

		slog.Info("Starting WaifuBot", "port", c.String("port"), "app_id", c.String("app-id"), "api_enabled", c.Bool(apiFlag.Name))
		mux := discord.New(&discord.Bot{
			Store:             store,
			WishlistStore:     wishlist.New(store.WishlistStore()),
			AnimeService:      anilist.New(anilist.MaxChar(c.Int64(anilistMaxCharsFlag.Name))),
			DropStore:         dropStore,
			InterStore:        interStore,
			GuildIndexer:      guild.NewIndexer(store, c.String(botTokenFlag.Name)),
			AppID:             corde.Snowflake(c.Uint64("app-id")),
			GuildID:           guildID,
			BotToken:          c.String(botTokenFlag.Name),
			PublicKey:         c.String("public-key"),
			RollCooldown:      c.Duration(rollCooldownFlag.Name),
			InteractionNeeded: c.Int64("interaction-needed"),
			TokensNeeded:      int32(c.Int(tokensNeededFlag.Name)),
		})

		port, err := strconv.Atoi(c.String("port"))
		if err != nil {
			return fmt.Errorf("invalid port number: %w", err)
		}

		r := chi.NewRouter()
		r.Use(middleware.Timeout(5 * time.Second))
		r.Use(rest.LoggerMiddleware(slog.Default()))
		r.Use(middleware.Compress(5))

		r.Use(cors.Handler(cors.Options{
			AllowedOrigins:   []string{"https://*", "http://*"},
			AllowedMethods:   []string{"GET", "OPTIONS"},
			AllowedHeaders:   []string{"Accept", "Content-Type"},
			MaxAge:           300,
			AllowCredentials: true,
		}))

		r.Handle("/metrics", promhttp.Handler())
		r.Handle("/", mux)

		if c.Bool(apiFlag.Name) {
			discordToken := c.String(botTokenFlag.Name)
			var discordService *services.DiscordService
			if discordToken != "" {
				discordService = services.NewDiscordService(discordToken)
			}

			restServer := rest.New(store, discordService)

			telemetry, err := rest.SetupTelemetry(prometheus.DefaultRegisterer)
			if err != nil {
				return fmt.Errorf("failed to setup telemetry: %w", err)
			}
			defer func() {
				if err := telemetry.Shutdown(c.Context); err != nil {
					slog.Error("Error shutting down telemetry", "error", err)
				}
			}()

			apiRouter, err := api.NewServer(
				restServer,
				api.WithMeterProvider(telemetry.MeterProvider()),
			)
			if err != nil {
				return fmt.Errorf("failed to create API router: %w", err)
			}

			r.Mount("/", apiRouter)
			slog.Info("REST API server started", "port", port)
		}

		slog.Info("Discord bot started", "port", port)

		if err := http.ListenAndServe(":"+strconv.Itoa(port), r); err != nil {
			slog.Error("Server crashed", "error", err, "port", port)
			return err
		}

		slog.Info("Server shutting down", "port", port)
		return nil
	},
}
