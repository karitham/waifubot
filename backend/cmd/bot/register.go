package main

import (
	"fmt"

	"github.com/Karitham/corde"
	"github.com/urfave/cli/v2"

	"github.com/karitham/waifubot/discord"
)

var RegisterCommand = &cli.Command{
	Name:    "register",
	Aliases: []string{"r"},
	Usage:   "Register the bot commands",
	Flags: []cli.Flag{
		botTokenFlag,
		&cli.Uint64Flag{
			Name:    "guild-id",
			EnvVars: []string{"GUILD_ID"},
		},
		appIDFlag,
	},
	Action: func(c *cli.Context) error {
		bot := &discord.Bot{
			AppID:    corde.SnowflakeFromString(c.String(appIDFlag.Name)),
			BotToken: c.String(botTokenFlag.Name),
		}

		if gid := c.Uint64(guildIDFlag.Name); gid != 0 {
			id := corde.Snowflake(gid)
			bot.GuildID = &id
		}

		if err := bot.RegisterCommands(); err != nil {
			return fmt.Errorf("error registering commands %v", err)
		}
		return nil
	},
}
