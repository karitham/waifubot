package discord

var commandDefinitions = []CommandDef{
	{Name: "list", Description: "View a user's character collection",
		Options: []OptionDef{
			{Name: "user", Description: "User to list characters for (optional)", Type: OptionUser},
		},
	},
	{Name: "verify", Description: "Check if a user owns a specific character",
		Options: []OptionDef{
			{Name: "id", Description: "ID of the character to verify", Type: OptionInt, Required: true, Autocomplete: true},
			{Name: "user", Description: "User to check (optional, defaults to you)", Type: OptionUser},
		},
	},
	{Name: "token", Description: "Manage your tokens",
		Options: []OptionDef{
			{Name: "balance", Description: "View your token balance", Type: OptionSubcommand},
			{Name: "give", Description: "Give tokens to another user", Type: OptionSubcommand,
				Options: []OptionDef{
					{Name: "user", Description: "User to give tokens to", Type: OptionUser, Required: true},
					{Name: "amount", Description: "Number of tokens to give", Type: OptionInt, Required: true},
				},
			},
			{Name: "sell", Description: "Sell a character for 1 token", Type: OptionSubcommand,
				Options: []OptionDef{
					{Name: "id", Description: "ID of the character to sell", Type: OptionInt, Required: true, Autocomplete: true},
				},
			},
		},
	},
	{Name: "roll", Description: "Roll for a random character"},
	{Name: "search", Description: "Search AniList for anime, manga, characters, or users",
		Options: []OptionDef{
			{Name: "anime", Description: "Search for an anime by name", Type: OptionSubcommand,
				Options: []OptionDef{
					{Name: "name", Description: "name you wish to search", Type: OptionString, Required: true},
				},
			},
			{Name: "manga", Description: "Search for a manga by name", Type: OptionSubcommand,
				Options: []OptionDef{
					{Name: "name", Description: "name you wish to search", Type: OptionString, Required: true},
				},
			},
			{Name: "char", Description: "Search for a character by name", Type: OptionSubcommand,
				Options: []OptionDef{
					{Name: "name", Description: "name you wish to search", Type: OptionString, Required: true},
				},
			},
			{Name: "user", Description: "Search for a user by name", Type: OptionSubcommand,
				Options: []OptionDef{
					{Name: "name", Description: "name you wish to search", Type: OptionString, Required: true},
				},
			},
		},
	},
	{Name: "profile", Description: "View or edit user profiles",
		Options: []OptionDef{
			{Name: "view", Description: "View a user's profile", Type: OptionSubcommand,
				Options: []OptionDef{
					{Name: "user", Description: "User to view profile for (optional)", Type: OptionUser},
				},
			},
			{Name: "edit", Description: "Edit your profile", Type: OptionSubcommandGroup,
				Options: []OptionDef{
					{Name: "favorite", Description: "Set your favorite character", Type: OptionSubcommand,
						Options: []OptionDef{
							{Name: "id", Description: "ID of the character", Type: OptionInt, Required: true, Autocomplete: true},
						},
					},
					{Name: "quote", Description: "Set your profile quote", Type: OptionSubcommand,
						Options: []OptionDef{
							{Name: "value", Description: "Quote to set", Type: OptionString, Required: true},
						},
					},
					{Name: "anilist", Description: "Set your AniList profile URL", Type: OptionSubcommand,
						Options: []OptionDef{
							{Name: "url", Description: "AniList URL (e.g., https://anilist.co/user/Username)", Type: OptionString, Required: true},
						},
					},
				},
			},
		},
	},
	{Name: "give", Description: "Give a character to another user",
		Options: []OptionDef{
			{Name: "user", Description: "User to give the character to", Type: OptionUser, Required: true},
			{Name: "id", Description: "ID of the character to give", Type: OptionInt, Required: true, Autocomplete: true},
		},
	},
	{Name: "claim", Description: "Claim a character by name",
		Options: []OptionDef{
			{Name: "name", Description: "Name of the character to claim", Type: OptionString, Required: true},
		},
	},
	{Name: "info", Description: "Get information about the bot"},
	{Name: "holders", Description: "List users in this server who own a specific character",
		Options: []OptionDef{
			{Name: "id", Description: "ID of the character", Type: OptionInt, Required: true, Autocomplete: true},
		},
	},
	{Name: "wishlist", Description: "Manage your character wishlist",
		Options: []OptionDef{
			{Name: "character", Description: "Manage characters in your wishlist", Type: OptionSubcommandGroup,
				Options: []OptionDef{
					{Name: "add", Description: "Add a character to your wishlist", Type: OptionSubcommand,
						Options: []OptionDef{
							{Name: "character", Description: "ID of the character to add", Type: OptionInt, Required: true, Autocomplete: true},
						},
					},
					{Name: "remove", Description: "Remove a character from your wishlist", Type: OptionSubcommand,
						Options: []OptionDef{
							{Name: "character", Description: "ID of the character to remove", Type: OptionInt, Required: true, Autocomplete: true},
						},
					},
					{Name: "list", Description: "List all characters in your wishlist", Type: OptionSubcommand},
					{Name: "remove-all", Description: "Remove all characters from your wishlist", Type: OptionSubcommand},
				},
			},
			{Name: "media", Description: "Manage media in your wishlist", Type: OptionSubcommandGroup,
				Options: []OptionDef{
					{Name: "add", Description: "Add all characters from a show/manga to your wishlist", Type: OptionSubcommand,
						Options: []OptionDef{
							{Name: "media", Description: "ID of the media to add", Type: OptionInt, Required: true, Autocomplete: true},
						},
					},
				},
			},
			{Name: "holders", Description: "Show users who have characters from your wishlist", Type: OptionSubcommand},
			{Name: "wanted", Description: "Show users who want characters you own", Type: OptionSubcommand},
			{Name: "compare", Description: "Compare your wishlist with another user's collection", Type: OptionSubcommand,
				Options: []OptionDef{
					{Name: "user", Description: "User to compare with", Type: OptionUser, Required: true},
				},
			},
		},
	},
}
