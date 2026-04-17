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
	"github.com/karitham/waifubot/auth"
	"github.com/karitham/waifubot/discord"
	"github.com/karitham/waifubot/guild"
	"github.com/karitham/waifubot/rest"
	"github.com/karitham/waifubot/rest/api"
	"github.com/karitham/waifubot/services"
	"github.com/karitham/waifubot/storage"
	"github.com/karitham/waifubot/storage/commandpg"
	"github.com/karitham/waifubot/storage/dropstore"
	"github.com/karitham/waifubot/storage/interactionstore"
	"github.com/karitham/waifubot/sync"
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
		seriesRollCostFlag,
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
		&cli.BoolFlag{
			Name:    "sync",
			Usage:   "Enable background character sync worker. Can be quite heavy on Anilist",
			EnvVars: []string{"SYNC"},
			Value:   false,
		},
		logLevelFlag,
		apiFlag,
		&cli.StringFlag{
			Name:    "oauth-client-secret",
			EnvVars: []string{"OAUTH_CLIENT_SECRET"},
			Usage:   "Discord OAuth2 client secret",
		},
		&cli.StringFlag{
			Name:    "oauth-redirect-uri",
			EnvVars: []string{"OAUTH_REDIRECT_URI"},
			Usage:   "OAuth2 callback URL",
			Value:   "http://localhost:8080/auth/callback",
		},
		&cli.StringFlag{
			Name:    "frontend-url",
			EnvVars: []string{"FRONTEND_URL"},
			Usage:   "Frontend URL for OAuth callback redirect",
			Value:   "http://localhost:3000",
		},
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
		collStore := newCollectionStore(store)
		wishStore := wishlist.New(store.WishlistStore())
		catalogStore := newCatalogStore(store)

		anilistClient := anilist.New(anilist.MaxChar(c.Int64(anilistMaxCharsFlag.Name)))

		slog.Info("Starting WaifuBot", "port", c.String("port"), "app_id", c.String("app-id"), "api_enabled", c.Bool(apiFlag.Name))
		router := discord.New(&discord.Router{
			Store:             collStore,
			Catalog:           catalogStore,
			CommandStore:      commandpg.New(store.CommandStore()),
			WishlistStore:     wishStore,
			AnimeService:      anilistClient,
			DropStore:         dropStore,
			InterStore:        interStore,
			GuildIndexer:      guild.NewIndexer(collStore, guild.NewDiscordFetcher(c.String(botTokenFlag.Name))),
			GuildOps:          collStore,
			AppID:             corde.Snowflake(c.Uint64("app-id")),
			GuildID:           guildID,
			BotToken:          c.String(botTokenFlag.Name),
			PublicKey:         c.String("public-key"),
			RollCooldown:      c.Duration(rollCooldownFlag.Name),
			InteractionNeeded: c.Int64("interaction-needed"),
			SeriesRollCost:    int32(c.Int(seriesRollCostFlag.Name)),
		})
		mux := router.Register()

		// Start background sync worker if enabled
		if c.Bool("sync") {
			go func() {
				slog.Info("character sync worker started")
				sync.NewService(catalogStore, anilistClient).Run(ctx)
				slog.Info("character sync worker stopped")
			}()
		}

		port := c.Int("port")

		r := chi.NewRouter()
		r.Use(middleware.Timeout(5 * time.Second))
		r.Use(rest.LoggerMiddleware(slog.Default()))
		r.Use(middleware.Compress(5))

		r.Use(cors.Handler(cors.Options{
			AllowedOrigins:   []string{"https://*", "http://*"},
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "If-None-Match"},
			MaxAge:           300,
			AllowCredentials: false,
		}))

		r.Handle("/metrics", promhttp.Handler())
		r.Handle("/", mux)

		if c.Bool(apiFlag.Name) {
			discordToken := c.String(botTokenFlag.Name)
			var discordService *services.DiscordService
			if discordToken != "" {
				discordService = services.NewDiscordService(discordToken)
			}

			restServer := rest.New(collStore, wishStore, discordService)

			telemetry, err := rest.SetupTelemetry(prometheus.DefaultRegisterer)
			if err != nil {
				return fmt.Errorf("failed to setup telemetry: %w", err)
			}
			defer func() {
				if err := telemetry.Shutdown(c.Context); err != nil {
					slog.Error("Error shutting down telemetry", "error", err)
				}
			}()

			// Session store and OAuth
			sessionStore := auth.NewPgSessionStore(store.SessionStore())
			oauthClient := auth.NewOAuthClient(auth.OAuthConfig{
				ClientID:     c.String("app-id"),
				ClientSecret: c.String("oauth-client-secret"),
				RedirectURI:  c.String("oauth-redirect-uri"),
			})

			// Auth routes (chi, not ogen)
			authHandler := auth.NewHandler(
				sessionStore,
				auth.NewUserStoreAdapter(collStore),
				oauthClient,
				c.String("frontend-url"),
			)
			r.Route("/auth", authHandler.Routes)

			// Security handler for ogen
			securityHandler := rest.NewSecurityHandler(sessionStore)

			apiRouter, err := api.NewServer(
				restServer,
				securityHandler,
				api.WithMeterProvider(telemetry.MeterProvider()),
			)
			if err != nil {
				return fmt.Errorf("failed to create API router: %w", err)
			}

			r.Mount("/", rest.ETagMiddleware(apiRouter))
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
