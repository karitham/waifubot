package sync

import (
	"context"
	"fmt"
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
	BackfillPage(ctx context.Context, page, perPage int) ([]collection.MediaCharacter, error)
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

// fetchBackfillPage fetches a backfill page via the Executor (rate-limited + retry) when available,
// falling back to a direct call when Executor is nil (e.g. in tests).
func (s *Service) fetchBackfillPage(ctx context.Context, page, perPage int) ([]collection.MediaCharacter, error) {
	if s.Executor != nil {
		return s.Executor.WithContext(ctx).Get(func() ([]collection.MediaCharacter, error) {
			return s.Anilist.BackfillPage(ctx, page, perPage)
		})
	}
	return s.Anilist.BackfillPage(ctx, page, perPage)
}

// Backfill paginates all characters from AniList sorted by favorites descending,
// upserting each into the local catalog. It then runs a deletion sweep to detect
// characters that were removed from AniList.
//
// Idempotent — safe to re-run. Rate-limited to 5 req/min.
func (s *Service) Backfill(ctx context.Context, perPage int) error {
	page := 1
	for {
		chars, err := s.fetchBackfillPage(ctx, page, perPage)
		if err != nil {
			return fmt.Errorf("failed to fetch page %d: %w", page, err)
		}
		if len(chars) == 0 {
			break
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

		slog.Debug("backfill page complete", "page", page, "count", len(chars))

		if len(chars) < perPage {
			break
		}
		page++
	}

	// Deletion sweep: detect characters in our DB that no longer exist on AniList
	if err := s.deletionSweep(ctx); err != nil {
		slog.Error("deletion sweep failed", "error", err)
	}

	slog.Info("backfill complete", "pages", page)
	return nil
}

// deletionSweep fetches all active character IDs from the local catalog,
// batch-fetches them from AniList, and marks any that no longer exist as inactive.
func (s *Service) deletionSweep(ctx context.Context) error {
	activeIDs, err := s.Store.GetActiveIDs(ctx)
	if err != nil {
		return fmt.Errorf("failed to get active IDs: %w", err)
	}

	const batchSize = 50
	for i := 0; i < len(activeIDs); i += batchSize {
		end := i + batchSize
		if end > len(activeIDs) {
			end = len(activeIDs)
		}
		batch := activeIDs[i:end]

		freshChars, err := s.fetchCharacters(ctx, batch)
		if err != nil {
			slog.Error("deletion sweep fetch failed", "batch_start", i, "error", err)
			continue
		}

		fetchedIDs := make(map[int64]struct{}, len(freshChars))
		for _, fc := range freshChars {
			fetchedIDs[fc.ID] = struct{}{}
		}

		for _, id := range batch {
			if _, found := fetchedIDs[id]; !found {
				if err := s.Store.MarkCharacterInactive(ctx, id); err != nil {
					slog.Error("failed to mark character inactive", "character_id", id, "error", err)
				} else {
					slog.Debug("marked character inactive", "character_id", id)
				}
			}
		}
	}

	return nil
}

// Run starts the background sync loop. It runs a full backfill + deletion sweep,
// then sleeps before repeating. The loop continues until the context is cancelled.
//
// Rate limit: 5 req/min = approximately 1 page every 12 seconds.
func (s *Service) Run(ctx context.Context) {
	slog.Info("starting character sync worker")

	for {
		if err := s.Backfill(ctx, 50); err != nil {
			slog.Error("backfill failed", "error", err)
		}

		select {
		case <-ctx.Done():
			slog.Info("character sync worker stopped")
			return
		case <-time.After(24 * time.Hour):
		}
	}
}
