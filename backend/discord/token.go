package discord

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/Karitham/corde"

	"github.com/karitham/waifubot/collection"
)

func (b *Bot) token(m *corde.Mux) {
	m.SlashCommand("balance", trace(b.tokenBalance))
	m.SlashCommand("give", wrap(b.tokenGive, trace[corde.SlashCommandInteractionData]))
	m.Route("sell", func(m *corde.Mux) {
		m.SlashCommand("", wrap(
			b.tokenSell,
			trace[corde.SlashCommandInteractionData],
			interact(b.InterStore, onInteraction[corde.SlashCommandInteractionData](b)),
		))
		m.Autocomplete("id", trace(b.userCollectionAutocomplete))
	})
	m.Route("roll", func(m *corde.Mux) {
		m.SlashCommand("", wrap(
			b.tokenRoll,
			trace[corde.SlashCommandInteractionData],
			interact(b.InterStore, onInteraction[corde.SlashCommandInteractionData](b)),
		))
		m.Autocomplete("series", trace(b.seriesAutocomplete))
	})
}

func (b *Bot) tokenBalance(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[corde.SlashCommandInteractionData]) {
	user, err := b.Store.GetUser(ctx, uint64(i.Member.User.ID))
	if err != nil {
		w.Respond(rspErr("Failed to get your balance"))
		return
	}
	w.Respond(corde.NewResp().Contentf("You have %d tokens", user.Tokens).Ephemeral())
}

func (b *Bot) tokenGive(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[corde.SlashCommandInteractionData]) {
	logger := slog.With("user_id", uint64(i.Member.User.ID), "guild_id", uint64(i.GuildID))

	recipient, errUser := i.Data.OptionsUser("user")
	if errUser != nil {
		logger.Debug("give command: no user selected")
		w.Respond(rspErr("select a user to give tokens to"))
		return
	}

	amount, errAmount := i.Data.Options.Int("amount")
	if errAmount != nil {
		logger.Debug("give command: invalid amount")
		w.Respond(rspErr("specify a valid amount of tokens"))
		return
	}

	err := collection.TransferTokens(ctx, b.Store, uint64(i.Member.User.ID), uint64(recipient.ID), int32(amount))
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

		logger.Error("error performing token transfer", "error", err, "recipient_id", uint64(recipient.ID), "amount", amount)
		w.Respond(rspErr("Failed to transfer tokens"))
		return
	}

	w.Respond(corde.NewResp().Contentf("Gave %d tokens to %s", amount, recipient.Username))
}

func (b *Bot) tokenSell(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[corde.SlashCommandInteractionData]) {
	logger := slog.With("user_id", uint64(i.Member.User.ID), "guild_id", uint64(i.GuildID))

	charID, errChar := i.Data.Options.Int("id")
	if errChar != nil {
		logger.Debug("sell command: no character selected")
		w.Respond(rspErr("select a character to sell"))
		return
	}

	if charID == 0 {
		logger.Debug("sell command: invalid character ID")
		w.Respond(rspErr("invalid character ID"))
		return
	}

	char, err := collection.Exchange(ctx, b.Store, uint64(i.Member.User.ID), int64(charID))
	if err != nil {
		if errors.Is(err, collection.ErrUserDoesNotOwnCharacter) {
			w.Respond(rspErr("You don't own that character"))
			return
		}
		logger.Error("error performing exchange", "error", err, "character_id", charID)
		w.Respond(rspErr("Failed to sell character"))
		return
	}
	w.Respond(Privf("Sold %s for 1 token", char.Name))
}

func (b *Bot) seriesAutocomplete(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[corde.AutocompleteInteractionData]) {
	search, err := i.Data.Options.String("series")
	if err != nil {
		w.Autocomplete(corde.NewResp())
		return
	}

	media, err := b.AnimeService.SearchMedia(ctx, search)
	if err != nil {
		slog.Error("Error searching media", "error", err, "term", search)
		w.Autocomplete(corde.NewResp())
		return
	}

	resp := corde.NewResp()
	for _, m := range media {
		resp.Choice(fmt.Sprintf("%s (%s)", m.Title, strings.ToTitle(strings.ToLower(m.Type))), m.ID)
	}

	w.Autocomplete(resp)
}

func (b *Bot) tokenRoll(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[corde.SlashCommandInteractionData]) {
	logger := slog.With("user_id", uint64(i.Member.User.ID), "guild_id", uint64(i.GuildID))

	mediaID, err := i.Data.Options.Int64("series")
	if err != nil {
		logger.Debug("series roll command: no series selected")
		w.Respond(rspErr("select a series to roll from"))
		return
	}

	config := collection.Config{
		RollCooldown:   b.RollCooldown,
		TokensNeeded:   b.TokensNeeded,
		SeriesRollCost: b.SeriesRollCost,
	}

	char, err := collection.SeriesRoll(ctx, b.Store, b.AnimeService, config, uint64(i.Member.User.ID), mediaID)
	if err != nil {
		if errors.Is(err, collection.ErrInsufficientTokens) {
			w.Respond(rspErr(fmt.Sprintf("You need %d tokens to roll for a series", b.SeriesRollCost)))
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
		logger.Error("error performing series roll", "error", err, "media_id", mediaID)
		w.Respond(rspErr("An error occurred, please try again later"))
		return
	}

	w.Respond(seriesRollEmbed(char, config.SeriesRollCost))
}
