package pg

import (
	"context"

	"github.com/karitham/waifubot/catalog"
	"github.com/karitham/waifubot/collection"
	"github.com/karitham/waifubot/guild"
	"github.com/karitham/waifubot/storage"
)

func storageIDs(ids []catalog.CharacterID) []int64 {
	out := make([]int64, len(ids))
	for i, id := range ids {
		out[i] = int64(id)
	}
	return out
}

func (b *base) GetUserCharacterWishlist(ctx context.Context, userID collection.UserID) ([]collection.WishlistEntry, error) {
	rows, err := b.q.GetUserCharacterWishlist(ctx, uint64(userID))
	if err != nil {
		return nil, err
	}

	entries := make([]collection.WishlistEntry, len(rows))
	for i, row := range rows {
		entries[i] = collection.WishlistEntry{
			Character: catalog.Character{
				ID:         catalog.CharacterID(row.ID),
				Name:       row.Name,
				Image:      row.Image,
				Popularity: int(row.Favorites),
			},
			AddedAt: row.Date.Time,
		}
	}

	return entries, nil
}

func (b *base) GetWishlistHolders(ctx context.Context, ids []catalog.CharacterID, userID collection.UserID, guildID guild.ID) ([]collection.WishlistUserSet, error) {
	if len(ids) == 0 {
		return []collection.WishlistUserSet{}, nil
	}

	rows, err := b.q.GetWishlistHolders(ctx, storage.GetWishlistHoldersParams{
		Column1: storageIDs(ids),
		UserID:  uint64(userID),
		GuildID: uint64(guildID),
	})
	if err != nil {
		return nil, err
	}

	out := make([]wishlistRow, len(rows))
	for i, r := range rows {
		out[i] = wishlistRow{userID: r.UserID, characterID: r.CharacterID, name: r.CharacterName, image: r.CharacterImage}
	}
	return groupByUser(out), nil
}

func (b *base) GetWantedCharacters(ctx context.Context, userID collection.UserID, guildID guild.ID) ([]collection.WishlistUserSet, error) {
	rows, err := b.q.GetWantedCharacters(ctx, storage.GetWantedCharactersParams{
		UserID:  uint64(userID),
		GuildID: uint64(guildID),
	})
	if err != nil {
		return nil, err
	}

	out := make([]wishlistRow, len(rows))
	for i, r := range rows {
		out[i] = wishlistRow{userID: r.UserID, characterID: r.CharacterID, name: r.CharacterName, image: r.CharacterImage}
	}
	return groupByUser(out), nil
}

// wishlistRow is the shape both discovery queries return.
type wishlistRow struct {
	userID      int64
	characterID int64
	name        string
	image       string
}

func groupByUser(rows []wishlistRow) []collection.WishlistUserSet {
	m := make(map[collection.UserID]*collection.WishlistUserSet)
	for _, r := range rows {
		id := collection.UserID(r.userID)
		if _, ok := m[id]; !ok {
			m[id] = &collection.WishlistUserSet{UserID: id}
		}
		m[id].Characters = append(m[id].Characters, collection.WishlistEntry{Character: catalog.Character{
			ID:    catalog.CharacterID(r.characterID),
			Name:  r.name,
			Image: r.image,
		}})
	}
	result := make([]collection.WishlistUserSet, 0, len(m))
	for _, v := range m {
		result = append(result, *v)
	}
	return result
}

func (b *base) CompareWithUser(ctx context.Context, userID1, userID2 collection.UserID) (collection.WishlistComparison, error) {
	rows, err := b.q.CompareWithUser(ctx, storage.CompareWithUserParams{
		UserID:   uint64(userID1),
		UserID_2: uint64(userID2),
	})
	if err != nil {
		return collection.WishlistComparison{}, err
	}

	var userHas, userWants []collection.WishlistEntry
	hasIDs := make(map[catalog.CharacterID]bool)
	wantsIDs := make(map[catalog.CharacterID]bool)

	for _, row := range rows {
		entry := collection.WishlistEntry{
			Character: catalog.Character{
				ID:    catalog.CharacterID(row.ID),
				Name:  row.Name,
				Image: row.Image,
			},
			AddedAt: row.Date.Time,
		}

		switch row.Type {
		case "has":
			userHas = append(userHas, entry)
			hasIDs[entry.ID] = true
		case "wants":
			userWants = append(userWants, entry)
			wantsIDs[entry.ID] = true
		}
	}

	mutual := 0
	for id := range hasIDs {
		if wantsIDs[id] {
			mutual++
		}
	}

	return collection.WishlistComparison{
		UserHasCharacters:   userHas,
		UserWantsCharacters: userWants,
		MutualMatches:       mutual,
	}, nil
}

func (b *base) GetUsersWantingCharacter(ctx context.Context, charID catalog.CharacterID, guildID guild.ID, excludeUserID collection.UserID) ([]guild.MemberID, error) {
	ids, err := b.q.GetUsersWantingCharacter(ctx, storage.GetUsersWantingCharacterParams{
		CharacterID: int64(charID),
		GuildID:     uint64(guildID),
		UserID:      uint64(excludeUserID),
	})
	if err != nil {
		return nil, err
	}

	memberIDs := make([]guild.MemberID, len(ids))
	for i, id := range ids {
		memberIDs[i] = guild.MemberID(id)
	}
	return memberIDs, nil
}
