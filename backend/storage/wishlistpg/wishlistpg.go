package wishlistpg

import (
	"context"

	"github.com/karitham/waifubot/catalog"
	"github.com/karitham/waifubot/collection"
	"github.com/karitham/waifubot/guild"
	"github.com/karitham/waifubot/storage"
	"github.com/karitham/waifubot/wishlist"
)

// Pg implements wishlist.Store over the generated wishlist queries.
type Pg struct {
	Q storage.Querier
}

// New creates the wishlist store.
func New(q storage.Querier) *Pg {
	return &Pg{Q: q}
}

func storageIDs(ids []catalog.CharacterID) []int64 {
	out := make([]int64, len(ids))
	for i, id := range ids {
		out[i] = int64(id)
	}
	return out
}

func (p *Pg) GetUserCharacterWishlist(ctx context.Context, userID collection.UserID) ([]wishlist.Entry, error) {
	rows, err := p.Q.GetUserCharacterWishlist(ctx, uint64(userID))
	if err != nil {
		return nil, err
	}

	entries := make([]wishlist.Entry, len(rows))
	for i, row := range rows {
		entries[i] = wishlist.Entry{
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

func (p *Pg) GetWishlistHolders(ctx context.Context, ids []catalog.CharacterID, userID collection.UserID, guildID guild.ID) ([]wishlist.UserCharacterSet, error) {
	if len(ids) == 0 {
		return []wishlist.UserCharacterSet{}, nil
	}

	rows, err := p.Q.GetWishlistHolders(ctx, storage.GetWishlistHoldersParams{
		Column1: storageIDs(ids),
		UserID:  uint64(userID),
		GuildID: uint64(guildID),
	})
	if err != nil {
		return nil, err
	}

	return groupByUser(rows), nil
}

func (p *Pg) GetWantedCharacters(ctx context.Context, userID collection.UserID, guildID guild.ID) ([]wishlist.UserCharacterSet, error) {
	rows, err := p.Q.GetWantedCharacters(ctx, storage.GetWantedCharactersParams{
		UserID:  uint64(userID),
		GuildID: uint64(guildID),
	})
	if err != nil {
		return nil, err
	}

	return groupByUserWanted(rows), nil
}

func groupByUser(rows []storage.GetWishlistHoldersRow) []wishlist.UserCharacterSet {
	m := make(map[collection.UserID]*wishlist.UserCharacterSet)
	for _, r := range rows {
		id := collection.UserID(r.UserID)
		if _, ok := m[id]; !ok {
			m[id] = &wishlist.UserCharacterSet{UserID: id}
		}
		m[id].Characters = append(m[id].Characters, wishlist.Entry{Character: catalog.Character{
			ID:    catalog.CharacterID(r.CharacterID),
			Name:  r.CharacterName,
			Image: r.CharacterImage,
		}})
	}
	result := make([]wishlist.UserCharacterSet, 0, len(m))
	for _, v := range m {
		result = append(result, *v)
	}
	return result
}

func groupByUserWanted(rows []storage.GetWantedCharactersRow) []wishlist.UserCharacterSet {
	m := make(map[collection.UserID]*wishlist.UserCharacterSet)
	for _, r := range rows {
		id := collection.UserID(r.UserID)
		if _, ok := m[id]; !ok {
			m[id] = &wishlist.UserCharacterSet{UserID: id}
		}
		m[id].Characters = append(m[id].Characters, wishlist.Entry{Character: catalog.Character{
			ID:    catalog.CharacterID(r.CharacterID),
			Name:  r.CharacterName,
			Image: r.CharacterImage,
		}})
	}
	result := make([]wishlist.UserCharacterSet, 0, len(m))
	for _, v := range m {
		result = append(result, *v)
	}
	return result
}

func (p *Pg) CompareWithUser(ctx context.Context, userID1, userID2 collection.UserID) (wishlist.WishlistComparison, error) {
	rows, err := p.Q.CompareWithUser(ctx, storage.CompareWithUserParams{
		UserID:   uint64(userID1),
		UserID_2: uint64(userID2),
	})
	if err != nil {
		return wishlist.WishlistComparison{}, err
	}

	var userHas, userWants []wishlist.Entry
	hasIDs := make(map[catalog.CharacterID]bool)
	wantsIDs := make(map[catalog.CharacterID]bool)

	for _, row := range rows {
		entry := wishlist.Entry{
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

	return wishlist.WishlistComparison{
		UserHasCharacters:   userHas,
		UserWantsCharacters: userWants,
		MutualMatches:       mutual,
	}, nil
}

func (p *Pg) GetUsersWantingCharacter(ctx context.Context, charID catalog.CharacterID, guildID guild.ID, excludeUserID collection.UserID) ([]guild.MemberID, error) {
	ids, err := p.Q.GetUsersWantingCharacter(ctx, storage.GetUsersWantingCharacterParams{
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
