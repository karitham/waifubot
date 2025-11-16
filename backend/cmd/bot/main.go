package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/Karitham/corde"
	"github.com/redis/go-redis/v9"
	"github.com/urfave/cli/v2"

	"github.com/karitham/waifubot/anilist"
	"github.com/karitham/waifubot/collection"
	"github.com/karitham/waifubot/discord"
	"github.com/karitham/waifubot/guild"
	"github.com/karitham/waifubot/storage"
	"github.com/karitham/waifubot/storage/collectionstore"

	"github.com/karitham/waifubot/storage/dropstore"
	"github.com/karitham/waifubot/storage/interactionstore"
)

func parseLogLevel(levelStr string) slog.Level {
	switch strings.ToUpper(levelStr) {
	case "DEBUG":
		return slog.LevelDebug
	case "INFO":
		return slog.LevelInfo
	case "WARN", "WARNING":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func main() {
	logLevel := parseLogLevel(os.Getenv("LOG_LEVEL"))
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: logLevel})))

	disc := &discordCmd{}
	d := &dbCmd{}

	app := &cli.App{
		Name:        "waifubot",
		Usage:       "Run the bot, and use utils",
		Description: "A discord gacha bot",
		Commands: []*cli.Command{
			{
				Name:    "register",
				Aliases: []string{"r"},
				Usage:   "Register the bot commands",
				Action:  disc.register,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "bot-token",
						EnvVars:     []string{"DISCORD_TOKEN", "BOT_TOKEN"},
						Destination: &disc.botToken,
						Required:    true,
					},
					&cli.StringFlag{
						Name:        "guild-id",
						EnvVars:     []string{"DISCORD_GUILD_ID", "GUILD_ID"},
						Destination: &disc.guildID,
					},
					&cli.StringFlag{
						Name:        "app-id",
						EnvVars:     []string{"DISCORD_APP_ID", "APP_ID"},
						Destination: &disc.appID,
						Required:    true,
					},
				},
			},
			{
				Name:    "update-character",
				Usage:   "Update the character in the database",
				Aliases: []string{"uc"},
				Action:  d.update,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "db-url",
						EnvVars:     []string{"DB_URL"},
						Destination: &d.DatabaseURL,
						Required:    true,
					},
				},
			},
			{
				Name:  "index",
				Usage: "Index a guild's members",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "guild-id",
						EnvVars:  []string{"GUILD_ID"},
						Required: true,
					},
				},
				Action: disc.indexGuild,
			},
			{
				Name:  "holders",
				Usage: "Query who owns a specific character in a guild",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "guild-id",
						EnvVars:  []string{"GUILD_ID"},
						Required: true,
					},
					&cli.Int64Flag{
						Name:     "id",
						Usage:    "Character ID to query",
						Required: true,
					},
				},
				Action: disc.holders,
			},
			{
				Name:  "profile",
				Usage: "Show user profile",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "user",
						EnvVars:  []string{"USER_ID"},
						Required: true,
					},
				},
				Action: disc.profile,
			},
			{
				Name:  "exchange",
				Usage: "Exchange a character for a token",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "user",
						EnvVars:  []string{"USER_ID"},
						Required: true,
					},
					&cli.Int64Flag{
						Name:     "id",
						Usage:    "Character ID to exchange",
						Required: true,
					},
				},
				Action: disc.exchange,
			},
			{
				Name:  "search",
				Usage: "Search for anime, manga, users, or characters",
				Subcommands: []*cli.Command{
					{
						Name:  "anime",
						Usage: "Search for anime",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "name",
								Required: true,
							},
						},
						Action: disc.searchAnime,
					},
					{
						Name:  "manga",
						Usage: "Search for manga",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "name",
								Required: true,
							},
						},
						Action: disc.searchManga,
					},
					{
						Name:  "user",
						Usage: "Search for users",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "name",
								Required: true,
							},
						},
						Action: disc.searchUser,
					},
					{
						Name:  "character",
						Usage: "Search for characters",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "name",
								Required: true,
							},
						},
						Action: disc.searchCharacter,
					},
				},
			},
			{
				Name:  "verify",
				Usage: "Check if user owns a character",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "user",
						EnvVars:  []string{"USER_ID"},
						Required: true,
					},
					&cli.Int64Flag{
						Name:     "id",
						Usage:    "Character ID to check",
						Required: true,
					},
				},
				Action: disc.verify,
			},
			{
				Name:  "list",
				Usage: "List user's characters",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "user",
						EnvVars:  []string{"USER_ID"},
						Required: true,
					},
				},
				Action: disc.list,
			},
			{
				Name:  "roll",
				Usage: "Roll a character for a user",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "user",
						EnvVars:  []string{"USER_ID"},
						Required: true,
					},
				},
				Action: disc.rollForUser,
			},
			{
				Name:  "give",
				Usage: "Give a character from one user to another",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "from",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "to",
						Required: true,
					},
					&cli.Int64Flag{
						Name:     "id",
						Usage:    "Character ID to give",
						Required: true,
					},
				},
				Action: disc.giveCharacter,
			},
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "bot-token",
				EnvVars:     []string{"DISCORD_TOKEN", "BOT_TOKEN"},
				Required:    true,
				Destination: &disc.botToken,
			},
			&cli.StringFlag{
				Name:        "guild-id",
				EnvVars:     []string{"DISCORD_GUILD_ID", "GUILD_ID"},
				Destination: &disc.guildID,
			},
			&cli.StringFlag{
				Name:        "app-id",
				EnvVars:     []string{"DISCORD_APP_ID", "APP_ID"},
				Destination: &disc.appID,
				Required:    true,
			},
			&cli.StringFlag{
				Name:        "public-key",
				EnvVars:     []string{"DISCORD_PUBLIC_KEY", "PUBLIC_KEY"},
				Destination: &disc.publicKey,
			},
			&cli.DurationFlag{
				Name:        "roll-cooldown",
				EnvVars:     []string{"ROLL_TIMEOUT", "ROLL_COOLDOWN"},
				Value:       time.Hour * 2,
				Destination: &disc.rollCooldown,
			},
			&cli.Int64Flag{
				Name:        "tokens-needed",
				EnvVars:     []string{"TOKENS_NEEDED"},
				Value:       3,
				Destination: &disc.tokensNeeded,
			},
			&cli.Int64Flag{
				Name:        "interaction-needed",
				EnvVars:     []string{"INTERACTION_NEEDED"},
				Value:       25,
				Destination: &disc.interactionNeeded,
			},
			&cli.StringFlag{
				Name:        "db-url",
				EnvVars:     []string{"DB_STR", "DB_URL"},
				Destination: &disc.dbURL,
			},
			&cli.StringFlag{
				Name:        "port",
				EnvVars:     []string{"PORT"},
				Value:       "8080",
				Destination: &disc.port,
			},
			&cli.Int64Flag{
				Name:        "anilist-max-chars",
				Value:       30_000,
				Destination: &disc.anilistMaxChars,
				EnvVars:     []string{"ANILIST_MAX_CHARS"},
			},
			&cli.StringFlag{
				Name:        "redis-url",
				EnvVars:     []string{"REDIS_URL"},
				Destination: &disc.redisURL,
			},
		},
		Action: disc.run,
	}

	if err := app.Run(os.Args); err != nil {
		slog.Error("error running app", "error", err)
		os.Exit(1)
	}
}

type discordCmd struct {
	botToken          string
	appID             string
	guildID           string
	publicKey         string
	anilistMaxChars   int64
	interactionNeeded int64
	tokensNeeded      int64
	rollCooldown      time.Duration
	dbURL             string
	port              string
	redisURL          string
}

func (dc *discordCmd) register(c *cli.Context) error {
	bot := &discord.Bot{
		AppID:    corde.SnowflakeFromString(dc.appID),
		BotToken: dc.botToken,
	}

	if dc.guildID != "" {
		id := corde.SnowflakeFromString(dc.guildID)
		bot.GuildID = &id
	}

	if err := bot.RegisterCommands(); err != nil {
		return fmt.Errorf("error registering commands %v", err)
	}
	return nil
}

func (dc *discordCmd) run(c *cli.Context) error {
	bot, err := dc.createBot(c.Context)
	if err != nil {
		return fmt.Errorf("error building bot: %w", err)
	}

	slog.Info("Starting WaifuBot Discord bot", "port", dc.port, "app_id", dc.appID)
	mux := discord.New(bot)
	slog.Info("Discord bot started successfully", "port", dc.port)
	err = mux.ListenAndServe(":" + dc.port)
	if err != nil {
		slog.Error("Discord bot crashed", "error", err, "port", dc.port)
		return err
	}
	slog.Info("Discord bot shutting down", "port", dc.port)
	return nil
}

func (dc *discordCmd) createBot(ctx context.Context) (*discord.Bot, error) {
	store, err := storage.NewStore(ctx, dc.dbURL)
	if err != nil {
		return nil, fmt.Errorf("error connecting to db: %w", err)
	}

	bot := &discord.Bot{
		Store:             store,
		AnimeService:      anilist.New(anilist.MaxChar(dc.anilistMaxChars)),
		GuildIndexer:      guild.NewIndexer(store, dc.botToken),
		AppID:             corde.SnowflakeFromString(dc.appID),
		BotToken:          dc.botToken,
		PublicKey:         dc.publicKey,
		RollCooldown:      dc.rollCooldown,
		TokensNeeded:      int32(dc.tokensNeeded),
		InteractionNeeded: dc.interactionNeeded,
		InterStore:        interactionstore.NewMemStore(),
		DropStore:         dropstore.NewMemStore[collection.MediaCharacter](),
	}

	if dc.redisURL != "" {
		opts, err := redis.ParseURL(dc.redisURL)
		if err != nil {
			return nil, fmt.Errorf("error parsing redis url: %w", err)
		}

		redis := redis.NewClient(opts)

		bot.DropStore = dropstore.NewRedis[collection.MediaCharacter](redis, "channel", "char")
		bot.InterStore = interactionstore.NewRedis(redis)
	}

	if dc.guildID != "" {
		id := corde.SnowflakeFromString(dc.guildID)
		bot.GuildID = &id
	}

	return bot, nil
}

func (dc *discordCmd) indexGuild(c *cli.Context) error {
	ctx := c.Context
	bot, err := dc.createBot(ctx)
	if err != nil {
		return err
	}

	guildIDStr := c.String("guild-id")
	guildID := corde.SnowflakeFromString(guildIDStr)
	if guildID == 0 {
		return fmt.Errorf("invalid guild ID: %s", guildIDStr)
	}

	err = guild.NewIndexer(bot.Store, dc.botToken).IndexGuild(ctx, guildID)
	if err != nil {
		return fmt.Errorf("failed to index guild: %w", err)
	}

	return json.NewEncoder(os.Stdout).Encode(map[string]any{
		"guild_id": guildIDStr,
	})
}

func (dc *discordCmd) holders(c *cli.Context) error {
	ctx := c.Context
	bot, err := dc.createBot(ctx)
	if err != nil {
		return err
	}

	guildIDStr := c.String("guild-id")
	guildID := corde.SnowflakeFromString(guildIDStr)
	if guildID == 0 {
		return fmt.Errorf("invalid guild ID: %s", guildIDStr)
	}

	charID := c.Int64("id")

	charName, holderIDs, err := collection.CharacterHolders(ctx, bot.Store, guildID, charID)
	if err != nil {
		return fmt.Errorf("failed to get character holders: %w", err)
	}

	result := map[string]any{
		"character_id":   charID,
		"character_name": charName,
		"holder_ids":     holderIDs,
	}

	return json.NewEncoder(os.Stdout).Encode(result)
}

func (dc *discordCmd) rollForUser(c *cli.Context) error {
	ctx := c.Context
	bot, err := dc.createBot(ctx)
	if err != nil {
		return err
	}

	userIDStr := c.String("user")
	userID := corde.SnowflakeFromString(userIDStr)
	if userID == 0 {
		return fmt.Errorf("invalid user ID: %s", userIDStr)
	}

	config := collection.Config{
		RollCooldown: bot.RollCooldown,
		TokensNeeded: bot.TokensNeeded,
	}
	char, err := collection.Roll(ctx, bot.Store, bot.AnimeService, config, userID)
	if err != nil {
		return fmt.Errorf("error rolling: %w", err)
	}

	return json.NewEncoder(os.Stdout).Encode(char)
}

func (dc *discordCmd) giveCharacter(c *cli.Context) error {
	ctx := c.Context
	bot, err := dc.createBot(ctx)
	if err != nil {
		return err
	}

	fromStr := c.String("from")
	from := corde.SnowflakeFromString(fromStr)
	if from == 0 {
		return fmt.Errorf("invalid from user ID: %s", fromStr)
	}

	toStr := c.String("to")
	to := corde.SnowflakeFromString(toStr)
	if to == 0 {
		return fmt.Errorf("invalid to user ID: %s", toStr)
	}

	charID := c.Int64("id")

	char, err := collection.Give(ctx, bot.Store, from, to, charID)
	if err != nil {
		return fmt.Errorf("error giving: %w", err)
	}

	return json.NewEncoder(os.Stdout).Encode(char)
}

func (dc *discordCmd) profile(c *cli.Context) error {
	ctx := c.Context
	bot, err := dc.createBot(ctx)
	if err != nil {
		return err
	}

	userIDStr := c.String("user")
	userID := corde.SnowflakeFromString(userIDStr)
	if userID == 0 {
		return fmt.Errorf("invalid user ID: %s", userIDStr)
	}

	profile, err := collection.UserProfile(ctx, bot.Store, userID)
	if err != nil {
		return fmt.Errorf("error getting profile: %w", err)
	}

	return json.NewEncoder(os.Stdout).Encode(profile)
}

func (dc *discordCmd) exchange(c *cli.Context) error {
	ctx := c.Context
	bot, err := dc.createBot(ctx)
	if err != nil {
		return err
	}

	userIDStr := c.String("user")
	userID := corde.SnowflakeFromString(userIDStr)
	if userID == 0 {
		return fmt.Errorf("invalid user ID: %s", userIDStr)
	}

	charID := c.Int64("id")

	char, err := collection.Exchange(ctx, bot.Store, userID, charID)
	if err != nil {
		return fmt.Errorf("error exchanging: %w", err)
	}

	result := map[string]any{
		"exchanged_character": char,
		"message":             fmt.Sprintf("Exchanged %s for a token", char.Name),
	}

	return json.NewEncoder(os.Stdout).Encode(result)
}

func (dc *discordCmd) searchAnime(c *cli.Context) error {
	ctx := c.Context
	bot, err := dc.createBot(ctx)
	if err != nil {
		return err
	}

	name := c.String("name")
	anime, err := bot.AnimeService.Anime(ctx, name)
	if err != nil {
		return fmt.Errorf("error searching anime: %w", err)
	}

	return json.NewEncoder(os.Stdout).Encode(anime)
}

func (dc *discordCmd) searchManga(c *cli.Context) error {
	ctx := c.Context
	bot, err := dc.createBot(ctx)
	if err != nil {
		return err
	}

	name := c.String("name")
	manga, err := bot.AnimeService.Manga(ctx, name)
	if err != nil {
		return fmt.Errorf("error searching manga: %w", err)
	}

	return json.NewEncoder(os.Stdout).Encode(manga)
}

func (dc *discordCmd) searchUser(c *cli.Context) error {
	ctx := c.Context
	bot, err := dc.createBot(ctx)
	if err != nil {
		return err
	}

	name := c.String("name")
	users, err := bot.AnimeService.User(ctx, name)
	if err != nil {
		return fmt.Errorf("error searching users: %w", err)
	}

	return json.NewEncoder(os.Stdout).Encode(users)
}

func (dc *discordCmd) searchCharacter(c *cli.Context) error {
	ctx := c.Context
	bot, err := dc.createBot(ctx)
	if err != nil {
		return err
	}

	name := c.String("name")
	characters, err := bot.AnimeService.Character(ctx, name)
	if err != nil {
		return fmt.Errorf("error searching characters: %w", err)
	}

	return json.NewEncoder(os.Stdout).Encode(characters)
}

func (dc *discordCmd) verify(c *cli.Context) error {
	ctx := c.Context
	bot, err := dc.createBot(ctx)
	if err != nil {
		return err
	}

	userIDStr := c.String("user")
	userID := corde.SnowflakeFromString(userIDStr)
	if userID == 0 {
		return fmt.Errorf("invalid user ID: %s", userIDStr)
	}

	charID := c.Int64("id")

	has, char, err := collection.CheckOwnership(ctx, bot.Store, userID, charID)
	if err != nil {
		return fmt.Errorf("error checking ownership: %w", err)
	}

	result := map[string]any{
		"has": has,
	}
	if has {
		result["character"] = char
	}

	return json.NewEncoder(os.Stdout).Encode(result)
}

func (dc *discordCmd) list(c *cli.Context) error {
	ctx := c.Context
	bot, err := dc.createBot(ctx)
	if err != nil {
		return err
	}

	userIDStr := c.String("user")
	userID := corde.SnowflakeFromString(userIDStr)
	if userID == 0 {
		return fmt.Errorf("invalid user ID: %s", userIDStr)
	}

	characters, err := collection.Characters(ctx, bot.Store, userID)
	if err != nil {
		return fmt.Errorf("error listing characters: %w", err)
	}

	return json.NewEncoder(os.Stdout).Encode(characters)
}

type dbCmd struct {
	DatabaseURL string
}

func (r *dbCmd) update(c *cli.Context) error {
	a := c.Args()
	if a.Len() < 1 {
		return fmt.Errorf("no character name provided")
	}

	s, err := storage.NewStore(c.Context, r.DatabaseURL)
	if err != nil {
		return fmt.Errorf("error connecting to db %v", err)
	}

	char, err := anilist.New(anilist.NoCache).Character(c.Context, c.Args().First())
	if err != nil {
		return err
	}
	if len(char) < 1 {
		return fmt.Errorf("character not found")
	}

	if _, err := s.CollectionStore().UpdateImageName(c.Context, collectionstore.UpdateImageNameParams{
		Image: char[0].ImageURL,
		Name:  strings.Join(strings.Fields(char[0].Name), " "),
		ID:    char[0].ID,
	}); err != nil {
		return fmt.Errorf("error updating db %v", err)
	}

	return nil
}
