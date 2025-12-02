package discord

import (
	"log/slog"

	"github.com/Karitham/corde"
)

// This init is called when ran with the build tag.
func (b *Bot) RegisterCommands() error {
	nameOpt := corde.NewStringOption("name", "name you wish to search", true)

	commands := []corde.CreateCommander{
		corde.NewSlashCommand("list", "View a user's character collection",
			corde.NewUserOption("user", "User to list characters for (optional)", false),
		),

		corde.NewSlashCommand("verify", "Check if a user owns a specific character",
			corde.NewIntOption("id", "ID of the character to verify", true).CanAutocomplete(),
			corde.NewUserOption("user", "User to check (optional, defaults to you)", false),
		),

		corde.NewSlashCommand("exchange", "Exchange a character for tokens",
			corde.NewIntOption("id", "ID of the character to exchange", true).CanAutocomplete(),
		),

		corde.NewSlashCommand("roll", "Roll for a random character"),

		corde.NewSlashCommand("search", "Search AniList for anime, manga, characters, or users",
			corde.NewSubcommand("anime", "Search for an anime by name", nameOpt),
			corde.NewSubcommand("manga", "Search for a manga by name", nameOpt),
			corde.NewSubcommand("char", "Search for a character by name", nameOpt),
			corde.NewSubcommand("user", "Search for a user by name", nameOpt),
		),

		corde.NewSlashCommand("profile", "View or edit user profiles",
			corde.NewSubcommand("view", "View a user's profile",
				corde.NewUserOption("user", "User to view profile for (optional)", false),
			),
			corde.NewSubcommandGroup("edit", "Edit your profile",
				corde.NewSubcommand("favorite", "Set your favorite character",
					corde.NewIntOption("id", "ID of the character", true).CanAutocomplete(),
				),
				corde.NewSubcommand("quote", "Set your profile quote",
					corde.NewStringOption("value", "Quote to set", true),
				),
				corde.NewSubcommand("anilist", "Set your AniList profile URL",
					corde.NewStringOption("url", "AniList URL (e.g., https://anilist.co/user/Username)", true),
				),
			),
		),

		corde.NewSlashCommand("give", "Give a character to another user",
			corde.NewUserOption("user", "User to give the character to", true),
			corde.NewIntOption("id", "ID of the character to give", true).CanAutocomplete(),
		),

		corde.NewSlashCommand("claim", "Claim a character by name",
			corde.NewStringOption("name", "Name of the character to claim", true),
		),

		corde.NewSlashCommand("info", "Get information about the bot"),

		corde.NewSlashCommand("holders", "List users in this server who own a specific character",
			corde.NewIntOption("id", "ID of the character", true).CanAutocomplete(),
		),

		corde.NewSlashCommand("wishlist", "Manage your character wishlist",
			corde.NewSubcommandGroup("character", "Manage characters in your wishlist",
				corde.NewSubcommand("add", "Add a character to your wishlist",
					corde.NewIntOption("character", "ID of the character to add", true).CanAutocomplete(),
				),
				corde.NewSubcommand("remove", "Remove a character from your wishlist",
					corde.NewIntOption("character", "ID of the character to remove", true).CanAutocomplete(),
				),
				corde.NewSubcommand("list", "List all characters in your wishlist"),
				corde.NewSubcommand("remove-all", "Remove all characters from your wishlist"),
			),
			corde.NewSubcommandGroup("media", "Manage media in your wishlist",
				corde.NewSubcommand("add", "Add all characters from a show/manga to your wishlist",
					corde.NewIntOption("media", "ID of the media to add", true).CanAutocomplete(),
				),
			),
			corde.NewSubcommand("holders", "Show users who have characters from your wishlist"),
			corde.NewSubcommand("wanted", "Show users who want characters you own"),
			corde.NewSubcommand("compare", "Compare your wishlist with another user's collection",
				corde.NewUserOption("user", "User to compare with", true),
			),
		),
	}

	var opt []func(*corde.CommandsOpt)
	if b.GuildID != nil {
		opt = append(opt, corde.GuildOpt(*b.GuildID))
	}

	slog.Info("registering commands", "app_id", b.AppID)
	return corde.NewMux(b.PublicKey, b.AppID, b.BotToken).BulkRegisterCommand(commands, opt...)
}
