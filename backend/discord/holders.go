package discord

import (
	"context"
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
	guildTxFn func(context.Context) (guild.TxQuerier, error),
) func(func(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[T])) func(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[T]) {
	return func(next func(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[T])) func(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[T]) {
		return func(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[T]) {
			go func() {
				if err := guildIndexer.IndexGuildIfNeeded(context.Background(), i.GuildID, guildTxFn); err != nil {
					slog.Error("failed to index guild", "error", err, "guild_id", i.GuildID)
				}
			}()

			next(ctx, w, i)
		}
	}
}

// HoldersHandler handles the /holders command and its autocomplete.
type HoldersHandler struct {
	guildOps     guild.GuildQuerier
	catalog      catalog.Store
	guildIndexer *guild.Indexer
	guildTxFn    func(context.Context) (guild.TxQuerier, error)
}

// Register wires the holders sub-routes on the mux.
func (h *HoldersHandler) Register(m *corde.Mux) {
	m.SlashCommand("", wrap(
		wrapCtx(h.Holders),
		indexMiddleware[corde.SlashCommandInteractionData](h.guildIndexer, h.guildTxFn),
		trace[corde.SlashCommandInteractionData],
	))
	m.Autocomplete("id", h.Autocomplete)
}

// holdersOptions holds the parsed options for the holders command.
type holdersOptions struct {
	charID int64
}

func parseHoldersOptions(cmd CommandContext) (holdersOptions, error) {
	charID, err := cmd.OptInt("id")
	if err != nil {
		return holdersOptions{}, fmt.Errorf("select a character to check: %w", err)
	}
	return holdersOptions{charID: int64(charID)}, nil
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
		w.Respond(Privf("Failed to check holders"))
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
