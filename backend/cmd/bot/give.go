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

var GiveCommand = &cli.Command{
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
		charIDFlag,
		dbURLFlag,
	},
	Action: func(c *cli.Context) error {
		fromStr := c.String("from")
		toStr := c.String("to")
		charID := c.Int64(charIDFlag.Name)
		dbURL := c.String(dbURLFlag.Name)

		ctx := c.Context
		store, err := storage.NewStore(ctx, dbURL)
		if err != nil {
			return fmt.Errorf("error connecting to db: %w", err)
		}

		from := corde.SnowflakeFromString(fromStr)
		if from == 0 {
			return fmt.Errorf("invalid from user ID: %s", fromStr)
		}

		to := corde.SnowflakeFromString(toStr)
		if to == 0 {
			return fmt.Errorf("invalid to user ID: %s", toStr)
		}

		char, err := collection.Give(ctx, store, from, to, charID)
		if err != nil {
			return fmt.Errorf("error giving: %w", err)
		}

		return json.NewEncoder(os.Stdout).Encode(char)
	},
}
