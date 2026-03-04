package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Karitham/corde"
	"github.com/urfave/cli/v2"

	"github.com/karitham/waifubot/collection"
	"github.com/karitham/waifubot/storage"
)

var ExchangeCommand = &cli.Command{
	Name:  "exchange",
	Usage: "Exchange a character for a token",
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

		userID := corde.SnowflakeFromString(userIDStr)
		if userID == 0 {
			return fmt.Errorf("invalid user ID: %s", userIDStr)
		}

		char, err := collection.Exchange(ctx, store, userID, charID)
		if err != nil {
			return fmt.Errorf("error exchanging: %w", err)
		}

		result := map[string]any{
			"exchanged_character": char,
			"message":             fmt.Sprintf("Exchanged %s for a token", char.Name),
		}

		return json.NewEncoder(os.Stdout).Encode(result)
	},
}
