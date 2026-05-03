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
	Usage: "Backfill the local character catalog from AniList",
	Description: `Pagination all characters from AniList sorted by favorites and upserts them
into the local characters table. This is idempotent — safe to re-run.

The backfill respects AniList rate limits (5 req/min) and includes a deletion
sweep to detect characters removed from AniList.`,
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
		svc := sync.NewService(catalogStore, anilistClient)

		slog.Info("starting character backfill")
		if err := svc.Backfill(ctx, 50); err != nil {
			return fmt.Errorf("backfill failed: %w", err)
		}
		slog.Info("character backfill complete")
		return nil
	},
}
