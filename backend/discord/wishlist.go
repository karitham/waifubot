package discord

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/Karitham/corde"

	"github.com/karitham/waifubot/catalog"
	"github.com/karitham/waifubot/collection"
	"github.com/karitham/waifubot/guild"
	"github.com/karitham/waifubot/wishlist"
)

// formatCharacter formats a character as "Name (ID)"
func formatCharacter(name string, id int64) string {
	return fmt.Sprintf("%s (%d)", name, id)
}

// buildCharacterList builds a comma-separated list of characters, limited to maxItems
func buildCharacterList(chars []wishlist.Character, maxItems int) string {
	var b strings.Builder
	for i, c := range chars {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(formatCharacter(c.Name, c.ID))
		if i >= maxItems-1 {
			if len(chars) > maxItems {
				b.WriteString("...")
			}
			break
		}
	}
	return b.String()
}

// truncateString truncates a string to maxLen, adding "..." if truncated
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen < 3 {
		return "..."
	}
	return s[:maxLen-3] + "..."
}

// formatUser formats a user ID as a Discord mention
func formatUser(userID uint64) string {
	return fmt.Sprintf("<@%d>", userID)
}

// WishlistHandler handles the /wishlist command and its subcommands.
type WishlistHandler struct {
	wishlist     wishlist.Store
	store        collection.Store
	animeService TrackingService
	catalog      catalog.Store
	guildIndexer *guild.Indexer
	guildTxFn    func(context.Context) (guild.TxQuerier, error)
}

// Register wires the wishlist sub-routes on the mux.
func (h *WishlistHandler) Register(m *corde.Mux) {
	m.Route("character", func(m *corde.Mux) {
		m.Route("add", func(m *corde.Mux) {
			m.SlashCommand("", trace(wrapCtx(h.CharacterAdd)))
			m.Autocomplete("character", h.CharacterAutocomplete)
		})
		m.Route("remove", func(m *corde.Mux) {
			m.SlashCommand("", trace(wrapCtx(h.CharacterRemove)))
			m.Autocomplete("character", h.WishlistAutocomplete)
		})
		m.SlashCommand("list", trace(wrapCtx(h.CharacterList)))
		m.SlashCommand("remove-all", trace(wrapCtx(h.CharacterRemoveAll)))
	})
	m.Route("media", func(m *corde.Mux) {
		m.Route("add", func(m *corde.Mux) {
			m.SlashCommand("", trace(wrapCtx(h.MediaAdd)))
			m.Autocomplete("media", h.MediaAutocomplete)
		})
	})
	m.SlashCommand("holders", wrap(wrapCtx(h.Holders), indexMiddleware[corde.SlashCommandInteractionData](h.guildIndexer, h.guildTxFn), trace[corde.SlashCommandInteractionData]))
	m.SlashCommand("wanted", wrap(wrapCtx(h.Wanted), indexMiddleware[corde.SlashCommandInteractionData](h.guildIndexer, h.guildTxFn), trace[corde.SlashCommandInteractionData]))
	m.SlashCommand("compare", trace(wrapCtx(h.Compare)))
}

// wishlistCharacterOptions holds the parsed options for wishlist character add/remove commands.
type wishlistCharacterOptions struct {
	charID int64
}

func parseWishlistCharacterOptions(cmd CommandContext) (wishlistCharacterOptions, error) {
	charID, _ := cmd.OptInt64("character")
	return wishlistCharacterOptions{charID: charID}, nil
}

// CharacterAdd adds a character to the user's wishlist.
func (h *WishlistHandler) CharacterAdd(ctx context.Context, w corde.ResponseWriter, cmd CommandContext) {
	logger := slog.With("user_id", cmd.UserID(), "guild_id", cmd.GuildID())

	opts, err := parseWishlistCharacterOptions(cmd)
	if err != nil {
		opts.charID, _ = cmd.OptInt64("character")
	}

	// Check if user already owns this character
	has, char, err := collectionCheckOwnership(ctx, h.store, cmd.UserID(), opts.charID)
	if err != nil {
		logger.Error("error checking ownership", "error", err, "character_id", opts.charID)
		w.Respond(rspErr("Unable to verify character ownership. Please try again."))
		return
	}
	if has {
		w.Respond(rspErr("You already own this character."))
		return
	}

	err = h.wishlist.AddCharactersToWishlist(ctx, cmd.UserID(), []int64{opts.charID})
	if err != nil {
		logger.Error("error adding character to wishlist", "error", err, "character_id", opts.charID)
		w.Respond(rspErr("Unable to add character to wishlist. Please try again."))
		return
	}

	w.Respond(corde.NewResp().Contentf("Added %s to your wishlist.", formatCharacter(char.Name, char.ID)).Ephemeral())
}

// CharacterRemove removes a character from the user's wishlist.
func (h *WishlistHandler) CharacterRemove(ctx context.Context, w corde.ResponseWriter, cmd CommandContext) {
	logger := slog.With("user_id", cmd.UserID(), "guild_id", cmd.GuildID())

	opts, err := parseWishlistCharacterOptions(cmd)
	if err != nil {
		opts.charID, _ = cmd.OptInt64("character")
	}

	err = h.wishlist.RemoveCharactersFromWishlist(ctx, cmd.UserID(), []int64{opts.charID})
	if err != nil {
		logger.Error("error removing character from wishlist", "error", err, "character_id", opts.charID)
		w.Respond(rspErr("Unable to remove character from wishlist. Please try again."))
		return
	}

	w.Respond(corde.NewResp().Contentf("Removed character %d from your wishlist.", opts.charID).Ephemeral())
}

// CharacterList displays the user's wishlist characters.
func (h *WishlistHandler) CharacterList(ctx context.Context, w corde.ResponseWriter, cmd CommandContext) {
	logger := slog.With("user_id", cmd.UserID(), "guild_id", cmd.GuildID())

	chars, err := h.wishlist.GetUserCharacterWishlist(ctx, cmd.UserID())
	if err != nil {
		logger.Error("error getting user wishlist", "error", err)
		w.Respond(rspErr("Unable to retrieve your wishlist. Please try again."))
		return
	}

	if len(chars) == 0 {
		w.Respond(corde.NewResp().Content("Your wishlist is empty.").Ephemeral())
		return
	}

	embed := corde.NewEmbed().
		Titlef("%s's Wishlist", cmd.Username()).
		Thumbnail(corde.Image{URL: cmd.AvatarPNG()})

	charList := buildCharacterList(chars, 50)
	embed.Description(truncateString(charList, 4096))

	w.Respond(corde.NewResp().Embeds(embed).Ephemeral())
}

// CharacterRemoveAll clears the user's wishlist.
func (h *WishlistHandler) CharacterRemoveAll(ctx context.Context, w corde.ResponseWriter, cmd CommandContext) {
	logger := slog.With("user_id", cmd.UserID(), "guild_id", cmd.GuildID())

	err := h.wishlist.RemoveAllFromWishlist(ctx, cmd.UserID())
	if err != nil {
		logger.Error("error removing all from wishlist", "error", err)
		w.Respond(rspErr("Unable to clear your wishlist. Please try again."))
		return
	}

	w.Respond(corde.NewResp().Content("Cleared your wishlist.").Ephemeral())
}

// wishlistMediaOptions holds the parsed options for wishlist media add command.
type wishlistMediaOptions struct {
	mediaID int64
}

func parseWishlistMediaOptions(cmd CommandContext) (wishlistMediaOptions, error) {
	mediaID, _ := cmd.OptInt64("media")
	return wishlistMediaOptions{mediaID: mediaID}, nil
}

// MediaAdd adds all characters from a media to the user's wishlist.
func (h *WishlistHandler) MediaAdd(ctx context.Context, w corde.ResponseWriter, cmd CommandContext) {
	logger := slog.With("user_id", cmd.UserID(), "guild_id", cmd.GuildID())

	opts, err := parseWishlistMediaOptions(cmd)
	if err != nil {
		opts.mediaID, _ = cmd.OptInt64("media")
	}

	count, err := wishlist.AddMediaToWishlist(ctx, h.wishlist, h.animeService, h.store, cmd.UserID(), opts.mediaID)
	if err != nil {
		logger.Error("error adding media to wishlist", "error", err, "media_id", opts.mediaID)
		w.Respond(rspErr("Unable to add characters from this media to your wishlist. Please try again."))
		return
	}

	if count == 0 {
		w.Respond(rspErr("No characters found for this media, or you already own all of them."))
		return
	}

	w.Respond(corde.NewResp().Contentf("Added %d characters from this media to your wishlist.", count).Ephemeral())
}

// Holders shows which guild members have characters from the user's wishlist.
func (h *WishlistHandler) Holders(ctx context.Context, w corde.ResponseWriter, cmd CommandContext) {
	logger := slog.With("user_id", cmd.UserID(), "guild_id", cmd.GuildID())

	wl, err := h.wishlist.GetUserCharacterWishlist(ctx, cmd.UserID())
	if err != nil {
		logger.Error("error getting user wishlist", "error", err)
		w.Respond(rspErr("Unable to retrieve your wishlist. Please try again."))
		return
	}

	if len(wl) == 0 {
		w.Respond(corde.NewResp().Content("Your wishlist is empty.").Ephemeral())
		return
	}

	characterIDs := make([]int64, len(wl))
	for j, c := range wl {
		characterIDs[j] = c.ID
	}

	holders, err := h.wishlist.GetWishlistHolders(ctx, characterIDs, cmd.UserID(), cmd.GuildID())
	if err != nil {
		logger.Error("error getting wishlist holders", "error", err)
		w.Respond(rspErr("Unable to retrieve wishlist holders. Please try again."))
		return
	}

	if len(holders) == 0 {
		w.Respond(corde.NewResp().Content("No one has characters from your wishlist.").Ephemeral())
		return
	}

	embed := corde.NewEmbed().
		Title("Characters from Your Wishlist")

	var desc strings.Builder
	desc.WriteString("Users who have characters from your wishlist:\n")
	for _, h := range holders {
		if len(h.Characters) == 0 {
			continue
		}
		fmt.Fprintf(&desc, "%s: ", formatUser(h.UserID))
		desc.WriteString(buildCharacterList(h.Characters, len(h.Characters)))
		desc.WriteString("\n")
	}
	embed.Description(truncateString(desc.String(), 4096))

	w.Respond(corde.NewResp().Embeds(embed).Ephemeral())
}

// Wanted shows which guild members want characters from the user's collection.
func (h *WishlistHandler) Wanted(ctx context.Context, w corde.ResponseWriter, cmd CommandContext) {
	logger := slog.With("user_id", cmd.UserID(), "guild_id", cmd.GuildID())

	wanted, err := h.wishlist.GetWantedCharacters(ctx, cmd.UserID(), cmd.GuildID())
	if err != nil {
		logger.Error("error getting wanted characters", "error", err)
		w.Respond(rspErr("Unable to retrieve wanted characters. Please try again."))
		return
	}

	if len(wanted) == 0 {
		w.Respond(corde.NewResp().Content("No one wants characters from your collection.").Ephemeral())
		return
	}

	embed := corde.NewEmbed().
		Title("People Who Want Your Characters")

	var desc strings.Builder
	desc.WriteString("People who want characters from your collection:\n")
	for _, w := range wanted {
		fmt.Fprintf(&desc, "%s: ", formatUser(w.UserID))
		desc.WriteString(buildCharacterList(w.Characters, len(w.Characters)))
		desc.WriteString("\n")
	}
	embed.Description(truncateString(desc.String(), 4096))

	w.Respond(corde.NewResp().Embeds(embed).Ephemeral())
}

// wishlistCompareOptions holds the parsed options for the wishlist compare command.
type wishlistCompareOptions struct {
	targetUserID   uint64
	targetUsername string
}

func parseWishlistCompareOptions(cmd CommandContext) (wishlistCompareOptions, error) {
	if user, ok := cmd.FirstResolvedUser(); ok {
		return wishlistCompareOptions{
			targetUserID:   uint64(user.ID),
			targetUsername: user.Username,
		}, nil
	}
	return wishlistCompareOptions{}, fmt.Errorf("you must specify a user to compare with")
}

// Compare compares wishlists with another user.
func (h *WishlistHandler) Compare(ctx context.Context, w corde.ResponseWriter, cmd CommandContext) {
	logger := slog.With("user_id", cmd.UserID(), "guild_id", cmd.GuildID())

	opts, err := parseWishlistCompareOptions(cmd)
	if err != nil {
		w.Respond(rspErr(err.Error()))
		return
	}

	comparison, err := h.wishlist.CompareWithUser(ctx, cmd.UserID(), opts.targetUserID)
	if err != nil {
		logger.Error("error comparing wishlists", "error", err, "other_user_id", opts.targetUserID)
		w.Respond(rspErr("Unable to compare wishlists. Please try again."))
		return
	}

	embed := corde.NewEmbed().
		Titlef("Wishlist Comparison with %s", opts.targetUsername)

	if len(comparison.UserHasCharacters) > 0 {
		hasList := buildCharacterList(comparison.UserHasCharacters, len(comparison.UserHasCharacters))
		embed.Field("They Have from Your Wishlist", truncateString(hasList, 1024))
	}

	if len(comparison.UserWantsCharacters) > 0 {
		wantsList := buildCharacterList(comparison.UserWantsCharacters, len(comparison.UserWantsCharacters))
		embed.Field("You Have from Their Wishlist", truncateString(wantsList, 1024))
	}

	if comparison.MutualMatches > 0 {
		embed.Field("Mutual Matches", fmt.Sprintf("%d", comparison.MutualMatches))
	}

	if len(comparison.UserHasCharacters) == 0 && len(comparison.UserWantsCharacters) == 0 {
		embed.Description("No matches found.")
	}

	w.Respond(corde.NewResp().Embeds(embed).Ephemeral())
}

// CharacterAutocomplete provides character suggestions for the wishlist character add command.
func (h *WishlistHandler) CharacterAutocomplete(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[corde.AutocompleteInteractionData]) {
	autocomplete(ctx, w, i, "character", h.catalog.SearchGlobalCharacters, formatCharacterChoice)
}

// WishlistAutocomplete provides character suggestions from the user's wishlist.
func (h *WishlistHandler) WishlistAutocomplete(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[corde.AutocompleteInteractionData]) {
	input, err := i.Data.Options.String("character")
	if err != nil {
		if intVal, intErr := i.Data.Options.Int("character"); intErr == nil {
			input = fmt.Sprintf("%d", intVal)
		} else {
			input = ""
		}
	}

	chars, err := h.wishlist.GetUserCharacterWishlist(ctx, uint64(i.Member.User.ID))
	if err != nil {
		slog.Error("Error getting user's wishlist", "error", err, "user", uint64(i.Member.User.ID))
		return
	}

	resp := corde.NewResp()
	for _, c := range chars {
		charIDStr := fmt.Sprintf("%d", c.ID)
		if input == "" || strings.HasPrefix(strings.ToLower(c.Name), strings.ToLower(input)) || strings.HasPrefix(charIDStr, input) {
			resp.Choice(formatCharacter(c.Name, c.ID), c.ID)
		}
	}

	w.Autocomplete(resp)
}

// MediaAutocomplete provides media suggestions for the wishlist media add command.
func (h *WishlistHandler) MediaAutocomplete(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[corde.AutocompleteInteractionData]) {
	autocomplete(ctx, w, i, "media", h.animeService.SearchMedia, formatMediaChoice)
}
