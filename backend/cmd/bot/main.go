package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Karitham/corde"
	"github.com/go-redis/redis/v8"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"

	"github.com/karitham/waifubot/anilist"
	"github.com/karitham/waifubot/db"
	"github.com/karitham/waifubot/discord"
	"github.com/karitham/waifubot/memstore"
)

func main() {
	log.Logger = log.Level(zerolog.DebugLevel)

	disc := &discordCmd{}
	d := &dbCmd{}
	dev := false

	app := &cli.App{
		Name:        "waifubot",
		Usage:       "Run the bot, and use utils",
		Version:     "v0.7.2",
		Description: "A discord gacha bot",
		Commands: []*cli.Command{
			{
				Name:    "register",
				Aliases: []string{"r"},
				Usage:   "Register the bot commands",
				Action:  disc.register,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "BOT_TOKEN",
						EnvVars:     []string{"DISCORD_TOKEN", "BOT_TOKEN"},
						Destination: &disc.botToken,
						Required:    true,
					},
					&cliSnowflake{
						EnvVars: []string{"DISCORD_GUILD_ID", "GUILD_ID"},
						Dest:    disc.guildID,
					},
					&cli.StringFlag{
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
						Name:        "DB_URL",
						EnvVars:     []string{"DB_URL"},
						Destination: &d.dbURL,
						Required:    true,
					},
				},
			},
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "BOT_TOKEN",
				EnvVars:     []string{"DISCORD_TOKEN", "BOT_TOKEN"},
				Required:    true,
				Destination: &disc.botToken,
			},
			&cliSnowflake{
				EnvVars: []string{"DISCORD_GUILD_ID", "GUILD_ID"},
				Dest:    disc.guildID,
			},
			&cli.StringFlag{
				EnvVars:     []string{"DISCORD_APP_ID", "APP_ID"},
				Destination: &disc.appID,
				Required:    true,
			},
			&cli.StringFlag{
				Name:        "PUBLIC_KEY",
				EnvVars:     []string{"DISCORD_PUBLIC_KEY", "PUBLIC_KEY"},
				Destination: &disc.publicKey,
			},
			&cli.DurationFlag{
				Name:        "ROLL_COOLDOWN",
				EnvVars:     []string{"ROLL_TIMEOUT", "ROLL_COOLDOWN"},
				Value:       time.Hour * 2,
				Destination: &disc.rollCooldown,
			},
			&cli.Int64Flag{
				Name:        "TOKENS_NEEDED",
				EnvVars:     []string{"TOKENS_NEEDED"},
				Value:       3,
				Destination: &disc.tokensNeeded,
			},
			&cli.Int64Flag{
				Name:        "INTERACTION_NEEDED",
				EnvVars:     []string{"INTERACTION_NEEDED"},
				Value:       25,
				Destination: &disc.interactionNeeded,
			},
			&cli.StringFlag{
				Name:        "DB_URL",
				EnvVars:     []string{"DB_STR", "DB_URL"},
				Destination: &disc.dbURL,
			},
			&cli.StringFlag{
				Name:        "PORT",
				EnvVars:     []string{"PORT"},
				Value:       "8080",
				Destination: &disc.port,
			},
			&cli.Int64Flag{
				Name:        "ANILIST_MAX_CHARS",
				Value:       30_000,
				Destination: &disc.anilistMaxChars,
				EnvVars:     []string{"ANILIST_MAX_CHARS"},
			},
			&cli.BoolFlag{
				Name:        "DEV",
				EnvVars:     []string{"DEV"},
				Destination: &dev,
			},
			&cli.StringFlag{
				Name:        "REDIS_URL",
				EnvVars:     []string{"REDIS_URL"},
				Required:    true,
				Destination: &disc.redisURL,
			},
		},
		Action: disc.run,
		Before: func(*cli.Context) error {
			if dev {
				log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
				log.Logger = log.Level(zerolog.TraceLevel)
			}
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal().Err(err).Msg("error running app")
	}
}

type discordCmd struct {
	botToken          string
	appID             string
	guildID           *corde.Snowflake
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
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	bot := &discord.Bot{
		AppID:    corde.SnowflakeFromString(dc.appID),
		BotToken: dc.botToken,
		GuildID:  dc.guildID,
	}

	if err := bot.RegisterCommands(); err != nil {
		return fmt.Errorf("error registering commands %v", err)
	}
	return nil
}

func (dc *discordCmd) run(c *cli.Context) error {
	db, err := db.NewDB(c.Context, dc.dbURL)
	if err != nil {
		return fmt.Errorf("error connecting to db %v", err)
	}

	opts, err := redis.ParseURL(dc.redisURL)
	if err != nil {
		return fmt.Errorf("error parsing redis url %v", err)
	}

	redis := memstore.New(opts)
	defer redis.Close()

	disc := discord.New(&discord.Bot{
		Store:             db,
		AnimeService:      anilist.New(anilist.MaxChar(dc.anilistMaxChars)),
		AppID:             corde.SnowflakeFromString(dc.appID),
		GuildID:           dc.guildID,
		BotToken:          dc.botToken,
		PublicKey:         dc.publicKey,
		RollCooldown:      dc.rollCooldown,
		TokensNeeded:      int32(dc.tokensNeeded),
		InteractionNeeded: dc.interactionNeeded,
		Inter:             redis,
	})

	return disc.ListenAndServe(":" + dc.port)
}

type dbCmd struct {
	dbURL string
}

func (r *dbCmd) update(c *cli.Context) error {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	a := c.Args()
	if a.Len() < 1 {
		return fmt.Errorf("no character name provided")
	}

	DB, err := db.NewDB(c.Context, r.dbURL)
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

	if _, err := DB.SetChar(c.Context, db.SetCharParams{
		Image: char[0].ImageURL,
		Name:  strings.Join(strings.Fields(char[0].Name), " "),
		ID:    char[0].ID,
	}); err != nil {
		return fmt.Errorf("error updating db %v", err)
	}

	return nil
}
