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
	},
	Action: func(c *cli.Context) error {
		name := c.String(nameFlag.Name)

		ctx := c.Context
		animeService := anilist.New()
		media, err := animeService.Anime(ctx, name)
		if err != nil {
			return fmt.Errorf("error searching anime: %w", err)
		}

		return json.NewEncoder(os.Stdout).Encode(media)
	},
}
