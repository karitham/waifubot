package discord

import (
	"context"
	"log/slog"
	"regexp"

	"github.com/Karitham/corde"

	"github.com/karitham/waifubot/collection"
	"github.com/karitham/waifubot/guild"
	"github.com/karitham/waifubot/storage/interactionstore"
)

// SearchHandler handles the /search command and its subcommands.
type SearchHandler struct {
	animeService  TrackingService
	interStore    interactionstore.Store
	onInteraction func(context.Context, int64, *corde.Interaction[corde.SlashCommandInteractionData])
	guildIndexer  *guild.Indexer
	guildTxFn     func(context.Context) (guild.TxQuerier, error)
}

// Register wires the search sub-routes on the mux.
func (h *SearchHandler) Register(m *corde.Mux) {
	t := trace[corde.SlashCommandInteractionData]
	i := interact(h.interStore, h.onInteraction)
	idx := indexMiddleware[corde.SlashCommandInteractionData](h.guildIndexer, h.guildTxFn)

	m.SlashCommand("char", wrap(wrapCtx(h.SearchChar), t, i, idx))
	m.SlashCommand("user", wrap(wrapCtx(h.SearchUser), t, i, idx))
	m.SlashCommand("manga", wrap(wrapCtx(h.SearchManga), t, i, idx))
	m.SlashCommand("anime", wrap(wrapCtx(h.SearchAnime), t, i, idx))
}

// SearchAnime searches for anime by name.
func (h *SearchHandler) SearchAnime(ctx context.Context, w corde.ResponseWriter, cmd CommandContext) {
	search, _ := cmd.OptString("name")

	media, err := h.animeService.Anime(ctx, search)
	if err != nil {
		slog.ErrorContext(ctx, "error with anime service", "error", err)
		w.Respond(rspErr("Error searching for this anime, either it doesn't exist or something went wrong"))
		return
	}

	if len(media) == 0 {
		w.Respond(rspErr("No anime found with that name"))
		return
	}

	w.Respond(mediaEmbed(media[0]))
}

// SearchManga searches for manga by name.
func (h *SearchHandler) SearchManga(ctx context.Context, w corde.ResponseWriter, cmd CommandContext) {
	search, _ := cmd.OptString("name")

	media, err := h.animeService.Manga(ctx, search)
	if err != nil {
		slog.ErrorContext(ctx, "error with anime service", "error", err)
		w.Respond(rspErr("Error searching for this manga, either it doesn't exist or something went wrong"))
		return
	}

	if len(media) == 0 {
		w.Respond(rspErr("No manga found with that name"))
		return
	}

	w.Respond(mediaEmbed(media[0]))
}

// SearchUser searches for an AniList user by name.
func (h *SearchHandler) SearchUser(ctx context.Context, w corde.ResponseWriter, cmd CommandContext) {
	search, _ := cmd.OptString("name")

	users, err := h.animeService.User(ctx, search)
	if err != nil {
		slog.ErrorContext(ctx, "error with user service", "error", err)
		w.Respond(rspErr("Error searching for this user, either it doesn't exist or something went wrong"))
		return
	}

	if len(users) == 0 {
		w.Respond(rspErr("No user found with that name"))
		return
	}

	w.Respond(userEmbed(users[0]))
}

// SearchChar searches for a character by name.
func (h *SearchHandler) SearchChar(ctx context.Context, w corde.ResponseWriter, cmd CommandContext) {
	search, _ := cmd.OptString("name")

	characters, err := h.animeService.Character(ctx, search)
	if err != nil {
		slog.ErrorContext(ctx, "error with char service", "error", err)
		w.Respond(rspErr("Error searching for this character, either it doesn't exist or something went wrong"))
		return
	}

	if len(characters) == 0 {
		w.Respond(rspErr("No character found with that name"))
		return
	}

	w.Respond(charEmbed(characters[0]))
}

func mediaEmbed(m collection.Media) *corde.EmbedB {
	return applyEmbedOpt(corde.NewEmbed().
		Title(m.Title).
		URL(m.URL).
		Color(m.CoverImageColor).
		ImageURL(m.BannerImageURL).
		Thumbnail(corde.Image{URL: m.CoverImageURL}).
		Description(m.Description),
		anilistFooter,
		sanitizeDescOpt,
	)
}

func userEmbed(u collection.TrackerUser) *corde.EmbedB {
	return applyEmbedOpt(corde.NewEmbed().
		Title(u.Name).
		URL(u.URL).
		Color(AnilistColor).
		ImageURL(u.ImageURL).
		Description(u.About),
		anilistFooter,
		sanitizeDescOpt,
	)
}

func charEmbed(c collection.MediaCharacter) *corde.EmbedB {
	return applyEmbedOpt(corde.NewEmbed().
		Title(c.Name).
		Color(AnilistColor).
		URL(c.URL).
		Thumbnail(corde.Image{URL: c.ImageURL}).
		Description(c.Description),
		anilistFooter,
		sanitizeDescOpt,
	)
}

func anilistFooter(b *corde.EmbedB) *corde.EmbedB {
	return b.Footer(corde.Footer{
		Text:    "View on Anilist",
		IconURL: AnilistIconURL,
	})
}

func applyEmbedOpt(b *corde.EmbedB, opts ...func(*corde.EmbedB) *corde.EmbedB) *corde.EmbedB {
	for _, opt := range opts {
		b = opt(b)
	}
	return b
}

func sanitizeDescOpt(b *corde.EmbedB) *corde.EmbedB {
	return b.Description(sanitizeDesc(b.Embed().Description))
}

var descRegexp = regexp.MustCompile(`!~|~!`)

func sanitizeDesc(desc string) string {
	return descRegexp.ReplaceAllString(desc, "||")
}
