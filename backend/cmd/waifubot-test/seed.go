package main

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/karitham/waifubot/catalog"
	"github.com/karitham/waifubot/collection"
	"github.com/karitham/waifubot/collectionwire"
	"github.com/karitham/waifubot/storage"
	"github.com/karitham/waifubot/storage/guildstore"
)

// seedCharacter inserts a character into the characters table using raw SQL.
func seedCharacter(ctx context.Context, store *storage.DBStore, charID int64, charName, charImage string) error {
	_, err := store.DB().Exec(ctx, `
		INSERT INTO characters (id, name, image)
		VALUES ($1, $2, $3)
		ON CONFLICT DO NOTHING
	`, charID, charName, charImage)
	return err
}

// seedWishlistUsers adds characters to the wishlist for the specified user IDs.
func seedWishlistUsers(ctx context.Context, store *storage.DBStore, userIDs []uint64, charID int64) error {
	db := collectionwire.NewDB(store)
	for _, userID := range userIDs {
		_, err := collection.AddToWishlist(ctx, db, collection.UserID(userID), []catalog.CharacterID{catalog.CharacterID(charID)})
		if err != nil {
			return err
		}
	}
	return nil
}

// seedGuildMembers inserts guild members into the guild_members table.
func seedGuildMembers(ctx context.Context, store *storage.DBStore, guildID uint64, memberIDs []uint64) error {
	if guildID == 0 || len(memberIDs) == 0 {
		return nil
	}

	// Convert []uint64 to []int64 for the SQL query
	int64MemberIDs := make([]int64, len(memberIDs))
	for i, id := range memberIDs {
		int64MemberIDs[i] = int64(id)
	}

	return store.GuildStore().UpsertGuildMembers(ctx, guildstore.UpsertGuildMembersParams{
		GuildID:   guildID,
		Column2:   int64MemberIDs,
		IndexedAt: pgtype.Timestamp{Time: time.Now(), Valid: true},
	})
}
