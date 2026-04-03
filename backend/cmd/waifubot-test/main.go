package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Karitham/corde"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5"
	"github.com/urfave/cli/v2"

	"github.com/karitham/waifubot/catalog"
	"github.com/karitham/waifubot/cmd/waifubot/flags"
	"github.com/karitham/waifubot/collection"
	"github.com/karitham/waifubot/discord"
	"github.com/karitham/waifubot/guild"
	"github.com/karitham/waifubot/storage"
	"github.com/karitham/waifubot/storage/catalogpg"
	"github.com/karitham/waifubot/storage/collectionpg"
	"github.com/karitham/waifubot/storage/commandpg"
	"github.com/karitham/waifubot/storage/droppg"
	"github.com/karitham/waifubot/storage/dropstore"
	"github.com/karitham/waifubot/storage/guildpg"
	"github.com/karitham/waifubot/storage/interactionstore"
	"github.com/karitham/waifubot/storage/userpg"
	"github.com/karitham/waifubot/wishlist"
)

// Test-specific flags
var (
	testCharIDFlag = &cli.Int64Flag{
		Name:    "test-char-id",
		EnvVars: []string{"TEST_CHAR_ID"},
		Value:   99999,
	}
	testCharNameFlag = &cli.StringFlag{
		Name:    "test-char-name",
		EnvVars: []string{"TEST_CHAR_NAME"},
		Value:   "Test Character",
	}
	testCharImageFlag = &cli.StringFlag{
		Name:    "test-char-image",
		EnvVars: []string{"TEST_CHAR_IMAGE"},
		Value:   "https://example.com/img.png",
	}
	testCharMediaFlag = &cli.StringFlag{
		Name:    "test-char-media",
		EnvVars: []string{"TEST_CHAR_MEDIA"},
		Value:   "Test Anime",
	}
	testWishingUsersFlag = &cli.StringFlag{
		Name:    "test-wishing-users",
		EnvVars: []string{"TEST_WISHING_USERS"},
		Value:   "",
	}
	testGuildIDFlag = &cli.Uint64Flag{
		Name:    "test-guild-id",
		EnvVars: []string{"TEST_GUILD_ID"},
		Value:   0,
	}
	portFlag = &cli.StringFlag{
		Name:    "port",
		EnvVars: []string{"PORT"},
		Value:   "8080",
	}
)

func main() {
	app := &cli.App{
		Name:        "waifubot-test",
		Usage:       "Run the bot test harness",
		Description: "Test harness for waifubot with fake services",
		Commands: []*cli.Command{
			{
				Name:  "run",
				Usage: "Run the test harness",
				Flags: []cli.Flag{
					flags.BotTokenFlag,
					flags.DbURLFlag,
					flags.AppIDFlag,
					&cli.StringFlag{
						Name:     "public-key",
						EnvVars:  []string{"DISCORD_PUBLIC_KEY", "PUBLIC_KEY"},
						Required: true,
					},
					flags.RollCooldownFlag,
					flags.TokensNeededFlag,
					&cli.Int64Flag{
						Name:        "interaction-needed",
						EnvVars:     []string{"INTERACTION_NEEDED"},
						DefaultText: "25",
					},
					testCharIDFlag,
					testCharNameFlag,
					testCharImageFlag,
					testCharMediaFlag,
					testWishingUsersFlag,
					testGuildIDFlag,
					portFlag,
					&cli.BoolFlag{
						Name:    "skip-migrate",
						Usage:   "Skip database migrations on startup",
						EnvVars: []string{"SKIP_MIGRATE"},
					},
					flags.LogLevelFlag,
				},
				Action: runTestHarness,
			},
		},
		DefaultCommand: "run",
	}

	if err := app.Run(os.Args); err != nil {
		slog.Error("error running app", "error", err)
		os.Exit(1)
	}
}

func runTestHarness(c *cli.Context) error {
	ctx := c.Context

	dbURL := c.String(flags.DbURLFlag.Name)
	if dbURL == "" {
		return fmt.Errorf("db-url is required")
	}

	publicKey := c.String("public-key")
	if publicKey == "" {
		return fmt.Errorf("public-key is required")
	}

	testCharID := c.Int64(testCharIDFlag.Name)
	testCharName := c.String(testCharNameFlag.Name)
	testCharImage := c.String(testCharImageFlag.Name)
	testCharMedia := c.String(testCharMediaFlag.Name)
	testWishingUsers := c.String(testWishingUsersFlag.Name)
	testGuildID := c.Uint64(testGuildIDFlag.Name)

	// Run migrations
	if !c.Bool("skip-migrate") {
		if err := storage.Migrate(dbURL); err != nil {
			return fmt.Errorf("error running migrations: %w", err)
		}
	}

	// Connect to DB
	store, err := storage.NewStore(ctx, dbURL)
	if err != nil {
		return fmt.Errorf("error connecting to db: %w", err)
	}

	// Seed test data
	if err := seedCharacter(ctx, store, testCharID, testCharName, testCharImage); err != nil {
		return fmt.Errorf("error seeding character: %w", err)
	}

	// Parse wishing users
	var wishingUserIDs []uint64
	if testWishingUsers != "" {
		for idStr := range strings.SplitSeq(testWishingUsers, ",") {
			idStr = strings.TrimSpace(idStr)
			if idStr == "" {
				continue
			}
			id, err := strconv.ParseUint(idStr, 10, 64)
			if err != nil {
				slog.Warn("invalid user ID", "id", idStr, "error", err)
				continue
			}
			wishingUserIDs = append(wishingUserIDs, id)
		}
	}

	if len(wishingUserIDs) > 0 {
		if err := seedWishlistUsers(ctx, store, wishingUserIDs, testCharID); err != nil {
			return fmt.Errorf("error seeding wishlist users: %w", err)
		}
	}

	// Seed guild members (the wishing users should be in the guild)
	if testGuildID != 0 && len(wishingUserIDs) > 0 {
		if err := seedGuildMembers(ctx, store, testGuildID, wishingUserIDs); err != nil {
			return fmt.Errorf("error seeding guild members: %w", err)
		}
	}

	// Create fake tracking service
	fakeService := &FakeTrackingService{
		CharID:     testCharID,
		CharName:   testCharName,
		CharImage:  testCharImage,
		MediaTitle: testCharMedia,
	}

	// Create stores (mirroring run.go)
	interStore := interactionstore.NewPostgresStore(store.InteractionStore())
	dropStore := dropstore.NewPostgresStore(store.DropStore())
	collStore := newCollectionStore(store)
	wishStore := wishlist.New(store.WishlistStore())

	// Get bot token for command registration
	botToken := c.String(flags.BotTokenFlag.Name)

	// Get app ID (optional, defaults to 0 for test)
	var appID uint64
	if appIDStr := c.String(flags.AppIDFlag.Name); appIDStr != "" {
		appID, _ = strconv.ParseUint(appIDStr, 10, 64)
	}

	slog.Info("Starting WaifuBot Test Harness",
		"port", c.String(portFlag.Name),
		"test_char_id", testCharID,
		"test_char_name", testCharName,
		"test_wishing_users", testWishingUsers,
		"test_guild_id", testGuildID,
	)

	// Create bot with fake AnimeService
	router := discord.New(&discord.Router{
		Store:             collStore,
		Catalog:           newCatalogStore(store),
		CommandStore:      commandpg.New(store.CommandStore()),
		WishlistStore:     wishStore,
		AnimeService:      fakeService,
		DropStore:         dropStore,
		InterStore:        interStore,
		GuildIndexer:      guild.NewIndexer(collStore, guild.NewDiscordFetcher(botToken)),
		GuildOps:          collStore,
		AppID:             corde.Snowflake(appID),
		GuildID:           nil,
		BotToken:          botToken,
		PublicKey:         publicKey,
		RollCooldown:      c.Duration(flags.RollCooldownFlag.Name),
		InteractionNeeded: c.Int64("interaction-needed"),
		TokensNeeded:      int32(c.Int(flags.TokensNeededFlag.Name)),
	})
	mux := router.Register()

	// Create HTTP server
	portNum, err := strconv.Atoi(c.String(portFlag.Name))
	if err != nil {
		return fmt.Errorf("invalid port number: %w", err)
	}

	r := chi.NewRouter()
	r.Use(middleware.Timeout(5 * time.Second))
	r.Use(middleware.Compress(5))

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Content-Type", "If-None-Match"},
		MaxAge:           300,
		AllowCredentials: true,
	}))

	r.Handle("/", mux)

	slog.Info("Test harness server started", "port", portNum)

	if err := http.ListenAndServe(":"+strconv.Itoa(portNum), r); err != nil {
		return fmt.Errorf("server crashed: %w", err)
	}

	return nil
}

// Wire functions copied from cmd/waifubot/wire.go

// newCollectionStore wires adapters into a collection.Store.
func newCollectionStore(s storage.Store) collection.Store {
	return newStoreFromStorage(s)
}

func newStoreFromStorage(s storage.Store) *collection.PostgresStore {
	catQ := s.CollectionStore()

	txFn := func(tx pgx.Tx) collection.Store {
		return collection.NewPostgresStore(
			userpg.New(s.UserStore()),
			collectionpg.New(catQ, s.WishlistStore()),
			droppg.New(s.DropStore()),
			guildpg.New(s.GuildStore()),
			catalogpg.New(catQ, s.GuildStore()),
			tx,
			nil,
		)
	}

	return collection.NewPostgresStore(
		userpg.New(s.UserStore()),
		collectionpg.New(catQ, s.WishlistStore()),
		droppg.New(s.DropStore()),
		guildpg.New(s.GuildStore()),
		catalogpg.New(catQ, s.GuildStore()),
		s.DB(),
		txFn,
	)
}

// newCatalogStore creates a catalog.Store from the underlying storage.
func newCatalogStore(s storage.Store) catalog.Store {
	return catalogpg.New(s.CollectionStore(), s.GuildStore())
}
