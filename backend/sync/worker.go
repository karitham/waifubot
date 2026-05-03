package sync

import (
	"context"
	"log/slog"
	"math/rand"
	"sync/atomic"
	"time"

	"github.com/failsafe-go/failsafe-go"
	"github.com/failsafe-go/failsafe-go/circuitbreaker"
	"github.com/failsafe-go/failsafe-go/ratelimiter"
	"github.com/failsafe-go/failsafe-go/retrypolicy"

	"github.com/karitham/waifubot/catalog"
	"github.com/karitham/waifubot/collection"
)

// CharacterFetcher fetches character data from an external source.
type CharacterFetcher interface {
	CharactersByIDs(ctx context.Context, ids []int64) ([]collection.MediaCharacter, error)
}

// MaxIDProvider optionally provides the maximum character ID.
// Implemented by *anilist.Anilist. When nil, the sync service uses a default bound.
type MaxIDProvider interface {
	MaxCharacterID(ctx context.Context) (int64, error)
}

// Service handles background synchronization of character data with AniList.
type Service struct {
	Store    catalog.Store
	Anilist  CharacterFetcher
	MaxID    MaxIDProvider // nil → use default bound
	Executor failsafe.Executor[[]collection.MediaCharacter]
	maxID    atomic.Int64
	rng      *rand.Rand
}

// NewService creates a sync Service with rate-limiting (5 req/min) and retry
// policies (3 retries, exponential backoff) for AniList calls.
func NewService(store catalog.Store, fetcher CharacterFetcher, maxIDProvider MaxIDProvider) *Service {
	rl := ratelimiter.NewSmoothBuilder[[]collection.MediaCharacter](5, time.Minute).
		WithMaxWaitTime(2 * time.Minute).
		Build()
	cb := circuitbreaker.NewBuilder[[]collection.MediaCharacter]().
		WithFailureThreshold(5).
		WithSuccessThreshold(2).
		WithDelay(30 * time.Second).
		Build()
	rp := retrypolicy.NewBuilder[[]collection.MediaCharacter]().
		WithMaxRetries(3).
		WithBackoff(2*time.Second, 10*time.Second).
		WithJitter(500 * time.Millisecond).
		Build()

	s := &Service{
		Store:    store,
		Anilist:  fetcher,
		MaxID:    maxIDProvider,
		Executor: failsafe.With(rl, cb, rp),
		maxID:    atomic.Int64{},
		rng:      rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	s.maxID.Store(maxCharacterID)
	return s
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
	limit := s.maxID.Load()
	if limit <= 0 {
		limit = maxCharacterID
	}

	ids := make([]int64, batchSize)
	for i := range ids {
		ids[i] = s.rng.Int63n(limit) + 1
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

	// Mark characters that no longer exist on AniList as inactive.
	if len(chars) < len(ids) {
		if err := s.markMissingInactive(ctx, ids, chars); err != nil {
			slog.Error("failed to mark missing characters inactive", "error", err)
		}
	}

	if len(chars) > 0 {
		slog.Debug("sync batch complete", "requested", batchSize, "found", len(chars))
	}
	return nil
}

// markMissingInactive marks characters that were requested but not returned by
// AniList as inactive. This detects deletions.
func (s *Service) markMissingInactive(ctx context.Context, requested []int64, found []collection.MediaCharacter) error {
	foundIDs := make(map[int64]struct{}, len(found))
	for _, c := range found {
		foundIDs[c.ID] = struct{}{}
	}

	missing := make([]int64, 0, len(requested))
	for _, id := range requested {
		if _, ok := foundIDs[id]; !ok {
			missing = append(missing, id)
		}
	}

	if len(missing) == 0 {
		return nil
	}

	return s.Store.MarkCharactersInactive(ctx, missing)
}

// Backfill is a convenience wrapper that calls Sync once.
// Kept for backward compatibility with CLI.
func (s *Service) Backfill(ctx context.Context) error {
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

	if s.MaxID != nil {
		s.refreshMaxID(ctx)
		go s.periodicRefreshLoop(ctx)
	}

	var backoff time.Duration
	for {
		if err := s.Sync(ctx); err != nil {
			slog.Error("sync failed", "error", err)
			if backoff == 0 {
				backoff = time.Second
			} else {
				backoff *= 2
				if backoff > time.Minute {
					backoff = time.Minute
				}
			}
			select {
			case <-ctx.Done():
				slog.Info("character sync worker stopped")
				return
			case <-time.After(backoff):
			}
			continue
		}
		backoff = 0

		select {
		case <-ctx.Done():
			slog.Info("character sync worker stopped")
			return
		default:
		}
	}
}

// refreshMaxID fetches the current max character ID from AniList and stores it.
// On error, keeps the current bound and logs a warning.
func (s *Service) refreshMaxID(ctx context.Context) {
	v, err := s.MaxID.MaxCharacterID(ctx)
	if err != nil {
		slog.Warn("failed to fetch max character id, keeping current bound", "error", err)
		return
	}
	if v > 0 {
		s.maxID.Store(v)
		slog.Debug("updated max character id", "new_max", v)
	}
}

// periodicRefreshLoop refreshes the max character ID every 6 hours.
func (s *Service) periodicRefreshLoop(ctx context.Context) {
	ticker := time.NewTicker(6 * time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.refreshMaxID(ctx)
		}
	}
}
