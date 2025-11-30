package main

import (
	"fmt"
	"log/slog"

	"github.com/Karitham/corde"
	"github.com/redis/go-redis/v9"
	"github.com/urfave/cli/v2"

	"github.com/karitham/waifubot/anilist"
	"github.com/karitham/waifubot/collection"
	"github.com/karitham/waifubot/discord"
	"github.com/karitham/waifubot/guild"
	"github.com/karitham/waifubot/storage"
	"github.com/karitham/waifubot/storage/dropstore"
	"github.com/karitham/waifubot/storage/interactionstore"
	"github.com/karitham/waifubot/wishlist"
)

var RunCommand = &cli.Command{
	Name:  "run",
	Usage: "Run the Discord bot",
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
		&cli.StringFlag{
			Name:    "redis-url",
			EnvVars: []string{"REDIS_URL"},
		},
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
	},
	Action: func(c *cli.Context) error {
		ctx := c.Context
		store, err := storage.NewStore(ctx, c.String(dbURLFlag.Name))
		if err != nil {
			return fmt.Errorf("error connecting to db: %w", err)
		}

		var interStore interactionstore.Store = interactionstore.NewMemStore()
		var dropStore dropstore.Store[collection.MediaCharacter] = dropstore.NewMemStore[collection.MediaCharacter]()

		if redisURL := c.String("redis-url"); redisURL != "" {
			opts, err := redis.ParseURL(redisURL)
			if err != nil {
				return fmt.Errorf("error parsing redis url: %w", err)
			}

			redis := redis.NewClient(opts)

			dropStore = dropstore.NewRedis[collection.MediaCharacter](redis, "channel", "char")
			interStore = interactionstore.NewRedis(redis)
		}

		var guildID *corde.Snowflake
		if gid := c.Uint64(guildIDFlag.Name); gid != 0 {
			id := corde.Snowflake(gid)
			guildID = &id
		}

		slog.Info("Starting WaifuBot Discord bot", "port", c.String("port"), "app_id", c.String("app-id"))
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

		slog.Info("Discord bot started successfully", "port", c.String("port"))

		err = mux.ListenAndServe(":" + c.String("port"))
		if err != nil {
			slog.Error("Discord bot crashed", "error", err, "port", c.String("port"))
			return err
		}

		slog.Info("Discord bot shutting down", "port", c.String("port"))
		return nil
	},
}
