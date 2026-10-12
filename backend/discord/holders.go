package discord

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/Karitham/corde"

	"github.com/karitham/waifubot/catalog"
	"github.com/karitham/waifubot/guild"
)

// indexMiddleware triggers guild indexing in a background goroutine.
func indexMiddleware[T corde.InteractionDataConstraint](
	guildIndexer *guild.Indexer,
) func(func(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[T])) func(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[T]) {
	return func(next func(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[T])) func(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[T]) {
		return func(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[T]) {
			go func() {
				if err := guildIndexer.IndexGuildIfNeeded(context.Background(), i.GuildID); err != nil {
					slog.Error("failed to index guild", "error", err, "guild_id", i.GuildID)
				}
			}()

			next(ctx, w, i)
		}
	}
}

// HoldersHandler handles the /holders command and its autocomplete.
type HoldersHandler struct {
	guildOps     guild.Querier
	catalog      catalog.Store
	guildIndexer *guild.Indexer
}

// Register wires the holders sub-routes on the mux.
func (h *HoldersHandler) Register(m *corde.Mux) {
	m.SlashCommand("", wrap(
		wrapCtx(h.Holders),
		indexMiddleware[corde.SlashCommandInteractionData](h.guildIndexer),
		trace[corde.SlashCommandInteractionData],
	))
	m.Autocomplete("id", h.Autocomplete)
}

// holdersOptions holds the parsed options for the holders command.
type holdersOptions struct {
	charID catalog.CharacterID
}

func parseHoldersOptions(cmd CommandContext) (holdersOptions, error) {
	charID, err := cmd.OptInt("id")
	if err != nil {
		return holdersOptions{}, fmt.Errorf("select a character to check: %w", err)
	}
	return holdersOptions{charID: catalog.CharacterID(charID)}, nil
}

// Holders shows which guild members own a character.
func (h *HoldersHandler) Holders(ctx context.Context, w corde.ResponseWriter, cmd CommandContext) {
	opts, err := parseHoldersOptions(cmd)
	if err != nil {
		w.Respond(rspErr(err.Error()))
		return
	}

	charName, holderIDs, err := guild.CharacterHolders(ctx, h.guildOps, h.catalog, cmd.GuildID(), opts.charID)
	if err != nil {
		switch {
		case errors.Is(err, guild.ErrNotInGuild):
			w.Respond(rspErr("This command can only be used in servers"))
		case errors.Is(err, guild.ErrNotIndexed):
			w.Respond(rspErr("Guild members are still being indexed, please try again in a moment"))
		case errors.Is(err, catalog.ErrNotFound):
			w.Respond(corde.NewResp().Contentf("No one in this server has (ID: %d)", opts.charID).Ephemeral())
		default:
			w.Respond(Privf("Failed to check holders"))
		}
		return
	}

	if len(holderIDs) == 0 {
		w.Respond(corde.NewResp().Contentf("No one in this server has **%s** (ID: %d)", charName, opts.charID).Ephemeral())
		return
	}

	var mentions strings.Builder

	fmt.Fprintf(&mentions, "Users in this server who have **%s** (ID: %d):\n", charName, opts.charID)
	for _, holderID := range holderIDs {
		fmt.Fprintf(&mentions, "- <@%d>\n", holderID)
	}

	w.Respond(corde.NewResp().Content(mentions.String()).Ephemeral())
}

// Autocomplete provides character suggestions for the holders command.
func (h *HoldersHandler) Autocomplete(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[corde.AutocompleteInteractionData]) {
	autocomplete(ctx, w, i, "id", h.catalog.SearchGlobalCharacters, formatCharacterChoice)
}
