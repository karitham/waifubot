package sync

import (
	"context"
	"log/slog"
	"math/rand"
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
	rng      *rand.Rand
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
		rng:      rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// maxCharacterID is the upper bound for random character ID generation.
// AniList character IDs are sequential; this covers the known space while
// keeping hit rate reasonable. Increase if ID range grows significantly.
const maxCharacterID = 1_000_000

// batchSize is the number of random IDs to fetch per API call.
const batchSize = 50

// Sync generates random character IDs, batch-fetches them from AniList,
// and upserts any valid results into the local catalog.
//
// Rate-limited to 5 req/min by the Executor.
func (s *Service) Sync(ctx context.Context) error {
	ids := make([]int64, batchSize)
	for i := range ids {
		ids[i] = s.rng.Int63n(maxCharacterID) + 1
	}

	chars, err := s.fetchCharacters(ctx, ids)
	if err != nil {
		return err
	}

	for _, c := range chars {
		if err := s.Store.UpsertCharacter(ctx, catalog.Character{
			ID:         c.ID,
			Name:       c.Name,
			Image:      c.ImageURL,
			MediaTitle: c.MediaTitle,
			Favorites:  c.Favorites,
		}); err != nil {
			slog.Error("failed to upsert character", "character_id", c.ID, "error", err)
		}
	}

	if len(chars) > 0 {
		slog.Debug("sync batch complete", "requested", batchSize, "found", len(chars))
	}
	return nil
}

// Backfill is a convenience wrapper that calls Sync once.
// Kept for backward compatibility with CLI.
func (s *Service) Backfill(ctx context.Context, _ int) error {
	return s.Sync(ctx)
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

// Run starts the background sync loop. It calls Sync repeatedly, paced by the
// rate limiter (~12s between requests at 5 req/min). The loop continues until
// the context is cancelled.
func (s *Service) Run(ctx context.Context) {
	slog.Info("starting character sync worker")

	for {
		if err := s.Sync(ctx); err != nil {
			slog.Error("sync failed", "error", err)
		}

		select {
		case <-ctx.Done():
			slog.Info("character sync worker stopped")
			return
		default:
		}
	}
}
