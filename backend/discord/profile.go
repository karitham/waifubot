package discord

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"strings"
	"time"

	"github.com/Karitham/corde"

	"github.com/karitham/waifubot/catalog"
	"github.com/karitham/waifubot/collection"
)

// ProfileHandler handles the /profile command and its subcommands.
type ProfileHandler struct {
	store collection.Store
}

// Register wires the profile sub-routes on the mux.
func (h *ProfileHandler) Register(m *corde.Mux) {
	m.SlashCommand("view", trace(wrapCtx(h.View)))
	m.Route("edit", func(m *corde.Mux) {
		m.SlashCommand("quote", trace(wrapCtx(h.EditQuote)))
		m.Route("favorite", func(m *corde.Mux) {
			m.SlashCommand("", trace(wrapCtx(h.EditFavorite)))
			m.Autocomplete("id", h.Autocomplete)
		})
		m.SlashCommand("anilist", trace(wrapCtx(h.EditAnilistURL)))
	})
}

// profileViewOptions holds the parsed options for the profile view command.
type profileViewOptions struct {
	targetUserID   uint64
	targetUsername string
}

func parseProfileViewOptions(cmd CommandContext) profileViewOptions {
	opts := profileViewOptions{
		targetUserID:   cmd.UserID(),
		targetUsername: cmd.Username(),
	}
	if user, ok := cmd.FirstResolvedUser(); ok {
		opts.targetUserID = uint64(user.ID)
		opts.targetUsername = user.Username
	}
	return opts
}

// View displays a user's profile.
func (h *ProfileHandler) View(ctx context.Context, w corde.ResponseWriter, cmd CommandContext) {
	logger := slog.With("user_id", cmd.UserID(), "guild_id", cmd.GuildID())

	opts := parseProfileViewOptions(cmd)

	data, err := collection.UserProfile(ctx, h.store, opts.targetUserID)
	if err != nil {
		logger.Error("error getting profile", "error", err, "target_user_id", opts.targetUserID)
		w.Respond(corde.NewResp().Content("An error occurred dialing the database, please try again later").Ephemeral())
		return
	}

	anilistURLDesc := ""
	if data.AnilistURL != "" {
		anilistURLDesc = fmt.Sprintf("Find them on [Anilist](%s)", data.AnilistURL)
	}

	resp := corde.NewEmbed().
		Title(opts.targetUsername).
		URL(fmt.Sprintf("https://waifugui.karitham.dev/#/list/%d", opts.targetUserID)).
		Descriptionf(
			"%s\n%s last rolled %s ago and has %d tokens.\nThey have %d characters.\nFavorite: %s\n%s",
			data.Quote,
			opts.targetUsername,
			time.Since(data.Date.UTC()).Truncate(time.Second),
			data.Tokens,
			data.CharacterCount,
			data.Favorite.Name,
			anilistURLDesc,
		).
		Field("Collection", fmt.Sprintf("[View Collection](https://waifugui.karitham.dev/#/list/%d)", opts.targetUserID)).
		Field("Wishlist", fmt.Sprintf("[View Wishlist](https://waifugui.karitham.dev/#/wishlist/%d)", opts.targetUserID))
	if data.Favorite.Image != "" {
		resp.Thumbnail(corde.Image{URL: data.Favorite.Image})
	}

	w.Respond(resp)
}

// EditFavorite sets the user's favorite character.
func (h *ProfileHandler) EditFavorite(ctx context.Context, w corde.ResponseWriter, cmd CommandContext) {
	logger := slog.With("user_id", cmd.UserID(), "guild_id", cmd.GuildID())

	optID, _ := cmd.OptInt64("id")
	if err := h.store.UpdateFavorite(ctx, cmd.UserID(), optID); err != nil {
		logger.Error("error setting favorite character", "error", err, "character_id", optID)
		w.Respond(corde.NewResp().Content("An error occurred setting this character").Ephemeral())
		return
	}

	w.Respond(corde.NewResp().Contentf("Favorite character set as char id %d", optID).Ephemeral())
}

// EditQuote sets the user's profile quote.
func (h *ProfileHandler) EditQuote(ctx context.Context, w corde.ResponseWriter, cmd CommandContext) {
	quote, _ := cmd.OptString("value")
	if len(quote) > 1024 {
		w.Respond(corde.NewResp().Content("quote is too long").Ephemeral())
		return
	}

	if err := h.store.UpdateQuote(ctx, cmd.UserID(), quote); err != nil {
		w.Respond(corde.NewResp().Content(err.Error()).Ephemeral())
		return
	}

	w.Respond(corde.NewResp().Content("Quote set").Ephemeral())
}

// EditAnilistURL sets the user's Anilist profile URL.
func (h *ProfileHandler) EditAnilistURL(ctx context.Context, w corde.ResponseWriter, cmd CommandContext) {
	anilistURL, _ := cmd.OptString("url")
	parsedURL, err := url.Parse(anilistURL)
	if err != nil {
		w.Respond(corde.NewResp().Content("invalid URL").Ephemeral())
		return
	}

	if parsedURL.Host != "anilist.co" {
		w.Respond(corde.NewResp().Content("invalid Anilist URL").Ephemeral())
		return
	}

	if !strings.HasPrefix(parsedURL.Path, "/user/") {
		w.Respond(corde.NewResp().Content("invalid Anilist URL").Ephemeral())
		return
	}

	if err := h.store.UpdateAnilistURL(ctx, cmd.UserID(), anilistURL); err != nil {
		w.Respond(corde.NewResp().Content(err.Error()).Ephemeral())
		return
	}

	w.Respond(corde.NewResp().Contentf("Anilist URL set as %s", anilistURL).Ephemeral())
}

// Autocomplete provides character suggestions for the profile edit favorite command.
func (h *ProfileHandler) Autocomplete(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[corde.AutocompleteInteractionData]) {
	autocomplete(ctx, w, i, "id", func(ctx context.Context, input string) ([]catalog.Character, error) {
		return h.store.SearchCharacters(ctx, uint64(i.Member.User.ID), input)
	}, formatCharacterChoice)
}
