package sync

import (
	"context"
	"log/slog"
	"time"

	"github.com/failsafe-go/failsafe-go"
	"github.com/failsafe-go/failsafe-go/ratelimiter"
	"github.com/failsafe-go/failsafe-go/retrypolicy"

	"github.com/karitham/waifubot/catalog"
	"github.com/karitham/waifubot/collection"
)

// CharacterFetcher fetches character data from an external source.
type CharacterFetcher interface {
	CharactersByIDs(ctx context.Context, ids []int64) ([]collection.MediaCharacter, error)
}

// Service handles background synchronization of character data with AniList.
type Service struct {
	Store    catalog.Store
	Anilist  CharacterFetcher
	Executor failsafe.Executor[[]collection.MediaCharacter]
}

// NewService creates a sync Service with rate-limiting (5 req/min) and retry
// policies (3 retries, exponential backoff) for AniList calls.
func NewService(store catalog.Store, fetcher CharacterFetcher) *Service {
	rl := ratelimiter.NewSmoothBuilder[[]collection.MediaCharacter](5, time.Minute).
		WithMaxWaitTime(2 * time.Minute).
		Build()
	rp := retrypolicy.NewBuilder[[]collection.MediaCharacter]().
		WithMaxRetries(3).
		WithBackoff(2*time.Second, 10*time.Second).
		WithJitter(500 * time.Millisecond).
		Build()

	return &Service{
		Store:    store,
		Anilist:  fetcher,
		Executor: failsafe.With(rl, rp),
	}
}

// fetchCharacters fetches characters via the Executor (rate-limited + retry) when available,
// falling back to a direct call when Executor is nil (e.g. in tests).
func (s *Service) fetchCharacters(ctx context.Context, ids []int64) ([]collection.MediaCharacter, error) {
	if s.Executor != nil {
		return s.Executor.WithContext(ctx).Get(func() ([]collection.MediaCharacter, error) {
			return s.Anilist.CharactersByIDs(ctx, ids)
		})
	}
	return s.Anilist.CharactersByIDs(ctx, ids)
}

// Run starts the background sync worker that continuously updates stale characters.
// It fetches characters that need syncing, updates their favorites and media_title
// from AniList, and applies rate limiting to respect API constraints.
//
// The worker runs indefinitely until the context is cancelled.
// Rate limit: 5 requests per minute = 1 batch every 12 seconds.
func (s *Service) Run(ctx context.Context) {
	const batchSize = 20

	slog.Info("starting favorites sync worker")

	// Cursor for pagination: (updated_at, id)
	// Initialize to epoch to fetch all characters
	cursorUpdatedAt := time.Unix(0, 0)
	var cursorID int64

	var backoff time.Duration

	for {
		select {
		case <-ctx.Done():
			slog.Info("favorites sync worker stopped")
			return
		default:
		}

		// Fetch stale characters
		characters, err := s.Store.GetStaleCharacters(ctx, cursorUpdatedAt, cursorID, batchSize)
		if err != nil {
			slog.Error("failed to get stale characters", "error", err)
			if backoff == 0 {
				backoff = time.Second
			} else {
				backoff *= 2
				if backoff > time.Minute {
					backoff = time.Minute
				}
			}
			time.Sleep(backoff)
			continue
		}
		backoff = 0

		if len(characters) == 0 {
			cursorUpdatedAt = time.Unix(0, 0)
			cursorID = 0
			time.Sleep(time.Minute)
			continue
		}

		// Bulk fetch and update
		cursorUpdatedAt, cursorID = s.ProcessBatch(ctx, characters)
	}
}

// ProcessBatch handles fetching and updating a batch of characters.
// The cursor always advances past the entire batch using the original pre-update
// timestamps, so partial failures don't cause the worker to get stuck.
func (s *Service) ProcessBatch(ctx context.Context, characters []catalog.Character) (time.Time, int64) {
	ids := make([]int64, len(characters))
	for i, char := range characters {
		ids[i] = char.ID
	}

	freshChars, err := s.fetchCharacters(ctx, ids)
	if err != nil {
		slog.Error("failed to bulk fetch characters", "error", err)
		// Still advance cursor to avoid getting stuck on this batch
		last := characters[len(characters)-1]
		return last.UpdatedAt, last.ID
	}

	freshMap := make(map[int64]collection.MediaCharacter, len(freshChars))
	for _, fc := range freshChars {
		freshMap[fc.ID] = fc
	}

	for _, char := range characters {
		fresh, ok := freshMap[char.ID]
		if !ok {
			slog.Warn("character not found in anilist, skipping", "character_id", char.ID)
			continue
		}

		_, err := s.Store.UpdateCharacterSync(ctx, catalog.Character{
			ID:         fresh.ID,
			Name:       fresh.Name,
			Image:      fresh.ImageURL,
			MediaTitle: fresh.MediaTitle,
			Favorites:  fresh.Favorites,
		})
		if err != nil {
			slog.Error("failed to update character", "character_id", char.ID, "error", err)
		}
	}

	// Always advance to last character in the batch using original pre-update values
	last := characters[len(characters)-1]
	return last.UpdatedAt, last.ID
}
