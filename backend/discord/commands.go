package discord

import (
	"github.com/Karitham/corde"
	"github.com/rs/zerolog/log"
)

// This init is called when ran with the build tag.
func (b *Bot) RegisterCommands() error {
	nameOpt := corde.NewStringOption("name", "name you wish to search", true)

	commands := []corde.CreateCommander{
		corde.NewSlashCommand("list", "list owned characters",
			corde.NewUserOption("user", "user to list characters for", false),
		),

		corde.NewSlashCommand("verify", "assert a user owns a character",
			corde.NewIntOption("id", "id of the character", true).CanAutocomplete(),
			corde.NewUserOption("user", "user which is supposed to own that character", false),
		),

		corde.NewSlashCommand("exchange", "exchange a character against a token",
			corde.NewIntOption("id", "id of the character", true).CanAutocomplete(),
		),

		corde.NewSlashCommand("roll", "roll a random character"),

		corde.NewSlashCommand("search", "search for anything on anilist",
			corde.NewSubcommand("anime", "search for an anime", nameOpt),
			corde.NewSubcommand("manga", "search for a manga", nameOpt),
			corde.NewSubcommand("char", "search for a character", nameOpt),
			corde.NewSubcommand("user", "search for a user", nameOpt),
		),

		corde.NewSlashCommand("profile", "interact with your profile or view someone else's",
			corde.NewSubcommand("view", "view a user's profile",
				corde.NewUserOption("user", "user to view profile for", false),
			),
			corde.NewSubcommandGroup("edit", "edit your profile",
				corde.NewSubcommand("favorite", "set your favorite character",
					corde.NewIntOption("id", "id of the character", true).CanAutocomplete(),
				),
				corde.NewSubcommand("quote", "set your quote",
					corde.NewStringOption("value", "quote value to set", true),
				),
				corde.NewSubcommand("anilist", "set your anilist url",
					corde.NewStringOption("url", "anilist url to set. In the form `https://anilist.co/user/{{User}}`", true),
				),
			),
		),

		corde.NewSlashCommand("give", "give a character to someone",
			corde.NewUserOption("user", "user to give character to", true),
			corde.NewIntOption("id", "id of the character", true).CanAutocomplete(),
		),

		corde.NewSlashCommand("claim", "claim a character",
			corde.NewStringOption("name", "character's name", true),
		),

		corde.NewSlashCommand("info", "information about the bot"),
	}

	var opt []func(*corde.CommandsOpt)
	if b.GuildID != nil {
		opt = append(opt, corde.GuildOpt(*b.GuildID))
	}

	log.Info().Msgf("registering commands for app %d", b.AppID)
	return corde.NewMux(b.PublicKey, b.AppID, b.BotToken).BulkRegisterCommand(commands, opt...)
}
