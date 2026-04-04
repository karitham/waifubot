package main

import (
	"github.com/karitham/waifubot/cmd/waifubot/flags"
)

// Re-export flags from shared package with original names for backwards compatibility
var (
	dbURLFlag           = flags.DbURLFlag
	userFlag            = flags.UserFlag
	guildIDFlag         = flags.GuildIDFlag
	appIDFlag           = flags.AppIDFlag
	charIDFlag          = flags.CharIDFlag
	botTokenFlag        = flags.BotTokenFlag
	rollCooldownFlag    = flags.RollCooldownFlag
	seriesRollCostFlag  = flags.SeriesRollCostFlag
	anilistMaxCharsFlag = flags.AnilistMaxCharsFlag
	nameFlag            = flags.NameFlag
	logLevelFlag        = flags.LogLevelFlag
	apiFlag             = flags.ApiFlag
)
