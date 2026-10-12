package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Karitham/corde"
	"github.com/urfave/cli/v2"

	"github.com/karitham/waifubot/catalog"
	"github.com/karitham/waifubot/collection"
	"github.com/karitham/waifubot/storage"
	"github.com/karitham/waifubot/storage/wishlistpg"
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

				err = collection.AddToWishlist(ctx, newCollectionStore(store), collection.UserID(userID), []catalog.CharacterID{catalog.CharacterID(charID)})
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

				err = newCollectionStore(store).RemoveFromWishlist(ctx, collection.UserID(userID), catalog.CharacterID(charID))
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

				wishlistStore := wishlistpg.New(store.WishlistStore())
				chars, err := wishlistStore.GetUserCharacterWishlist(ctx, collection.UserID(userID))
				if err != nil {
					return fmt.Errorf("error getting user wishlist: %w", err)
				}

				views := wishlistEntryViews(chars)

				result := map[string]any{
					"user_id":  userID.String(),
					"wishlist": views,
					"total":    len(views),
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

				wishlistStore := wishlistpg.New(store.WishlistStore())
				wishlist, err := wishlistStore.GetUserCharacterWishlist(ctx, collection.UserID(userID))
				if err != nil {
					return fmt.Errorf("error getting wishlist: %w", err)
				}

				characterIDs := make([]catalog.CharacterID, len(wishlist))
				for i, c := range wishlist {
					characterIDs[i] = c.ID
				}

				holders, err := wishlistStore.GetWishlistHolders(ctx, characterIDs, collection.UserID(userID), 0)
				if err != nil {
					return fmt.Errorf("error getting wishlist holders: %w", err)
				}

				views := userCharacterSetViews(holders)

				result := map[string]any{
					"user_id": userID.String(),
					"holders": views,
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

				wishlistStore := wishlistpg.New(store.WishlistStore())
				wanted, err := wishlistStore.GetWantedCharacters(ctx, collection.UserID(userID), 0)
				if err != nil {
					return fmt.Errorf("error getting wanted characters: %w", err)
				}

				views := userCharacterSetViews(wanted)

				result := map[string]any{
					"user_id": userID.String(),
					"wanted":  views,
				}

				return json.NewEncoder(os.Stdout).Encode(result)
			},
		},
		{
			Name:  "remove-all",
			Usage: "Remove all characters from a user's wishlist",
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

				wishlistStore := wishlistpg.New(store.WishlistStore())
				err = wishlistStore.RemoveAllFromWishlist(ctx, collection.UserID(userID))
				if err != nil {
					return fmt.Errorf("error removing all from wishlist: %w", err)
				}

				result := map[string]any{
					"user_id": userID.String(),
					"action":  "removed_all",
				}

				return json.NewEncoder(os.Stdout).Encode(result)
			},
		},
	},
}
