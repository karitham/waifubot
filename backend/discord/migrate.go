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

	existing, err := b.fetchExistingCommands(ctx)
	if err != nil {
		return fmt.Errorf("fetch existing commands: %w", err)
	}

	for _, cmd := range existing {
		if err := b.deleteCommand(ctx, cmd.ID); err != nil {
			return fmt.Errorf("delete command %s: %w", cmd.Name, err)
		}
	}

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

type discordCommand struct {
	ID   corde.Snowflake
	Name string
}

func (b *Bot) fetchExistingCommands(ctx context.Context) ([]discordCommand, error) {
	var opts []func(*corde.CommandsOpt)
	if b.GuildID != nil {
		opts = append(opts, corde.GuildOpt(*b.GuildID))
	}

	commands, err := corde.NewMux(b.PublicKey, b.AppID, b.BotToken).GetCommands(opts...)
	if err != nil {
		return nil, err
	}

	result := make([]discordCommand, len(commands))
	for i, cmd := range commands {
		result[i] = discordCommand{ID: cmd.ID, Name: cmd.Name}
	}
	return result, nil
}

func (b *Bot) deleteCommand(ctx context.Context, id corde.Snowflake) error {
	var opts []func(*corde.CommandsOpt)
	if b.GuildID != nil {
		opts = append(opts, corde.GuildOpt(*b.GuildID))
	}
	return corde.NewMux(b.PublicKey, b.AppID, b.BotToken).DeleteCommand(id, opts...)
}

func (b *Bot) MustMigrateCommands() {
	if err := b.MigrateCommands(context.Background()); err != nil {
		slog.Error("command migration failed", "error", err)
		os.Exit(1)
	}
}
