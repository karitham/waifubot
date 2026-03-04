package main

import (
	"log/slog"

	"github.com/urfave/cli/v2"

	"github.com/karitham/waifubot/storage"
)

var MigrateCommand = &cli.Command{
	Name:  "migrate",
	Usage: "Run database migrations",
	Flags: []cli.Flag{
		dbURLFlag,
	},
	Action: func(c *cli.Context) error {
		slog.Info("Running database migrations")
		if err := storage.Migrate(c.String(dbURLFlag.Name)); err != nil {
			return err
		}
		slog.Info("Migrations complete")
		return nil
	},
}
