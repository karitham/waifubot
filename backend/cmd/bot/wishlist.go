package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Karitham/corde"
	"github.com/urfave/cli/v2"

	"github.com/karitham/waifubot/storage"
	"github.com/karitham/waifubot/wishlist"
)

var WishlistCommand = &cli.Command{
	Name:  "wishlist",
	Usage: "Manage user wishlists",
	Subcommands: []*cli.Command{
		{
			Name:  "add",
			Usage: "Add a character to a user's wishlist",
			Flags: []cli.Flag{
				userFlag,
				charIDFlag,
				dbURLFlag,
			},
			Action: func(c *cli.Context) error {
				userID := corde.SnowflakeFromString(c.String(userFlag.Name))
				if userID == 0 {
					return fmt.Errorf("invalid user ID: %s", c.String(userFlag.Name))
				}

				charID := c.Int64(charIDFlag.Name)
				dbURL := c.String(dbURLFlag.Name)

				ctx := c.Context
				store, err := storage.NewStore(ctx, dbURL)
				if err != nil {
					return fmt.Errorf("error connecting to db: %w", err)
				}

				err = wishlist.AddCharacter(ctx, wishlist.New(store.WishlistStore()), uint64(userID), charID)
				if err != nil {
					return fmt.Errorf("error adding character to wishlist: %w", err)
				}

				result := map[string]any{
					"user_id":      userID.String(),
					"character_id": charID,
					"action":       "added",
				}

				return json.NewEncoder(os.Stdout).Encode(result)
			},
		},
		{
			Name:  "remove",
			Usage: "Remove a character from a user's wishlist",
			Flags: []cli.Flag{
				userFlag,
				charIDFlag,
				dbURLFlag,
			},
			Action: func(c *cli.Context) error {
				userID := corde.SnowflakeFromString(c.String(userFlag.Name))
				if userID == 0 {
					return fmt.Errorf("invalid user ID: %s", c.String(userFlag.Name))
				}

				charID := c.Int64(charIDFlag.Name)
				dbURL := c.String(dbURLFlag.Name)

				ctx := c.Context
				store, err := storage.NewStore(ctx, dbURL)
				if err != nil {
					return fmt.Errorf("error connecting to db: %w", err)
				}

				wishlistStore := wishlist.New(store.WishlistStore())
				err = wishlist.RemoveCharacter(ctx, wishlistStore, uint64(userID), charID)
				if err != nil {
					return fmt.Errorf("error removing character from wishlist: %w", err)
				}

				result := map[string]any{
					"user_id":      userID.String(),
					"character_id": charID,
					"action":       "removed",
				}

				return json.NewEncoder(os.Stdout).Encode(result)
			},
		},
		{
			Name:  "list",
			Usage: "List a user's wishlist",
			Flags: []cli.Flag{
				userFlag,
				dbURLFlag,
			},
			Action: func(c *cli.Context) error {
				userID := corde.SnowflakeFromString(c.String(userFlag.Name))
				if userID == 0 {
					return fmt.Errorf("invalid user ID: %s", c.String(userFlag.Name))
				}

				dbURL := c.String(dbURLFlag.Name)

				ctx := c.Context
				store, err := storage.NewStore(ctx, dbURL)
				if err != nil {
					return fmt.Errorf("error connecting to db: %w", err)
				}

				wishlistStore := wishlist.New(store.WishlistStore())
				chars, err := wishlist.GetUserWishlist(ctx, wishlistStore, uint64(userID))
				if err != nil {
					return fmt.Errorf("error getting user wishlist: %w", err)
				}

				result := map[string]any{
					"user_id":  userID.String(),
					"wishlist": chars,
					"total":    len(chars),
				}

				return json.NewEncoder(os.Stdout).Encode(result)
			},
		},
		{
			Name:  "holders",
			Usage: "Show who has characters from a user's wishlist",
			Flags: []cli.Flag{
				userFlag,
				dbURLFlag,
			},
			Action: func(c *cli.Context) error {
				userID := corde.SnowflakeFromString(c.String(userFlag.Name))
				if userID == 0 {
					return fmt.Errorf("invalid user ID: %s", c.String(userFlag.Name))
				}

				dbURL := c.String(dbURLFlag.Name)

				ctx := c.Context
				store, err := storage.NewStore(ctx, dbURL)
				if err != nil {
					return fmt.Errorf("error connecting to db: %w", err)
				}

				wishlistStore := wishlist.New(store.WishlistStore())
				holders, err := wishlist.GetWishlistHolders(ctx, wishlistStore, uint64(userID), 0)
				if err != nil {
					return fmt.Errorf("error getting wishlist holders: %w", err)
				}

				result := map[string]any{
					"user_id": userID.String(),
					"holders": holders,
				}

				return json.NewEncoder(os.Stdout).Encode(result)
			},
		},
		{
			Name:  "wanted",
			Usage: "Show who wants characters from a user's collection",
			Flags: []cli.Flag{
				userFlag,
				dbURLFlag,
			},
			Action: func(c *cli.Context) error {
				userID := corde.SnowflakeFromString(c.String(userFlag.Name))
				if userID == 0 {
					return fmt.Errorf("invalid user ID: %s", c.String(userFlag.Name))
				}

				dbURL := c.String(dbURLFlag.Name)

				ctx := c.Context
				store, err := storage.NewStore(ctx, dbURL)
				if err != nil {
					return fmt.Errorf("error connecting to db: %w", err)
				}

				wishlistStore := wishlist.New(store.WishlistStore())
				wanted, err := wishlist.GetWantedCharacters(ctx, wishlistStore, uint64(userID), 0)
				if err != nil {
					return fmt.Errorf("error getting wanted characters: %w", err)
				}

				result := map[string]any{
					"user_id": userID.String(),
					"wanted":  wanted,
				}

				return json.NewEncoder(os.Stdout).Encode(result)
			},
		},
	},
}
