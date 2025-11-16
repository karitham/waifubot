package collection

import (
	"context"

	"github.com/Karitham/corde"
)

// Characters retrieves a user's character collection
func Characters(ctx context.Context, store Store, userID corde.Snowflake) ([]Character, error) {
	charRows, err := store.CollectionStore().List(ctx, uint64(userID))
	if err != nil {
		return nil, err
	}

	chars := make([]Character, len(charRows))
	for i, c := range charRows {
		chars[i] = Character{
			Date:   c.Date.Time,
			Image:  c.Image,
			Name:   c.Name,
			Type:   c.Type,
			UserID: corde.Snowflake(c.UserID),
			ID:     c.ID,
		}
	}

	return chars, nil
}
