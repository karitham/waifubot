package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

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
	},
	Action: func(c *cli.Context) error {
		userIDStr := c.String(userFlag.Name)
		dbURL := c.String(dbURLFlag.Name)
		rollCooldown := c.Duration(rollCooldownFlag.Name)

		ctx := c.Context
		store, err := storage.NewStore(ctx, dbURL)
		if err != nil {
			return fmt.Errorf("error connecting to db: %w", err)
		}

		userID, err := strconv.ParseUint(userIDStr, 10, 64)
		if err != nil || userID == 0 {
			return fmt.Errorf("invalid user ID: %s", userIDStr)
		}

		config := collection.Config{
			RollCooldown: rollCooldown,
		}
		animeService := anilist.New(anilist.MaxChar(30_000))
		char, err := collection.Roll(ctx, newCollectionStore(store), animeService, config, userID)
		if err != nil {
			return fmt.Errorf("error rolling: %w", err)
		}

		return json.NewEncoder(os.Stdout).Encode(char)
	},
}
