package wishlist

import (
	"context"

	"github.com/karitham/waifubot/catalog"
	"github.com/karitham/waifubot/collection"
	"github.com/karitham/waifubot/guild"
	"github.com/karitham/waifubot/storage/wishliststore"
)

type store struct {
	q wishliststore.Querier
}

func New(q wishliststore.Querier) Store {
	return &store{q: q}
}

func storageIDs(ids []catalog.CharacterID) []int64 {
	out := make([]int64, len(ids))
	for i, id := range ids {
		out[i] = int64(id)
	}
	return out
}

func (s *store) RemoveAllFromWishlist(ctx context.Context, userID collection.UserID) error {
	return s.q.RemoveAllFromWishlist(ctx, uint64(userID))
}

func (s *store) GetUserCharacterWishlist(ctx context.Context, userID collection.UserID) ([]Entry, error) {
	rows, err := s.q.GetUserCharacterWishlist(ctx, uint64(userID))
	if err != nil {
		return nil, err
	}

	entries := make([]Entry, len(rows))
	for i, row := range rows {
		entries[i] = Entry{
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

func (s *store) GetWishlistHolders(ctx context.Context, ids []catalog.CharacterID, userID collection.UserID, guildID guild.ID) ([]UserCharacterSet, error) {
	if len(ids) == 0 {
		return []UserCharacterSet{}, nil
	}

	rows, err := s.q.GetWishlistHolders(ctx, wishliststore.GetWishlistHoldersParams{
		Column1: storageIDs(ids),
		UserID:  uint64(userID),
		GuildID: uint64(guildID),
	})
	if err != nil {
		return nil, err
	}

	return groupByUser(rows), nil
}

func (s *store) GetWantedCharacters(ctx context.Context, userID collection.UserID, guildID guild.ID) ([]UserCharacterSet, error) {
	rows, err := s.q.GetWantedCharacters(ctx, wishliststore.GetWantedCharactersParams{
		UserID:  uint64(userID),
		GuildID: uint64(guildID),
	})
	if err != nil {
		return nil, err
	}

	return groupByUserWanted(rows), nil
}

func groupByUser(rows []wishliststore.GetWishlistHoldersRow) []UserCharacterSet {
	m := make(map[collection.UserID]*UserCharacterSet)
	for _, r := range rows {
		id := collection.UserID(r.UserID)
		if _, ok := m[id]; !ok {
			m[id] = &UserCharacterSet{UserID: id}
		}
		m[id].Characters = append(m[id].Characters, Entry{Character: catalog.Character{
			ID:    catalog.CharacterID(r.CharacterID),
			Name:  r.CharacterName,
			Image: r.CharacterImage,
		}})
	}
	result := make([]UserCharacterSet, 0, len(m))
	for _, v := range m {
		result = append(result, *v)
	}
	return result
}

func groupByUserWanted(rows []wishliststore.GetWantedCharactersRow) []UserCharacterSet {
	m := make(map[collection.UserID]*UserCharacterSet)
	for _, r := range rows {
		id := collection.UserID(r.UserID)
		if _, ok := m[id]; !ok {
			m[id] = &UserCharacterSet{UserID: id}
		}
		m[id].Characters = append(m[id].Characters, Entry{Character: catalog.Character{
			ID:    catalog.CharacterID(r.CharacterID),
			Name:  r.CharacterName,
			Image: r.CharacterImage,
		}})
	}
	result := make([]UserCharacterSet, 0, len(m))
	for _, v := range m {
		result = append(result, *v)
	}
	return result
}

func (s *store) CompareWithUser(ctx context.Context, userID1, userID2 collection.UserID) (WishlistComparison, error) {
	rows, err := s.q.CompareWithUser(ctx, wishliststore.CompareWithUserParams{
		UserID:   uint64(userID1),
		UserID_2: uint64(userID2),
	})
	if err != nil {
		return WishlistComparison{}, err
	}

	var userHas, userWants []Entry
	hasIDs := make(map[catalog.CharacterID]bool)
	wantsIDs := make(map[catalog.CharacterID]bool)

	for _, row := range rows {
		entry := Entry{
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

	return WishlistComparison{
		UserHasCharacters:   userHas,
		UserWantsCharacters: userWants,
		MutualMatches:       mutual,
	}, nil
}

func (s *store) GetUsersWantingCharacter(ctx context.Context, charID catalog.CharacterID, guildID guild.ID, excludeUserID collection.UserID) ([]guild.MemberID, error) {
	ids, err := s.q.GetUsersWantingCharacter(ctx, wishliststore.GetUsersWantingCharacterParams{
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
