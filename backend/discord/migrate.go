package discord

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/Karitham/corde"
	"github.com/jackc/pgx/v5"
)

func (b *Bot) MigrateCommands(ctx context.Context) error {
	hash := Hash(commandDefinitions)

	stored, err := b.Store.CommandStore().GetCommandHash(ctx)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("get command hash: %w", err)
	}

	if stored == hash {
		slog.Debug("command hash unchanged, skipping migration")
		return nil
	}

	slog.Info("command hash changed, migrating", "old", stored, "new", hash)

	var opts []func(*corde.CommandsOpt)
	if b.GuildID != nil {
		opts = append(opts, corde.GuildOpt(*b.GuildID))
	}

	tmpMux := corde.NewMux(b.PublicKey, b.AppID, b.BotToken)
	if err := tmpMux.BulkRegisterCommand(ToCorde(commandDefinitions), opts...); err != nil {
		return fmt.Errorf("register commands: %w", err)
	}

	if stored == "" {
		if err := b.Store.CommandStore().SetCommandHash(ctx, hash); err != nil {
			return fmt.Errorf("set command hash: %w", err)
		}
	} else {
		if err := b.Store.CommandStore().UpdateCommandHash(ctx, hash); err != nil {
			return fmt.Errorf("update command hash: %w", err)
		}
	}

	slog.Info("command migration complete")
	return nil
}

func (b *Bot) MustMigrateCommands() {
	if err := b.MigrateCommands(context.Background()); err != nil {
		slog.Error("command migration failed", "error", err)
		os.Exit(1)
	}
}
