package main

import (
	"fmt"
	"log/slog"

	"github.com/urfave/cli/v2"

	"github.com/karitham/waifubot/anilist"
	"github.com/karitham/waifubot/storage"
	"github.com/karitham/waifubot/sync"
)

var BackfillCommand = &cli.Command{
	Name:  "backfill",
	Usage: "Sample random characters from AniList into the local catalog",
	Description: `Generates random character IDs, batch-fetches them from AniList,
and upserts valid results into the local characters table.

The process respects AniList rate limits (5 req/min) and converges on the
full character set over time.`,
	Flags: []cli.Flag{
		dbURLFlag,
	},
	Action: func(c *cli.Context) error {
		ctx := c.Context
		dbURL := c.String(dbURLFlag.Name)

		store, err := storage.NewStore(ctx, dbURL)
		if err != nil {
			return fmt.Errorf("error connecting to db: %w", err)
		}

		anilistClient := anilist.New()

		catalogStore := newCatalogStore(store)
		svc := sync.NewService(catalogStore, anilistClient, anilistClient)

		slog.Info("starting character sync")
		if err := svc.Backfill(ctx, 0); err != nil {
			return fmt.Errorf("sync failed: %w", err)
		}
		slog.Info("character sync complete")
		return nil
	},
}
