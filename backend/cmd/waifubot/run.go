package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
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
			Usage:   "Background character sync from AniList (5 req/min)",
			EnvVars: []string{"SYNC"},
			Value:   true,
		},
		logLevelFlag,
		apiFlag,
		&cli.StringFlag{
			Name:    "oauth-client-id",
			EnvVars: []string{"DISCORD_OAUTH_CLIENT_ID", "DISCORD_CLIENT_ID", "OAUTH_CLIENT_ID"},
			Usage:   "Discord OAuth application client ID (for frontend login)",
		},
		&cli.StringFlag{
			Name:    "oauth-client-secret",
			EnvVars: []string{"DISCORD_OAUTH_CLIENT_SECRET", "DISCORD_CLIENT_SECRET", "OAUTH_CLIENT_SECRET"},
			Usage:   "Discord OAuth application client secret",
		},
		&cli.StringFlag{
			Name:    "oauth-redirect-url",
			EnvVars: []string{"OAUTH_REDIRECT_URL", "DISCORD_OAUTH_REDIRECT_URL", "DISCORD_REDIRECT_URL"},
			Usage:   "Discord OAuth redirect URL, e.g. https://api.example.com/api/v1/auth/callback",
		},
		&cli.StringFlag{
			Name:    "allowed-frontend-origins",
			EnvVars: []string{"ALLOWED_FRONTEND_ORIGINS"},
			Usage:   "Comma-separated list of origins the browser may be redirected to after Discord auth. Defaults to the origin of oauth-redirect-url.",
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

		anilistClient := anilist.New()

		slog.Info("Starting WaifuBot", "port", c.String("port"), "app_id", c.String("app-id"), "api_enabled", c.Bool(apiFlag.Name))
		slog.Info("OAuth login config",
			"client_id_set", c.String("oauth-client-id") != "",
			"client_secret_set", c.String("oauth-client-secret") != "",
			"redirect_url", c.String("oauth-redirect-url"),
			"allowed_origins", parseAllowedOrigins(c.String("allowed-frontend-origins"), c.String("oauth-redirect-url")),
		)
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
				sync.NewService(catalogStore, anilistClient, anilistClient).Run(ctx)
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
			AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
			AllowedHeaders:   []string{"Accept", "Content-Type", "If-None-Match", "Authorization"},
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

			restServer := rest.New(collStore, wishStore, discordService,
				store.AuthStore(),
				store.UserStore(),
				rest.OAuthConfig{
					ClientID:       c.String("oauth-client-id"),
					ClientSecret:   c.String("oauth-client-secret"),
					RedirectURL:    c.String("oauth-redirect-url"),
					AllowedOrigins: parseAllowedOrigins(c.String("allowed-frontend-origins"), c.String("oauth-redirect-url")),
				},
			)

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
				restServer,
				api.WithMeterProvider(telemetry.MeterProvider()),
			)
			if err != nil {
				return fmt.Errorf("failed to create API router: %w", err)
			}

			// Auth login + callback are browser-side OAuth redirects, not API
			// endpoints. Mount them on chi so we can use http.Redirect directly.
			r.Method("GET", "/api/v1/auth/login", http.HandlerFunc(restServer.HandleAuthLogin))
			r.Method("GET", "/api/v1/auth/callback", http.HandlerFunc(restServer.HandleAuthCallback))

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

// parseAllowedOrigins splits a comma-separated ALLOWED_FRONTEND_ORIGINS value.
// If empty, falls back to the origin of the OAuth redirect URL (covers the
// same-origin case where the frontend and API share a host).
func parseAllowedOrigins(s, redirectURL string) []string {
	if s != "" {
		var out []string
		for _, p := range strings.Split(s, ",") {
			p = strings.TrimSpace(p)
			if p != "" {
				out = append(out, p)
			}
		}
		return out
	}
	u, err := url.Parse(redirectURL)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return nil
	}
	return []string{u.Scheme + "://" + u.Host}
}
