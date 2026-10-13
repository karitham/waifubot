package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/urfave/cli/v2"

	"github.com/karitham/waifubot/catalog"
	"github.com/karitham/waifubot/collection"
	"github.com/karitham/waifubot/storage"
	"github.com/karitham/waifubot/storage/pg"
)

var VerifyCommand = &cli.Command{
	Name:  "verify",
	Usage: "Check if user owns a character",
	Flags: []cli.Flag{
		userFlag,
		charIDFlag,
		dbURLFlag,
	},
	Action: func(c *cli.Context) error {
		userIDStr := c.String(userFlag.Name)
		charID := c.Int64(charIDFlag.Name)
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

		has, char, err := collection.CheckOwnership(ctx, pg.NewPool(store), collection.UserID(userID), catalog.CharacterID(charID))
		if err != nil {
			return fmt.Errorf("error checking ownership: %w", err)
		}

		result := map[string]any{
			"has": has,
		}
		if has {
			result["character"] = characterView(char)
		}

		return json.NewEncoder(os.Stdout).Encode(result)
	},
}
