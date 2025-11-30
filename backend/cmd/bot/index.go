package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Karitham/corde"
	"github.com/urfave/cli/v2"

	"github.com/karitham/waifubot/guild"
	"github.com/karitham/waifubot/storage"
)

var IndexCommand = &cli.Command{
	Name:  "index",
	Usage: "Index a guild's members",
	Flags: []cli.Flag{
		guildIDFlag,
		botTokenFlag,
		dbURLFlag,
	},
	Action: func(c *cli.Context) error {
		guildID := corde.Snowflake(c.Uint64(guildIDFlag.Name))
		botToken := c.String(botTokenFlag.Name)
		dbURL := c.String(dbURLFlag.Name)

		ctx := c.Context
		store, err := storage.NewStore(ctx, dbURL)
		if err != nil {
			return fmt.Errorf("error connecting to db: %w", err)
		}

		if guildID == 0 {
			return fmt.Errorf("invalid guild ID: %d", guildID)
		}

		indexer := guild.NewIndexer(store, botToken)
		err = indexer.IndexGuild(ctx, guildID)
		if err != nil {
			return fmt.Errorf("failed to index guild: %w", err)
		}

		return json.NewEncoder(os.Stdout).Encode(map[string]any{
			"guild_id": guildID,
		})
	},
}
