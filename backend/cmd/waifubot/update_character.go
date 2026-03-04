package main

import (
	"fmt"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/karitham/waifubot/anilist"
	"github.com/karitham/waifubot/storage"
	"github.com/karitham/waifubot/storage/collectionstore"
)

var UpdateCharacterCommand = &cli.Command{
	Name:    "update-character",
	Usage:   "Update the character in the database",
	Aliases: []string{"uc"},
	Flags: []cli.Flag{
		dbURLFlag,
	},
	Action: func(c *cli.Context) error {
		a := c.Args()
		if a.Len() < 1 {
			return fmt.Errorf("no character name provided")
		}

		s, err := storage.NewStore(c.Context, c.String(dbURLFlag.Name))
		if err != nil {
			return fmt.Errorf("error connecting to db %v", err)
		}

		char, err := anilist.New(anilist.NoCache).Character(c.Context, c.Args().First())
		if err != nil {
			return err
		}
		if len(char) < 1 {
			return fmt.Errorf("character not found")
		}

		if _, err := s.CollectionStore().UpdateImageName(c.Context, collectionstore.UpdateImageNameParams{
			Image: char[0].ImageURL,
			Name:  strings.Join(strings.Fields(char[0].Name), " "),
			ID:    char[0].ID,
		}); err != nil {
			return fmt.Errorf("error updating db %v", err)
		}

		return nil
	},
}
