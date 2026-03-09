package interfaces

import (
	"context"
)

type WishlistComparison struct {
	UserHas       []int64
	UserWants     []int64
	MutualMatches []int64
}

type WishlistRepository interface {
	AddCharacter(ctx context.Context, userID uint64, charID int64) error
	AddMultipleCharacters(ctx context.Context, userID uint64, charIDs []int64) error
	RemoveCharacter(ctx context.Context, userID uint64, charID int64) error
	RemoveMultipleCharacters(ctx context.Context, userID uint64, charIDs []int64) error
	RemoveAll(ctx context.Context, userID uint64) error
	GetUserWishlist(ctx context.Context, userID uint64) ([]int64, error)
	GetWantedCharacters(ctx context.Context, charID int64) ([]uint64, error)
	GetWishlistHolders(ctx context.Context, charIDs []int64) (map[int64][]uint64, error)
	CompareWithUser(ctx context.Context, userID uint64, otherUserID uint64) (WishlistComparison, error)
}
