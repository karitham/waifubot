package discord

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Karitham/corde"
)

// migrateCommands registers the command definitions when their hash changed
// and persists the new hash.
func migrateCommands(ctx context.Context, mux *corde.Mux, cfg Config, store CommandStore) error {
	hash := Hash(commandDefinitions)

	stored, err := store.GetCommandHash(ctx)
	if err != nil {
		return fmt.Errorf("get command hash: %w", err)
	}

	if stored == hash {
		slog.Debug("command hash unchanged, skipping migration")
		return nil
	}

	slog.Info("command hash changed, migrating", "old", stored, "new", hash)

	var opts []func(*corde.CommandsOpt)
	if cfg.GuildID != nil {
		opts = append(opts, corde.GuildOpt(*cfg.GuildID))
	}

	if err := mux.BulkRegisterCommand(ToCorde(commandDefinitions), opts...); err != nil {
		return fmt.Errorf("register commands: %w", err)
	}

	if stored == "" {
		if err := store.SetCommandHash(ctx, hash); err != nil {
			return fmt.Errorf("set command hash: %w", err)
		}
	} else {
		if err := store.UpdateCommandHash(ctx, hash); err != nil {
			return fmt.Errorf("update command hash: %w", err)
		}
	}

	slog.Info("command migration complete")
	return nil
}
