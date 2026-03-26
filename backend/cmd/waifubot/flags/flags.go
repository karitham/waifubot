package flags

import (
	"time"

	"github.com/urfave/cli/v2"
)

var (
	// DbURLFlag is the database URL flag
	DbURLFlag = &cli.StringFlag{
		Name:    "db-url",
		EnvVars: []string{"DB_STR", "DB_URL"},
	}

	// UserFlag is the user ID flag
	UserFlag = &cli.StringFlag{
		Name:     "user",
		EnvVars:  []string{"USER_ID"},
		Required: true,
	}

	// GuildIDFlag is the guild ID flag
	GuildIDFlag = &cli.Uint64Flag{
		Name:     "guild-id",
		EnvVars:  []string{"GUILD_ID"},
		Required: true,
	}

	// AppIDFlag is the Discord app ID flag
	AppIDFlag = &cli.StringFlag{
		Name:     "app-id",
		EnvVars:  []string{"DISCORD_APP_ID", "APP_ID"},
		Required: true,
	}

	// CharIDFlag is the character ID flag
	CharIDFlag = &cli.Int64Flag{
		Name:     "id",
		Usage:    "Character ID",
		Required: true,
	}

	// BotTokenFlag is the Discord bot token flag
	BotTokenFlag = &cli.StringFlag{
		Name:    "bot-token",
		EnvVars: []string{"DISCORD_TOKEN", "BOT_TOKEN"},
	}

	// RollCooldownFlag is the roll cooldown duration flag
	RollCooldownFlag = &cli.DurationFlag{
		Name:    "roll-cooldown",
		EnvVars: []string{"ROLL_TIMEOUT", "ROLL_COOLDOWN"},
		Value:   time.Hour * 2,
	}

	// TokensNeededFlag is the tokens needed for a wish flag
	TokensNeededFlag = &cli.IntFlag{
		Name:    "tokens-needed",
		EnvVars: []string{"TOKENS_NEEDED"},
		Value:   3,
	}

	// SeriesRollCostFlag is the token cost for a series roll
	SeriesRollCostFlag = &cli.IntFlag{
		Name:    "series-roll-cost",
		EnvVars: []string{"SERIES_ROLL_COST"},
		Value:   20,
	}

	// AnilistMaxCharsFlag is the maximum characters for Anilist queries
	AnilistMaxCharsFlag = &cli.Int64Flag{
		Name:  "anilist-max-chars",
		Value: 30000,
	}

	// NameFlag is the name flag
	NameFlag = &cli.StringFlag{
		Name:     "name",
		Required: true,
	}

	// LogLevelFlag is the log level flag
	LogLevelFlag = &cli.StringFlag{
		Name:    "log-level",
		Aliases: []string{"l"},
		EnvVars: []string{"LOG_LEVEL"},
		Value:   "INFO",
	}

	// ApiFlag is the API server flag
	ApiFlag = &cli.BoolFlag{
		Name:  "api",
		Usage: "Enable REST API server (default: true)",
		Value: true,
	}
)
