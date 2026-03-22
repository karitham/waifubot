package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/urfave/cli/v2"

	"github.com/karitham/waifubot/collection"
	"github.com/karitham/waifubot/storage"
)

var ListCommand = &cli.Command{
	Name:  "list",
	Usage: "List user's characters",
	Flags: []cli.Flag{
		userFlag,
		dbURLFlag,
	},
	Action: func(c *cli.Context) error {
		userIDStr := c.String(userFlag.Name)
		dbURL := c.String(dbURLFlag.Name)

		ctx := c.Context
		store, err := storage.NewStore(ctx, dbURL)
		if err != nil {
			return fmt.Errorf("error connecting to db: %w", err)
		}

		userID, err := strconv.ParseUint(userIDStr, 10, 64)
		if err != nil || userID == 0 {
			return fmt.Errorf("invalid user ID: %s", userIDStr)
		}

		characters, err := collection.Characters(ctx, newCollectionStore(store), userID)
		if err != nil {
			return fmt.Errorf("error listing characters: %w", err)
		}

		return json.NewEncoder(os.Stdout).Encode(characters)
	},
}
