package main

import (
	"time"

	"github.com/urfave/cli/v2"
)

var (
	dbURLFlag = &cli.StringFlag{
		Name:    "db-url",
		EnvVars: []string{"DB_STR", "DB_URL"},
	}

	userFlag = &cli.StringFlag{
		Name:     "user",
		EnvVars:  []string{"USER_ID"},
		Required: true,
	}

	guildIDFlag = &cli.Uint64Flag{
		Name:     "guild-id",
		EnvVars:  []string{"GUILD_ID"},
		Required: true,
	}

	appIDFlag = &cli.StringFlag{
		Name:     "app-id",
		EnvVars:  []string{"DISCORD_APP_ID", "APP_ID"},
		Required: true,
	}

	charIDFlag = &cli.Int64Flag{
		Name:     "id",
		Usage:    "Character ID",
		Required: true,
	}

	botTokenFlag = &cli.StringFlag{
		Name:    "bot-token",
		EnvVars: []string{"DISCORD_TOKEN", "BOT_TOKEN"},
	}

	rollCooldownFlag = &cli.DurationFlag{
		Name:    "roll-cooldown",
		EnvVars: []string{"ROLL_TIMEOUT", "ROLL_COOLDOWN"},
		Value:   time.Hour * 2,
	}

	tokensNeededFlag = &cli.IntFlag{
		Name:    "tokens-needed",
		EnvVars: []string{"TOKENS_NEEDED"},
		Value:   3,
	}

	anilistMaxCharsFlag = &cli.Int64Flag{
		Name:  "anilist-max-chars",
		Value: 30000,
	}

	nameFlag = &cli.StringFlag{
		Name:     "name",
		Required: true,
	}
)
