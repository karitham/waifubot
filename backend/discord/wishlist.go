package discord

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/Karitham/corde"

	"github.com/karitham/waifubot/collection"
	"github.com/karitham/waifubot/storage"
	"github.com/karitham/waifubot/storage/collectionstore"
	"github.com/karitham/waifubot/wishlist"
)

// collectionServiceWrapper wraps the storage.Store to implement wishlist.CollectionService
type collectionServiceWrapper struct {
	store storage.Store
}

func (c *collectionServiceWrapper) CheckOwnership(ctx context.Context, userID corde.Snowflake, charID int64) (bool, collectionstore.Character, error) {
	return collection.CheckOwnership(ctx, c.store, userID, charID)
}

func (c *collectionServiceWrapper) GetUserCollectionIDs(ctx context.Context, userID corde.Snowflake) ([]int64, error) {
	return c.store.CollectionStore().ListIDs(ctx, uint64(userID))
}

func (c *collectionServiceWrapper) UpsertCharacter(ctx context.Context, charID int64, name, image string) error {
	_, err := c.store.CollectionStore().UpsertCharacter(ctx, collectionstore.UpsertCharacterParams{
		ID:    charID,
		Name:  name,
		Image: image,
	})
	return err
}

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

func (b *Bot) wishlist(m *corde.Mux) {
	m.Route("character", func(m *corde.Mux) {
		m.Route("add", func(m *corde.Mux) {
			m.SlashCommand("", trace(b.wishlistCharacterAdd))
			m.Autocomplete("character", trace(b.characterAutocomplete))
		})
		m.Route("remove", func(m *corde.Mux) {
			m.SlashCommand("", trace(b.wishlistCharacterRemove))
			m.Autocomplete("character", trace(b.wishlistAutocomplete))
		})
		m.SlashCommand("list", trace(b.wishlistCharacterList))
	})
	m.Route("media", func(m *corde.Mux) {
		m.Route("add", func(m *corde.Mux) {
			m.SlashCommand("", trace(b.wishlistMediaAdd))
			m.Autocomplete("media", trace(b.mediaAutocomplete))
		})
	})
	m.SlashCommand("holders", trace(b.wishlistHolders))
	m.SlashCommand("wanted", trace(b.wishlistWanted))
	m.SlashCommand("compare", trace(b.wishlistCompare))
}

func (b *Bot) wishlistCharacterAdd(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[corde.SlashCommandInteractionData]) {
	logger := slog.With("user_id", uint64(i.Member.User.ID), "guild_id", uint64(i.GuildID))

	charID, _ := i.Data.Options.Int64("character")

	// Check if user already owns this character
	has, char, err := collection.CheckOwnership(ctx, b.Store, i.Member.User.ID, charID)
	if err != nil {
		logger.Error("error checking ownership", "error", err, "character_id", charID)
		w.Respond(rspErr("Unable to verify character ownership. Please try again."))
		return
	}
	if has {
		w.Respond(rspErr("You already own this character."))
		return
	}

	err = wishlist.AddCharacter(ctx, b.WishlistStore, uint64(i.Member.User.ID), charID)
	if err != nil {
		logger.Error("error adding character to wishlist", "error", err, "character_id", charID)
		w.Respond(rspErr("Unable to add character to wishlist. Please try again."))
		return
	}

	w.Respond(corde.NewResp().Contentf("Added %s to your wishlist.", formatCharacter(char.Name, char.ID)).Ephemeral())
}

func (b *Bot) wishlistCharacterRemove(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[corde.SlashCommandInteractionData]) {
	logger := slog.With("user_id", uint64(i.Member.User.ID), "guild_id", uint64(i.GuildID))

	charID, _ := i.Data.Options.Int64("character")

	err := wishlist.RemoveCharacter(ctx, b.WishlistStore, uint64(i.Member.User.ID), charID)
	if err != nil {
		logger.Error("error removing character from wishlist", "error", err, "character_id", charID)
		w.Respond(rspErr("Unable to remove character from wishlist. Please try again."))
		return
	}

	w.Respond(corde.NewResp().Contentf("Removed character %d from your wishlist.", charID).Ephemeral())
}

func (b *Bot) wishlistCharacterList(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[corde.SlashCommandInteractionData]) {
	logger := slog.With("user_id", uint64(i.Member.User.ID), "guild_id", uint64(i.GuildID))

	chars, err := wishlist.GetUserWishlist(ctx, b.WishlistStore, uint64(i.Member.User.ID))
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
		Titlef("%s's Wishlist", i.Member.User.Username).
		Thumbnail(corde.Image{URL: i.Member.User.AvatarPNG()})

	charList := buildCharacterList(chars, 50)
	embed.Description(truncateString(charList, 4096))

	w.Respond(corde.NewResp().Embeds(embed).Ephemeral())
}

// formatUser formats a user ID as a Discord mention
func formatUser(userID uint64) string {
	return fmt.Sprintf("<@%d>", userID)
}

func (b *Bot) wishlistHolders(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[corde.SlashCommandInteractionData]) {
	logger := slog.With("user_id", uint64(i.Member.User.ID), "guild_id", uint64(i.GuildID))

	holders, err := wishlist.GetWishlistHolders(ctx, b.WishlistStore, uint64(i.Member.User.ID), uint64(i.GuildID))
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
		desc.WriteString(fmt.Sprintf("%s: ", formatUser(h.UserID)))
		desc.WriteString(buildCharacterList(h.Characters, len(h.Characters)))
		desc.WriteString("\n")
	}
	embed.Description(truncateString(desc.String(), 4096))

	w.Respond(corde.NewResp().Embeds(embed).Ephemeral())
}

func (b *Bot) wishlistWanted(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[corde.SlashCommandInteractionData]) {
	logger := slog.With("user_id", uint64(i.Member.User.ID), "guild_id", uint64(i.GuildID))

	wanted, err := wishlist.GetWantedCharacters(ctx, b.WishlistStore, uint64(i.Member.User.ID))
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
		desc.WriteString(fmt.Sprintf("%s: ", formatUser(w.UserID)))
		desc.WriteString(buildCharacterList(w.Characters, len(w.Characters)))
		desc.WriteString("\n")
	}
	embed.Description(truncateString(desc.String(), 4096))

	w.Respond(corde.NewResp().Embeds(embed).Ephemeral())
}

func (b *Bot) wishlistCompare(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[corde.SlashCommandInteractionData]) {
	logger := slog.With("user_id", uint64(i.Member.User.ID), "guild_id", uint64(i.GuildID))

	if len(i.Data.Resolved.Users) == 0 {
		w.Respond(rspErr("You must specify a user to compare with."))
		return
	}

	user := i.Data.Resolved.Users.First()

	comparison, err := wishlist.CompareWithUser(ctx, b.WishlistStore, uint64(i.Member.User.ID), uint64(user.ID))
	if err != nil {
		logger.Error("error comparing wishlists", "error", err, "other_user_id", uint64(user.ID))
		w.Respond(rspErr("Unable to compare wishlists. Please try again."))
		return
	}

	embed := corde.NewEmbed().
		Titlef("Wishlist Comparison with %s", user.Username)

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

func (b *Bot) wishlistAutocomplete(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[corde.AutocompleteInteractionData]) {
	input, err := i.Data.Options.String("character")
	if err != nil {
		// If it's an int, convert to string for filtering
		if intVal, intErr := i.Data.Options.Int("character"); intErr == nil {
			input = fmt.Sprintf("%d", intVal)
		} else {
			input = ""
		}
	}

	chars, err := wishlist.GetUserWishlist(ctx, b.WishlistStore, uint64(i.Member.User.ID))
	if err != nil {
		slog.Error("Error getting user's wishlist", "error", err, "user", uint64(i.Member.User.ID))
		return
	}

	resp := corde.NewResp()
	for _, c := range chars {
		charIDStr := fmt.Sprintf("%d", c.ID)
		// Show characters whose name or ID starts with the input
		if input == "" || strings.HasPrefix(strings.ToLower(c.Name), strings.ToLower(input)) || strings.HasPrefix(charIDStr, input) {
			resp.Choice(formatCharacter(c.Name, c.ID), c.ID)
		}
	}

	w.Autocomplete(resp)
}

func (b *Bot) characterAutocomplete(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[corde.AutocompleteInteractionData]) {
	id, err := i.Data.Options.String("character")
	if err != nil {
		i, _ := i.Data.Options.Int("character")
		id = strconv.Itoa(i)
	}

	chars, err := collection.SearchGlobalCharacters(ctx, b.Store, id)
	if err != nil {
		slog.Error("Error searching global characters", "error", err, "term", id)
		return
	}

	resp := corde.NewResp()
	for _, c := range chars {
		resp.Choice(formatCharacter(c.Name, c.ID), c.ID)
	}

	w.Autocomplete(resp)
}

func (b *Bot) wishlistMediaAdd(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[corde.SlashCommandInteractionData]) {
	logger := slog.With("user_id", uint64(i.Member.User.ID), "guild_id", uint64(i.GuildID))

	mediaID, _ := i.Data.Options.Int64("media")

	// Create a collection service wrapper
	collectionService := &collectionServiceWrapper{store: b.Store}

	count, err := wishlist.AddMediaToWishlist(ctx, b.WishlistStore, b.AnimeService, collectionService, i.Member.User.ID, mediaID)
	if err != nil {
		logger.Error("error adding media to wishlist", "error", err, "media_id", mediaID)
		w.Respond(rspErr("Unable to add characters from this media to your wishlist. Please try again."))
		return
	}

	if count == 0 {
		w.Respond(rspErr("No characters found for this media, or you already own all of them."))
		return
	}

	w.Respond(corde.NewResp().Contentf("Added %d characters from this media to your wishlist.", count).Ephemeral())
}

func (b *Bot) mediaAutocomplete(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[corde.AutocompleteInteractionData]) {
	search, err := i.Data.Options.String("media")
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
		displayName := fmt.Sprintf("%s (%s)", m.Title, strings.Title(strings.ToLower(m.Type)))
		resp.Choice(displayName, m.ID)
	}

	w.Autocomplete(resp)
}
