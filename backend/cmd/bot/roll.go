package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Karitham/corde"
	"github.com/urfave/cli/v2"

	"github.com/karitham/waifubot/anilist"
	"github.com/karitham/waifubot/collection"
	"github.com/karitham/waifubot/storage"
)

var RollCommand = &cli.Command{
	Name:  "roll",
	Usage: "Roll a character for a user",
	Flags: []cli.Flag{
		userFlag,
		dbURLFlag,
		rollCooldownFlag,
		tokensNeededFlag,
	},
	Action: func(c *cli.Context) error {
		userIDStr := c.String(userFlag.Name)
		dbURL := c.String(dbURLFlag.Name)
		rollCooldown := c.Duration(rollCooldownFlag.Name)
		tokensNeeded := c.Int(tokensNeededFlag.Name)

		ctx := c.Context
		store, err := storage.NewStore(ctx, dbURL)
		if err != nil {
			return fmt.Errorf("error connecting to db: %w", err)
		}

		userID := corde.SnowflakeFromString(userIDStr)
		if userID == 0 {
			return fmt.Errorf("invalid user ID: %s", userIDStr)
		}

		config := collection.Config{
			RollCooldown: rollCooldown,
			TokensNeeded: int32(tokensNeeded),
		}
		animeService := anilist.New(anilist.MaxChar(30_000)) // default
		char, err := collection.Roll(ctx, store, animeService, config, userID)
		if err != nil {
			return fmt.Errorf("error rolling: %w", err)
		}

		return json.NewEncoder(os.Stdout).Encode(char)
	},
}
