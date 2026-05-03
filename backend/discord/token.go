package discord

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/Karitham/corde"

	"github.com/karitham/waifubot/catalog"
	"github.com/karitham/waifubot/collection"
)

// TokenHandler handles the /token command and its subcommands.
type TokenHandler struct {
	store        collection.Store
	animeService TrackingService
	rollService  *collection.RollService
	config       collection.Config
}

// Register wires the token sub-routes on the mux.
func (h *TokenHandler) Register(m *corde.Mux) {
	m.SlashCommand("balance", trace(wrapCtx(h.Balance)))
	m.SlashCommand("give", wrap(wrapCtx(h.Give), trace[corde.SlashCommandInteractionData]))
	m.Route("sell", func(m *corde.Mux) {
		m.SlashCommand("", wrap(
			wrapCtx(h.Sell),
			trace[corde.SlashCommandInteractionData],
		))
		m.Autocomplete("id", h.userCollectionAutocomplete)
	})
	m.Route("roll", func(m *corde.Mux) {
		m.SlashCommand("", wrap(
			wrapCtx(h.Roll),
			trace[corde.SlashCommandInteractionData],
		))
		m.Autocomplete("series", h.SeriesAutocomplete)
	})
}

// Balance shows the user's token balance.
func (h *TokenHandler) Balance(ctx context.Context, w corde.ResponseWriter, cmd CommandContext) {
	user, err := h.store.GetUser(ctx, cmd.UserID())
	if err != nil {
		w.Respond(rspErr("Failed to get your balance"))
		return
	}
	w.Respond(corde.NewResp().Contentf("You have %d tokens", user.Tokens).Ephemeral())
}

// tokenGiveOptions holds the parsed options for the token give command.
type tokenGiveOptions struct {
	recipientID uint64
	recipient   corde.User
	amount      int32
}

func parseTokenGiveOptions(cmd CommandContext) (tokenGiveOptions, error) {
	user, err := cmd.OptUser("user")
	if err != nil {
		return tokenGiveOptions{}, fmt.Errorf("select a user to give tokens to: %w", err)
	}
	amount, err := cmd.OptInt("amount")
	if err != nil {
		return tokenGiveOptions{}, fmt.Errorf("specify a valid amount of tokens: %w", err)
	}
	return tokenGiveOptions{
		recipientID: uint64(user.ID),
		recipient:   user,
		amount:      int32(amount),
	}, nil
}

// Give transfers tokens to another user.
func (h *TokenHandler) Give(ctx context.Context, w corde.ResponseWriter, cmd CommandContext) {
	logger := slog.With("user_id", cmd.UserID(), "guild_id", cmd.GuildID())

	opts, err := parseTokenGiveOptions(cmd)
	if err != nil {
		logger.Debug("give command: no user selected")
		w.Respond(rspErr(err.Error()))
		return
	}

	err = collection.TransferTokens(ctx, h.store, cmd.UserID(), opts.recipientID, opts.amount)
	if err != nil {
		if errors.Is(err, collection.ErrInsufficientTokens) {
			w.Respond(rspErr("You don't have enough tokens"))
			return
		}
		if errors.Is(err, collection.ErrInvalidAmount) {
			w.Respond(rspErr("Amount must be positive"))
			return
		}
		if errors.Is(err, collection.ErrSameUserTransfer) {
			w.Respond(rspErr("You cannot transfer tokens to yourself"))
			return
		}

		logger.Error("error performing token transfer", "error", err, "recipient_id", opts.recipientID, "amount", opts.amount)
		w.Respond(rspErr("Failed to transfer tokens"))
		return
	}

	w.Respond(corde.NewResp().Contentf("Gave %d tokens to %s", opts.amount, opts.recipient.Username))
}

// tokenSellOptions holds the parsed options for the token sell command.
type tokenSellOptions struct {
	charID int
}

func parseTokenSellOptions(cmd CommandContext) (tokenSellOptions, error) {
	charID, err := cmd.OptInt("id")
	if err != nil {
		return tokenSellOptions{}, fmt.Errorf("select a character to sell: %w", err)
	}
	if charID == 0 {
		return tokenSellOptions{}, fmt.Errorf("invalid character ID")
	}
	return tokenSellOptions{charID: charID}, nil
}

// Sell sells a character for tokens.
func (h *TokenHandler) Sell(ctx context.Context, w corde.ResponseWriter, cmd CommandContext) {
	logger := slog.With("user_id", cmd.UserID(), "guild_id", cmd.GuildID())

	opts, err := parseTokenSellOptions(cmd)
	if err != nil {
		logger.Debug("sell command: no character selected")
		w.Respond(rspErr(err.Error()))
		return
	}

	char, err := collection.Exchange(ctx, h.store, cmd.UserID(), int64(opts.charID))
	if err != nil {
		if errors.Is(err, collection.ErrUserDoesNotOwnCharacter) {
			w.Respond(rspErr("You don't own that character"))
			return
		}
		logger.Error("error performing exchange", "error", err, "character_id", opts.charID)
		w.Respond(rspErr("Failed to sell character"))
		return
	}
	w.Respond(Privf("Sold %s for 1 token", char.Name))
}

// tokenRollOptions holds the parsed options for the token roll command.
type tokenRollOptions struct {
	mediaID int64
}

func parseTokenRollOptions(cmd CommandContext) (tokenRollOptions, error) {
	mediaID, err := cmd.OptInt64("series")
	if err != nil {
		return tokenRollOptions{}, fmt.Errorf("select a series to roll from: %w", err)
	}
	return tokenRollOptions{mediaID: mediaID}, nil
}

// Roll performs a series roll using tokens.
func (h *TokenHandler) Roll(ctx context.Context, w corde.ResponseWriter, cmd CommandContext) {
	logger := slog.With("user_id", cmd.UserID(), "guild_id", cmd.GuildID())

	opts, err := parseTokenRollOptions(cmd)
	if err != nil {
		logger.Debug("series roll command: no series selected")
		w.Respond(rspErr(err.Error()))
		return
	}

	char, err := h.rollService.SeriesRoll(ctx, cmd.UserID(), opts.mediaID, h.config.SeriesRollCost, h.animeService)
	if err != nil {
		if errors.Is(err, collection.ErrInsufficientTokens) {
			w.Respond(rspErr(fmt.Sprintf("You need %d tokens to roll for a series", h.config.SeriesRollCost)))
			return
		}
		if errors.Is(err, collection.ErrNoUnownedCharacters) {
			w.Respond(rspErr("You already own all characters from this series"))
			return
		}
		if errors.Is(err, collection.ErrMediaNotFound) {
			w.Respond(rspErr("No characters found for this series"))
			return
		}
		logger.Error("error performing series roll", "error", err, "media_id", opts.mediaID)
		w.Respond(rspErr("An error occurred, please try again later"))
		return
	}

	w.Respond(seriesRollEmbed(char, h.config.SeriesRollCost))
}

// SeriesAutocomplete provides media suggestions for the token roll command.
func (h *TokenHandler) SeriesAutocomplete(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[corde.AutocompleteInteractionData]) {
	autocomplete(ctx, w, i, "series", h.animeService.SearchMedia, formatMediaChoice)
}

// userCollectionAutocomplete provides character suggestions for the token sell command.
func (h *TokenHandler) userCollectionAutocomplete(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[corde.AutocompleteInteractionData]) {
	autocomplete(ctx, w, i, "id", func(ctx context.Context, input string) ([]catalog.Character, error) {
		return h.store.SearchCharacters(ctx, uint64(i.Member.User.ID), input)
	}, formatCharacterChoice)
}
