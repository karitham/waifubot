package main

import (
	"log/slog"
	"os"
	"strings"

	"github.com/urfave/cli/v2"
)

var version = "dev"

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

	app := &cli.App{
		Name:        "waifubot",
		Usage:       "Run the bot, API, and use utils",
		Description: "A discord gacha bot with REST API",
		Version:     version,
		Commands: []*cli.Command{
			RunCommand,
			MigrateCommand,
			MigrateRedisCommand,
			IndexCommand,
			HoldersCommand,
			ProfileCommand,
			ExchangeCommand,
			SearchAnimeCommand,
			SearchMangaCommand,
			VerifyCommand,
			ListCommand,
			RollCommand,
			GiveCommand,
			WishlistCommand,
			UpdateCharacterCommand,
		},
		DefaultCommand: "run",
	}

	if err := app.Run(os.Args); err != nil {
		slog.Error("error running app", "error", err)
		os.Exit(1)
	}
}
