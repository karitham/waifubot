package discord

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/Karitham/corde"
)

func (r *Router) MigrateCommands(ctx context.Context) error {
	hash := Hash(commandDefinitions)

	stored, err := r.CommandStore.GetCommandHash(ctx)
	if err != nil {
		return fmt.Errorf("get command hash: %w", err)
	}

	if stored == hash {
		slog.Debug("command hash unchanged, skipping migration")
		return nil
	}

	slog.Info("command hash changed, migrating", "old", stored, "new", hash)

	var opts []func(*corde.CommandsOpt)
	if r.GuildID != nil {
		opts = append(opts, corde.GuildOpt(*r.GuildID))
	}

	tmpMux := corde.NewMux(r.PublicKey, r.AppID, r.BotToken)
	if err := tmpMux.BulkRegisterCommand(ToCorde(commandDefinitions), opts...); err != nil {
		return fmt.Errorf("register commands: %w", err)
	}

	if stored == "" {
		if err := r.CommandStore.SetCommandHash(ctx, hash); err != nil {
			return fmt.Errorf("set command hash: %w", err)
		}
	} else {
		if err := r.CommandStore.UpdateCommandHash(ctx, hash); err != nil {
			return fmt.Errorf("update command hash: %w", err)
		}
	}

	slog.Info("command migration complete")
	return nil
}

func (r *Router) MustMigrateCommands() {
	if err := r.MigrateCommands(context.Background()); err != nil {
		slog.Error("command migration failed", "error", err)
		os.Exit(1)
	}
}
