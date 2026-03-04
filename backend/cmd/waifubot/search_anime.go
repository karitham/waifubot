package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/karitham/waifubot/anilist"
)

var SearchAnimeCommand = &cli.Command{
	Name:  "search/anime",
	Usage: "Search for an anime by name",
	Flags: []cli.Flag{
		nameFlag,
		anilistMaxCharsFlag,
	},
	Action: func(c *cli.Context) error {
		name := c.String(nameFlag.Name)
		maxChars := c.Int64(anilistMaxCharsFlag.Name)

		ctx := c.Context
		animeService := anilist.New(anilist.MaxChar(maxChars))
		media, err := animeService.Anime(ctx, name)
		if err != nil {
			return fmt.Errorf("error searching anime: %w", err)
		}

		return json.NewEncoder(os.Stdout).Encode(media)
	},
}
