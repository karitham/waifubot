package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/karitham/waifubot/guild"
	"github.com/karitham/waifubot/storage"
)

var HoldersCommand = &cli.Command{
	Name:  "holders",
	Usage: "Query who owns a specific character in a guild",
	Flags: []cli.Flag{
		guildIDFlag,
		charIDFlag,
		dbURLFlag,
	},
	Action: func(c *cli.Context) error {
		guildID := c.Uint64(guildIDFlag.Name)
		charID := c.Int64(charIDFlag.Name)
		dbURL := c.String(dbURLFlag.Name)

		ctx := c.Context
		store, err := storage.NewStore(ctx, dbURL)
		if err != nil {
			return fmt.Errorf("error connecting to db: %w", err)
		}

		if guildID == 0 {
			return fmt.Errorf("invalid guild ID: %d", guildID)
		}

		collStore := newCollectionStore(store)
		charName, holderIDs, err := guild.CharacterHolders(ctx, collStore, collStore, guildID, charID)
		if err != nil {
			return fmt.Errorf("failed to get character holders: %w", err)
		}

		result := map[string]any{
			"character_id":   charID,
			"character_name": charName,
			"holder_ids":     holderIDs,
		}

		return json.NewEncoder(os.Stdout).Encode(result)
	},
}
