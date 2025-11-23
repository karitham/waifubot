package collection

import (
	"context"
	"fmt"
	"time"

	"github.com/Karitham/corde"
	"github.com/jackc/pgx/v5/pgtype"

	characters "github.com/karitham/waifubot/storage/collectionstore"
	users "github.com/karitham/waifubot/storage/userstore"
	wishliststore "github.com/karitham/waifubot/storage/wishliststore"
)

type ErrRollCooldown struct {
	Until         time.Time
	MissingTokens int
}

func (e ErrRollCooldown) Error() string {
	return fmt.Sprintf("You need another %d tokens to roll, or you can wait %s until next free roll.", e.MissingTokens, time.Until(e.Until).Round(time.Second))
}

// Roll executes the roll logic for a user
func Roll(ctx context.Context, store Store, animeService AnimeService, config Config, userID corde.Snowflake) (MediaCharacter, error) {
	tx, err := store.Tx(ctx)
	if err != nil {
		return MediaCharacter{}, err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback(ctx)
		}
	}()

	user, err := tx.UserStore().Get(ctx, uint64(userID))
	if err != nil {
		return MediaCharacter{}, err
	}

	now := time.Now()
	canRollFree := now.After(user.Date.Time.Add(config.RollCooldown))
	hasTokens := user.Tokens >= config.TokensNeeded

	if !canRollFree && !hasTokens {
		return MediaCharacter{}, ErrRollCooldown{
			Until:         user.Date.Time.Add(config.RollCooldown),
			MissingTokens: int(config.TokensNeeded - user.Tokens),
		}
	}

	charsIDs, err := tx.CollectionStore().ListIDs(ctx, uint64(userID))
	if err != nil {
		return MediaCharacter{}, err
	}

	char, err := animeService.RandomChar(ctx, charsIDs...)
	if err != nil {
		return MediaCharacter{}, err
	}

	_, err = tx.CollectionStore().UpsertCharacter(ctx, characters.UpsertCharacterParams{
		ID:    char.ID,
		Name:  char.Name,
		Image: char.ImageURL,
	})
	if err != nil {
		return MediaCharacter{}, err
	}

	_, err = tx.CollectionStore().Insert(ctx, characters.InsertParams{
		UserID:      uint64(userID),
		CharacterID: char.ID,
		Source:      "ROLL",
		AcquiredAt:  pgtype.Timestamp{Time: now, Valid: true},
	})
	if err != nil {
		return MediaCharacter{}, err
	}

	// Remove from wishlist if present
	_ = tx.WishlistStore().RemoveCharacterFromWishlist(ctx, wishliststore.RemoveCharacterFromWishlistParams{
		UserID:      uint64(userID),
		CharacterID: char.ID,
	})

	if canRollFree {
		err = tx.UserStore().UpdateDate(ctx, users.UpdateDateParams{
			Date:   pgtype.Timestamp{Time: now, Valid: true},
			UserID: uint64(userID),
		})
	} else {
		_, err = tx.UserStore().ConsumeTokens(ctx, users.ConsumeTokensParams{
			Tokens: config.TokensNeeded,
			UserID: uint64(userID),
		})
	}

	if err != nil {
		return MediaCharacter{}, err
	}

	err = tx.Commit(ctx)
	committed = err == nil
	return char, err
}
