package discord

import (
	"context"
	"fmt"
	"time"

	"github.com/Karitham/corde"
	"github.com/rs/zerolog/log"
)

const (
	AnilistColor   = 0x02a9ff
	AnilistIconURL = "https://anilist.co/img/icons/favicon-32x32.png"
)

// Store is the database
type Store interface {
	PutChar(context.Context, corde.Snowflake, Character) error
	Chars(context.Context, corde.Snowflake) ([]Character, error)
	VerifyChar(context.Context, corde.Snowflake, int64) (Character, error)
	GetCharByID(context.Context, int64) (Character, error)
	CharsIDs(ctx context.Context, userID corde.Snowflake) ([]int64, error)
	DeleteChar(context.Context, corde.Snowflake, int64) (Character, error)
	CharsStartingWith(context.Context, corde.Snowflake, string) ([]Character, error)
	GlobalCharsStartingWith(context.Context, string) ([]Character, error)
	User(context.Context, corde.Snowflake) (User, error)
	ProfileOverview(context.Context, corde.Snowflake) (Profile, error)
	SetUserDate(context.Context, corde.Snowflake, time.Time) error
	SetUserFavorite(context.Context, corde.Snowflake, int64) error
	SetUserQuote(context.Context, corde.Snowflake, string) error
	SetUserAnilistURL(context.Context, corde.Snowflake, string) error
	GiveUserChar(ctx context.Context, dst, src corde.Snowflake, charID int64) error
	AddDropToken(context.Context, corde.Snowflake) error
	ConsumeDropTokens(context.Context, corde.Snowflake, int32) (User, error)
	UsersOwningCharFiltered(ctx context.Context, charID int64, allowedUserIDs []corde.Snowflake) ([]corde.Snowflake, error)
	GetGuildMembers(ctx context.Context, guildID corde.Snowflake) ([]corde.Snowflake, error)
	UsersOwningCharInGuild(ctx context.Context, charID int64, guildID corde.Snowflake) ([]corde.Snowflake, error)
	IsGuildIndexed(ctx context.Context, guildID corde.Snowflake, maxAge time.Duration) (bool, error)
	GetIndexingStatus(ctx context.Context, guildID corde.Snowflake) (string, time.Time, error)
	StartIndexingJob(ctx context.Context, guildID corde.Snowflake) error
	CompleteIndexingJob(ctx context.Context, guildID corde.Snowflake) error
	InsertGuildMembers(ctx context.Context, guildID corde.Snowflake, userIDs []corde.Snowflake) error
	Tx(ctx context.Context, fn func(s Store) error) error
}

// Interacter
type Interacter interface {
	GetInteractionCount(ctx context.Context, channelID corde.Snowflake) (int64, error)
	ResetInteractionCount(ctx context.Context, channelID corde.Snowflake) error
	IncrementInteractionCount(ctx context.Context, channelID corde.Snowflake) error

	SetChannelChar(ctx context.Context, channelID corde.Snowflake, char MediaCharacter) error
	GetChannelChar(ctx context.Context, channelID corde.Snowflake) (MediaCharacter, error)
	RemoveChannelChar(ctx context.Context, channelID corde.Snowflake) error
}

// TrackingService is the interface for the anilist service
type TrackingService interface {
	RandomCharer
	AnimeSearcher
	CharSearcher
	MangaSearcher
	UserSearcher
}

// Bot holds the bot state
type Bot struct {
	mux               *corde.Mux
	Store             Store
	AnimeService      TrackingService
	Inter             Interacter
	AppID             corde.Snowflake
	GuildID           *corde.Snowflake
	BotToken          string
	PublicKey         string
	RollCooldown      time.Duration
	InteractionNeeded int64
	TokensNeeded      int32
}

// New runs the bot
func New(b *Bot) *corde.Mux {
	b.mux = corde.NewMux(b.PublicKey, b.AppID, b.BotToken)
	b.mux.OnNotFound = b.RemoveUnknownCommands

	t := trace[corde.SlashCommandInteractionData]
	i := interact(b.Inter, onInteraction[corde.SlashCommandInteractionData](b))
	idx := indexMiddleware[corde.SlashCommandInteractionData](b)

	b.mux.Route("give", b.give)
	b.mux.Route("search", b.search)
	b.mux.Route("profile", b.profile)
	b.mux.Route("verify", b.verify)
	b.mux.Route("exchange", b.exchange)
	b.mux.Route("holders", b.holders)
	b.mux.SlashCommand("list", wrap(b.list, t, i, idx))
	b.mux.SlashCommand("roll", wrap(b.roll, t, i, idx))
	b.mux.SlashCommand("info", wrap(b.info, t))
	b.mux.SlashCommand("claim", wrap(b.claim, t))

	return b.mux
}

func onInteraction[T corde.InteractionDataConstraint](b *Bot) func(ctx context.Context, count int64, i *corde.Interaction[T]) {
	return func(ctx context.Context, count int64, i *corde.Interaction[T]) {
		if count < b.InteractionNeeded {
			return
		}

		if b.GuildID != nil && *b.GuildID != i.GuildID {
			return
		}

		b.Inter.ResetInteractionCount(ctx, i.ChannelID)
		b.drop(ctx, i.ChannelID)
	}
}

// interaction middleware
func interact[T corde.InteractionDataConstraint](inter Interacter, interact func(ctx context.Context, count int64, i *corde.Interaction[T])) func(func(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[T])) func(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[T]) {
	return func(next func(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[T])) func(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[T]) {
		return func(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[T]) {
			go func() {
				ctx := context.Background()

				err := inter.IncrementInteractionCount(ctx, i.ChannelID)
				if err != nil {
					log.Debug().Err(err).Msg("failed to increment interaction count")
				}

				count, err := inter.GetInteractionCount(ctx, i.ChannelID)
				if err != nil {
					log.Err(err).Msg("failed to get interaction count")
					return
				}

				interact(ctx, count, i)
			}()

			next(ctx, w, i)
		}
	}
}

func wrap[T corde.InteractionDataConstraint](
	next func(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[T]),
	fns ...func(func(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[T])) func(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[T]),
) func(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[T]) {
	// apply middleware in reverse order
	for i := len(fns) - 1; i >= 0; i-- {
		next = fns[i](next)
	}
	return next
}

const maxAge = 7 * 24 * time.Hour

func (b *Bot) indexGuildIfNeeded(ctx context.Context, guildID corde.Snowflake) {
	indexed, err := b.Store.IsGuildIndexed(ctx, guildID, maxAge)
	if err != nil {
		log.Err(err).Uint64("guild_id", uint64(guildID)).Msg("failed to check if guild is indexed")
		return
	}

	if indexed {
		return
	}

	var shouldStart bool
	err = b.Store.Tx(ctx, func(s Store) error {
		status, updatedAt, err := s.GetIndexingStatus(ctx, guildID)
		if err != nil {
			shouldStart = true
			return s.StartIndexingJob(ctx, guildID)
		}
		if status == "in_progress" {
			if time.Since(updatedAt) < 10*time.Minute {
				shouldStart = false
				return nil
			}

			log.Info().Uint64("guild_id", uint64(guildID)).Msg("restarting stale indexing job")
		}

		shouldStart = true
		return s.StartIndexingJob(ctx, guildID)
	})
	if err != nil {
		log.Err(err).Uint64("guild_id", uint64(guildID)).Msg("failed to start indexing job")
		return
	}

	if !shouldStart {
		return
	}

	go b.indexGuild(guildID)
}

func (b *Bot) indexGuild(guildID corde.Snowflake) {
	defer func() {
		if r := recover(); r != nil {
			log.Err(fmt.Errorf("panic in indexing goroutine: %v", r)).Uint64("guild_id", uint64(guildID)).Msg("indexing panic")
		}
	}()

	ctx := context.Background()
	memberIDs, err := FetchGuildMemberIDs(ctx, b.BotToken, guildID)
	if err != nil {
		log.Err(err).Uint64("guild_id", uint64(guildID)).Msg("failed to fetch guild members for indexing")
		return
	}

	err = b.Store.InsertGuildMembers(ctx, guildID, memberIDs)
	if err != nil {
		log.Err(err).Uint64("guild_id", uint64(guildID)).Msg("failed to insert guild members")
		return
	}

	err = b.Store.CompleteIndexingJob(ctx, guildID)
	if err != nil {
		log.Err(err).Uint64("guild_id", uint64(guildID)).Msg("failed to complete indexing job")
		return
	}

	log.Info().Uint64("guild_id", uint64(guildID)).Int("members", len(memberIDs)).Msg("completed guild indexing")
}

func (b *Bot) PerformRoll(ctx context.Context, userID corde.Snowflake) (MediaCharacter, error) {
	var char MediaCharacter

	err := b.Store.Tx(ctx, func(s Store) error {
		charsIDs, err := s.CharsIDs(ctx, userID)
		if err != nil {
			return err
		}

		c, err := b.AnimeService.RandomChar(ctx, charsIDs...)
		if err != nil {
			return err
		}
		char = MediaCharacter{
			ID:          c.ID,
			Name:        c.Name,
			ImageURL:    c.ImageURL,
			URL:         c.URL,
			Description: c.Description,
			MediaTitle:  c.MediaTitle,
		}

		return s.PutChar(ctx, userID, Character{
			Date:   time.Now(),
			Image:  c.ImageURL,
			Name:   c.Name,
			Type:   "ROLL",
			UserID: userID,
			ID:     int64(c.ID),
		})
	})

	return char, err
}

func (b *Bot) PerformGive(ctx context.Context, from, to corde.Snowflake, charID int64) (Character, error) {
	c, err := b.Store.VerifyChar(ctx, from, charID)
	if err != nil {
		return Character{}, fmt.Errorf("from user does not own char %d: %w", charID, err)
	}

	_, err = b.Store.VerifyChar(ctx, to, charID)
	if err == nil {
		return Character{}, fmt.Errorf("to user already owns char %d", charID)
	}

	err = b.Store.GiveUserChar(ctx, to, from, charID)
	if err != nil {
		return Character{}, fmt.Errorf("error giving char: %w", err)
	}

	return c, nil
}

func (b *Bot) RemoveUnknownCommands(ctx context.Context, r corde.ResponseWriter, i *corde.Interaction[corde.JsonRaw]) {
	log.Error().Str("command", i.Route).Int("type", int(i.Type)).Msg("Unknown command")
	r.Respond(corde.NewResp().Content("I don't know what that means, you shouldn't be able to do that").Ephemeral())

	var opt []func(*corde.CommandsOpt)
	if b.GuildID != nil {
		opt = append(opt, corde.GuildOpt(*b.GuildID))
	}

	b.mux.DeleteCommand(i.ID, opt...)
}
