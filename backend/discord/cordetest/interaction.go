package cordetest

import (
	"encoding/json"

	"github.com/Karitham/corde"
)

// SlashCommandInteraction creates a slash command interaction with the given options.
// It sets reasonable defaults for Member, GuildID, ChannelID, and other fields.
func SlashCommandInteraction(
	userID uint64,
	guildID uint64,
	channelID uint64,
	username string,
	opts corde.OptionsInteractions,
) *corde.Interaction[corde.SlashCommandInteractionData] {
	raw := map[string]any{
		"id":   1,
		"name": "test",
		"type": 1,
	}
	if len(opts) > 0 {
		raw["options"] = opts
	}
	dataJSON, _ := json.Marshal(raw)
	var data corde.SlashCommandInteractionData
	_ = json.Unmarshal(dataJSON, &data)

	return &corde.Interaction[corde.SlashCommandInteractionData]{
		ID:            corde.Snowflake(userID),
		ApplicationID: 1,
		Type:          corde.INTERACTION_TYPE_APPLICATION_COMMAND,
		GuildID:       corde.Snowflake(guildID),
		ChannelID:     corde.Snowflake(channelID),
		Member: corde.Member{
			User: corde.User{
				ID:       corde.Snowflake(userID),
				Username: username,
			},
		},
		InnerInteractionType: corde.SlashCommandInteraction,
		Data:                 data,
	}
}

// AutocompleteInteraction creates an autocomplete interaction with the given options.
// It sets reasonable defaults for Member, GuildID, ChannelID, and other fields.
func AutocompleteInteraction(
	userID uint64,
	guildID uint64,
	channelID uint64,
	username string,
	opts corde.OptionsInteractions,
) *corde.Interaction[corde.AutocompleteInteractionData] {
	return &corde.Interaction[corde.AutocompleteInteractionData]{
		ID:            corde.Snowflake(userID),
		ApplicationID: 1,
		Type:          corde.INTERACTION_TYPE_APPLICATION_COMMAND_AUTOCOMPLETE,
		GuildID:       corde.Snowflake(guildID),
		ChannelID:     corde.Snowflake(channelID),
		Member: corde.Member{
			User: corde.User{
				ID:       corde.Snowflake(userID),
				Username: username,
			},
		},
		InnerInteractionType: corde.AutocompleteInteraction,
		Data: corde.AutocompleteInteractionData{
			ID:      1,
			Name:    "test",
			Type:    1,
			Options: opts,
		},
	}
}
