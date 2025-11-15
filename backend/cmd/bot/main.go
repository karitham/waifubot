package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/Karitham/corde"
	"github.com/go-redis/redis/v8"
	"github.com/urfave/cli/v2"

	"github.com/karitham/waifubot/anilist"
	"github.com/karitham/waifubot/db"
	"github.com/karitham/waifubot/db/characters"
	"github.com/karitham/waifubot/discord"
	"github.com/karitham/waifubot/memstore"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})))

	disc := &discordCmd{}
	d := &dbCmd{}
	dev := false

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
				Action: disc.queryHolders,
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
			&cli.BoolFlag{
				Name:        "dev",
				EnvVars:     []string{"DEV"},
				Destination: &dev,
			},
			&cli.StringFlag{
				Name:        "redis-url",
				EnvVars:     []string{"REDIS_URL"},
				Required:    true,
				Destination: &disc.redisURL,
			},
		},
		Action: disc.run,
		Before: func(*cli.Context) error {
			if dev {
				slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})))
			}
			return nil
		},
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

	return discord.New(bot).ListenAndServe(":" + dc.port)
}

func (dc *discordCmd) createBot(ctx context.Context) (*discord.Bot, error) {
	store, err := db.NewStore(ctx, dc.dbURL)
	if err != nil {
		return nil, fmt.Errorf("error connecting to db: %w", err)
	}

	opts, err := redis.ParseURL(dc.redisURL)
	if err != nil {
		return nil, fmt.Errorf("error parsing redis url: %w", err)
	}

	redis := memstore.New(opts)

	bot := &discord.Bot{
		Store:             store,
		AnimeService:      anilist.New(anilist.MaxChar(dc.anilistMaxChars)),
		AppID:             corde.SnowflakeFromString(dc.appID),
		BotToken:          dc.botToken,
		PublicKey:         dc.publicKey,
		RollCooldown:      dc.rollCooldown,
		TokensNeeded:      int32(dc.tokensNeeded),
		InteractionNeeded: dc.interactionNeeded,
		Inter:             redis,
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

	fmt.Printf("Indexing guild %s...\n", guildIDStr)

	memberIDs, err := discord.FetchGuildMemberIDs(ctx, dc.botToken, guildID)
	if err != nil {
		return fmt.Errorf("failed to fetch members: %w", err)
	}

	err = bot.Store.InsertGuildMembers(ctx, guildID, memberIDs)
	if err != nil {
		return fmt.Errorf("failed to insert members: %w", err)
	}

	fmt.Printf("Indexed %d members\n", len(memberIDs))
	return nil
}

func (dc *discordCmd) queryHolders(c *cli.Context) error {
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

	charID := c.Int64("character-id")

	char, err := bot.Store.GetCharByID(ctx, charID)
	if err != nil {
		return fmt.Errorf("character not found: %w", err)
	}

	memberIDs, err := bot.Store.GetGuildMembers(ctx, guildID)
	if err != nil {
		return fmt.Errorf("failed to fetch guild members: %w", err)
	}

	if len(memberIDs) == 0 {
		fmt.Println("Guild members not indexed yet")
		return nil
	}

	holderIDs, err := bot.Store.UsersOwningCharInGuild(ctx, charID, guildID)
	if err != nil {
		return fmt.Errorf("failed to fetch character holders: %w", err)
	}

	if len(holderIDs) == 0 {
		fmt.Printf("No one in this server has %s (ID: %d)\n", char.Name, charID)
		return nil
	}

	fmt.Printf("Users in this server who have %s (ID: %d):\n", char.Name, charID)
	for _, holderID := range holderIDs {
		fmt.Printf("- %d\n", holderID)
	}

	return nil
}

func (dc *discordCmd) rollForUser(c *cli.Context) error {
	ctx := c.Context
	bot, err := dc.createBot(ctx)
	if err != nil {
		return err
	}

	userIDStr := c.String("user-id")
	userID := corde.SnowflakeFromString(userIDStr)
	if userID == 0 {
		return fmt.Errorf("invalid user ID: %s", userIDStr)
	}

	char, err := bot.PerformRoll(ctx, userID)
	if err != nil {
		return fmt.Errorf("error rolling: %w", err)
	}

	fmt.Printf("Rolled %s (ID: %d) from %s\n", char.Name, char.ID, char.MediaTitle)
	return nil
}

func (dc *discordCmd) giveCharacter(c *cli.Context) error {
	ctx := c.Context
	bot, err := dc.createBot(ctx)
	if err != nil {
		return err
	}

	fromStr := c.String("from-user-id")
	from := corde.SnowflakeFromString(fromStr)
	if from == 0 {
		return fmt.Errorf("invalid from user ID: %s", fromStr)
	}

	toStr := c.String("to-user-id")
	to := corde.SnowflakeFromString(toStr)
	if to == 0 {
		return fmt.Errorf("invalid to user ID: %s", toStr)
	}

	charID := c.Int64("character-id")

	char, err := bot.PerformGive(ctx, from, to, charID)
	if err != nil {
		return fmt.Errorf("error giving: %w", err)
	}

	fmt.Printf("Gave %s (%d) from %d to %d\n", char.Name, charID, from, to)
	return nil
}

type dbCmd struct {
	DatabaseURL string
}

func (r *dbCmd) update(c *cli.Context) error {
	a := c.Args()
	if a.Len() < 1 {
		return fmt.Errorf("no character name provided")
	}

	s, err := db.NewStore(c.Context, r.DatabaseURL)
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

	if _, err := s.CharacterStore.UpdateImageName(c.Context, characters.UpdateImageNameParams{
		Image: char[0].ImageURL,
		Name:  strings.Join(strings.Fields(char[0].Name), " "),
		ID:    char[0].ID,
	}); err != nil {
		return fmt.Errorf("error updating db %v", err)
	}

	return nil
}
