package discord

import (
	"context"

	"github.com/Karitham/corde"

	"github.com/karitham/waifubot/collection"
	"github.com/karitham/waifubot/guild"
)

// VerifyHandler handles the /verify command and its autocomplete.
type VerifyHandler struct {
	store        collection.Store
	guildIndexer *guild.Indexer
	guildTxFn    func(context.Context) (guild.TxQuerier, error)
}

// verifyOptions holds the parsed options for the verify command.
type verifyOptions struct {
	targetUserID   uint64
	targetUsername string
	charID         int64
}

func parseVerifyOptions(cmd CommandContext) verifyOptions {
	opts := verifyOptions{
		targetUserID:   cmd.UserID(),
		targetUsername: cmd.Username(),
	}
	if user, ok := cmd.FirstResolvedUser(); ok {
		opts.targetUserID = uint64(user.ID)
		opts.targetUsername = user.Username
	}
	opts.charID, _ = cmd.OptInt64("id")
	return opts
}

// Register wires the verify sub-routes on the mux.
func (h *VerifyHandler) Register(m *corde.Mux) {
	m.SlashCommand("", wrap(
		wrapCtx(h.Verify),
		trace[corde.SlashCommandInteractionData],
		indexMiddleware[corde.SlashCommandInteractionData](h.guildIndexer, h.guildTxFn),
	))
	m.Autocomplete("id", h.Autocomplete)
}

// Verify checks if a user owns a character.
func (h *VerifyHandler) Verify(ctx context.Context, w corde.ResponseWriter, cmd CommandContext) {
	opts := parseVerifyOptions(cmd)

	has, char, err := collectionCheckOwnership(ctx, h.store, opts.targetUserID, opts.charID)
	if err != nil {
		w.Respond(newErrf("Error checking ownership: %s", err.Error()))
		return
	}

	if has {
		w.Respond(newErrf("%s has %s in their collection", opts.targetUsername, char.Name))
		return
	}

	w.Respond(newErrf("%s doesn't have this character", opts.targetUsername))
}

// Autocomplete provides character suggestions for the verify command.
func (h *VerifyHandler) Autocomplete(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[corde.AutocompleteInteractionData]) {
	autocomplete(ctx, w, i, "id", h.store.SearchGlobalCharacters, formatCharacterChoice)
}
