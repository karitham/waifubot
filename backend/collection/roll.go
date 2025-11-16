package collection

import (
	"context"
	"fmt"
	"time"

	"github.com/Karitham/corde"
	"github.com/jackc/pgx/v5/pgtype"

	characters "github.com/karitham/waifubot/storage/collectionstore"
	users "github.com/karitham/waifubot/storage/userstore"
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
	txI, err := store.Tx(ctx)
	if err != nil {
		return MediaCharacter{}, err
	}
	tx := txI.(Store)

	var char MediaCharacter
	err = func() error {
		u, err := tx.UserStore().Get(ctx, uint64(userID))
		if err != nil {
			return err
		}

		var updateUser func() error
		switch {
		case time.Now().After(u.Date.Time.Add(config.RollCooldown)):
			updateUser = func() error {
				return tx.UserStore().UpdateDate(ctx, users.UpdateDateParams{
					Date:   pgtype.Timestamp{Time: time.Now(), Valid: true},
					UserID: uint64(userID),
				})
			}
		case u.Tokens >= config.TokensNeeded:
			updateUser = func() error {
				_, err := tx.UserStore().ConsumeTokens(ctx, users.ConsumeTokensParams{
					Tokens: config.TokensNeeded,
					UserID: uint64(userID),
				})
				return err
			}
		default:
			return ErrRollCooldown{Until: u.Date.Time.Add(config.RollCooldown), MissingTokens: int(config.TokensNeeded - u.Tokens)}
		}

		charsIDs, err := tx.CollectionStore().ListIDs(ctx, uint64(userID))
		if err != nil {
			return err
		}

		c, err := animeService.RandomChar(ctx, charsIDs...)
		if err != nil {
			return err
		}
		char = c

		err = tx.CollectionStore().Insert(ctx, characters.InsertParams{
			ID:     int64(c.ID),
			UserID: uint64(userID),
			Image:  c.ImageURL,
			Name:   c.Name,
			Type:   "ROLL",
		})
		if err != nil {
			return err
		}

		return updateUser()
	}()
	if err != nil {
		_ = tx.Rollback(ctx)
		return MediaCharacter{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return MediaCharacter{}, err
	}

	return char, nil
}
