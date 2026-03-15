package wishlist

import (
	"context"
	"time"

	"github.com/karitham/waifubot/storage/wishliststore"
)

type store struct {
	q wishliststore.Querier
}

func New(q wishliststore.Querier) Store {
	return &store{q: q}
}

func (s *store) AddMultipleCharactersToWishlist(ctx context.Context, userID uint64, characterIDs []int64) error {
	return s.q.AddMultipleCharactersToWishlist(ctx, wishliststore.AddMultipleCharactersToWishlistParams{
		UserID:  userID,
		Column2: characterIDs,
	})
}

func (s *store) RemoveMultipleCharactersFromWishlist(ctx context.Context, userID uint64, characterIDs []int64) error {
	return s.q.RemoveMultipleCharactersFromWishlist(ctx, wishliststore.RemoveMultipleCharactersFromWishlistParams{
		UserID:  userID,
		Column2: characterIDs,
	})
}

func (s *store) RemoveAllFromWishlist(ctx context.Context, userID uint64) error {
	return s.q.RemoveAllFromWishlist(ctx, userID)
}

func (s *store) GetUserCharacterWishlist(ctx context.Context, userID uint64) ([]Character, error) {
	rows, err := s.q.GetUserCharacterWishlist(ctx, userID)
	if err != nil {
		return nil, err
	}

	characters := make([]Character, len(rows))
	for i, row := range rows {
		characters[i] = Character{
			ID:    row.ID,
			Name:  row.Name,
			Image: row.Image,
			Date:  row.Date.Time.Format(time.RFC3339),
		}
	}

	return characters, nil
}

func (s *store) GetWishlistHolders(ctx context.Context, characterIDs []int64, userID, guildID uint64) ([]UserCharacterSet, error) {
	if len(characterIDs) == 0 {
		return []UserCharacterSet{}, nil
	}

	rows, err := s.q.GetWishlistHolders(ctx, wishliststore.GetWishlistHoldersParams{
		Column1: characterIDs,
		UserID:  userID,
		GuildID: guildID,
	})
	if err != nil {
		return nil, err
	}

	return groupByUser(rows), nil
}

func (s *store) GetWantedCharacters(ctx context.Context, userID, guildID uint64) ([]UserCharacterSet, error) {
	rows, err := s.q.GetWantedCharacters(ctx, wishliststore.GetWantedCharactersParams{
		UserID:  userID,
		GuildID: guildID,
	})
	if err != nil {
		return nil, err
	}

	return groupWantedByUser(rows), nil
}

func groupByUser(rows []wishliststore.GetWishlistHoldersRow) []UserCharacterSet {
	m := make(map[uint64]*UserCharacterSet)
	for _, r := range rows {
		id := uint64(r.UserID)
		if _, ok := m[id]; !ok {
			m[id] = &UserCharacterSet{UserID: id}
		}
		m[id].Characters = append(m[id].Characters, Character{
			ID:    r.CharacterID,
			Name:  r.CharacterName,
			Image: r.CharacterImage,
		})
	}
	result := make([]UserCharacterSet, 0, len(m))
	for _, v := range m {
		result = append(result, *v)
	}
	return result
}

func groupWantedByUser(rows []wishliststore.GetWantedCharactersRow) []UserCharacterSet {
	m := make(map[uint64]*UserCharacterSet)
	for _, r := range rows {
		id := uint64(r.UserID)
		if _, ok := m[id]; !ok {
			m[id] = &UserCharacterSet{UserID: id}
		}
		m[id].Characters = append(m[id].Characters, Character{
			ID:    r.CharacterID,
			Name:  r.CharacterName,
			Image: r.CharacterImage,
		})
	}
	result := make([]UserCharacterSet, 0, len(m))
	for _, v := range m {
		result = append(result, *v)
	}
	return result
}

func (s *store) CompareWithUser(ctx context.Context, userID1, userID2 uint64) (WishlistComparison, error) {
	rows, err := s.q.CompareWithUser(ctx, wishliststore.CompareWithUserParams{
		UserID:   userID1,
		UserID_2: userID2,
	})
	if err != nil {
		return WishlistComparison{}, err
	}

	var userHas, userWants []Character
	hasIDs := make(map[int64]bool)
	wantsIDs := make(map[int64]bool)

	for _, row := range rows {
		char := Character{
			ID:    row.ID,
			Name:  row.Name,
			Image: row.Image,
			Date:  row.Date.Time.Format(time.RFC3339),
		}

		switch row.Type {
		case "has":
			userHas = append(userHas, char)
			hasIDs[row.ID] = true
		case "wants":
			userWants = append(userWants, char)
			wantsIDs[row.ID] = true
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
