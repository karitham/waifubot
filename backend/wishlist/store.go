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

func (s *store) AddCharacterToWishlist(ctx context.Context, userID uint64, characterID int64) error {
	return s.q.AddCharacterToWishlist(ctx, wishliststore.AddCharacterToWishlistParams{
		UserID:      userID,
		CharacterID: characterID,
	})
}

func (s *store) AddMultipleCharactersToWishlist(ctx context.Context, userID uint64, characterIDs []int64) error {
	return s.q.AddMultipleCharactersToWishlist(ctx, wishliststore.AddMultipleCharactersToWishlistParams{
		UserID:  userID,
		Column2: characterIDs,
	})
}

func (s *store) RemoveCharacterFromWishlist(ctx context.Context, userID uint64, characterID int64) error {
	return s.q.RemoveCharacterFromWishlist(ctx, wishliststore.RemoveCharacterFromWishlistParams{
		UserID:      userID,
		CharacterID: characterID,
	})
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

func (s *store) GetWishlistHolders(ctx context.Context, userID uint64, guildID uint64) ([]WishlistHolder, error) {
	// First get the user's wishlist character IDs
	wishlist, err := s.GetUserCharacterWishlist(ctx, userID)
	if err != nil {
		return nil, err
	}

	if len(wishlist) == 0 {
		return []WishlistHolder{}, nil
	}

	characterIDs := make([]int64, len(wishlist))
	for i, c := range wishlist {
		characterIDs[i] = c.ID
	}

	rows, err := s.q.GetWishlistHolders(ctx, wishliststore.GetWishlistHoldersParams{
		Column1: characterIDs,
		UserID:  userID,
		GuildID: guildID,
	})
	if err != nil {
		return nil, err
	}

	// Group by user
	holdersMap := make(map[uint64]*WishlistHolder)
	for _, row := range rows {
		if _, exists := holdersMap[row.UserID]; !exists {
			holdersMap[row.UserID] = &WishlistHolder{
				UserID:     row.UserID,
				Characters: []Character{},
			}
		}
		holdersMap[row.UserID].Characters = append(holdersMap[row.UserID].Characters, Character{
			ID:    row.CharacterID,
			Name:  row.CharacterName,
			Image: row.CharacterImage,
		})
	}

	holders := make([]WishlistHolder, 0, len(holdersMap))
	for _, h := range holdersMap {
		holders = append(holders, *h)
	}

	return holders, nil
}

func (s *store) GetWantedCharacters(ctx context.Context, userID uint64) ([]WantedCharacter, error) {
	rows, err := s.q.GetWantedCharacters(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Group by user
	wantedMap := make(map[uint64]*WantedCharacter)
	for _, row := range rows {
		if _, exists := wantedMap[row.UserID]; !exists {
			wantedMap[row.UserID] = &WantedCharacter{
				UserID:     row.UserID,
				Characters: []Character{},
			}
		}
		wantedMap[row.UserID].Characters = append(wantedMap[row.UserID].Characters, Character{
			ID:    row.CharacterID,
			Name:  row.CharacterName,
			Image: row.CharacterImage,
		})
	}

	wanted := make([]WantedCharacter, 0, len(wantedMap))
	for _, w := range wantedMap {
		wanted = append(wanted, *w)
	}

	return wanted, nil
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
	mutual := 0

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
		case "wants":
			userWants = append(userWants, char)
		}

		// Count mutual matches (characters both want and the other has)
		if row.Type == "has" {
			// Check if userID2 wants this character
			for _, r2 := range rows {
				if r2.Type == "wants" && r2.ID == row.ID {
					mutual++
					break
				}
			}
		}
	}

	return WishlistComparison{
		UserHasCharacters:   userHas,
		UserWantsCharacters: userWants,
		MutualMatches:       mutual,
	}, nil
}
