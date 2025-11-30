package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/karitham/waifubot/anilist"
)

var SearchMangaCommand = &cli.Command{
	Name:  "search/manga",
	Usage: "Search for a manga by name",
	Flags: []cli.Flag{
		nameFlag,
		anilistMaxCharsFlag,
	},
	Action: func(c *cli.Context) error {
		name := c.String(nameFlag.Name)
		maxChars := c.Int64(anilistMaxCharsFlag.Name)

		ctx := c.Context
		animeService := anilist.New(anilist.MaxChar(maxChars))
		media, err := animeService.Manga(ctx, name)
		if err != nil {
			return fmt.Errorf("error searching manga: %w", err)
		}

		return json.NewEncoder(os.Stdout).Encode(media)
	},
}
