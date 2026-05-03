package sync

import (
	"context"
	"log/slog"
	"sync/atomic"
	"time"

	"github.com/failsafe-go/failsafe-go"
	"github.com/failsafe-go/failsafe-go/circuitbreaker"
	"github.com/failsafe-go/failsafe-go/ratelimiter"
	"github.com/failsafe-go/failsafe-go/retrypolicy"

	"github.com/karitham/waifubot/anilist"
	"github.com/karitham/waifubot/catalog"
	"github.com/karitham/waifubot/collection"
	"github.com/karitham/waifubot/sync/idprovider"
)

// MaxIDProvider optionally provides the maximum character ID.
// Implemented by *anilist.Anilist. When nil, the sync service uses a default bound.
type MaxIDProvider interface {
	MaxCharacterID(ctx context.Context) (int64, error)
}

// Service handles background synchronization of character data with AniList.
type Service struct {
	Store    catalog.Store
	Anilist  anilist.CharacterFetcher
	MaxID    MaxIDProvider // nil → use default bound
	executor failsafe.Executor[[]collection.MediaCharacter]
	maxID    atomic.Int64
}

// NewService creates a sync Service with rate-limiting (5 req/min) and retry
// policies (3 retries, exponential backoff) for AniList calls.
func NewService(store catalog.Store, fetcher anilist.CharacterFetcher, maxIDProvider MaxIDProvider) *Service {
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
		executor: failsafe.With(rl, cb, rp),
		maxID:    atomic.Int64{},
	}
	s.maxID.Store(1_000_000) // fallback bound, updated by refreshMaxID
	return s
}

// discoveryWindow is the number of IDs at the top of the range to probe on
// each pass. New characters get IDs above the previous max, so a short
// window at the tail catches new entries without scanning the full dead space.
const discoveryWindow = 10_000

// Sync fetches one batch of character IDs and processes them.
// Used by the CLI backfill command.
func (s *Service) Sync(ctx context.Context) error {
	provider := idprovider.New(s.maxID.Load(), idprovider.Config{BatchSize: 50})
	batch := provider.NextBatch()
	if batch == nil {
		return nil
	}
	return s.processBatch(ctx, batch)
}

// buildDiscoveryPool returns known active IDs from the catalog plus a
// discovery window at the top of the range for catching new characters.
// Returns nil on error — caller should fall back to the full range.
func (s *Service) buildDiscoveryPool(ctx context.Context) []int64 {
	activeIDs, err := s.Store.GetActiveIDs(ctx)
	if err != nil {
		slog.Warn("failed to get active IDs, falling back to full range", "error", err)
		return nil
	}

	maxID := s.maxID.Load()
	tailStart := maxID - discoveryWindow
	if tailStart < 1 {
		tailStart = 1
	}

	// Deduplicate: active IDs and discovery window may overlap.
	seen := make(map[int64]struct{}, len(activeIDs)+discoveryWindow)
	for _, id := range activeIDs {
		seen[id] = struct{}{}
	}
	for id := tailStart; id <= maxID; id++ {
		seen[id] = struct{}{}
	}

	pool := make([]int64, 0, len(seen))
	for id := range seen {
		pool = append(pool, id)
	}
	return pool
}

// processBatch fetches a batch of IDs from AniList, upserts valid results,
// and marks missing IDs as inactive.
func (s *Service) processBatch(ctx context.Context, ids []int64) error {
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

	if len(chars) < len(ids) {
		if err := s.markMissingInactive(ctx, ids, chars); err != nil {
			slog.Error("failed to mark missing characters inactive", "error", err)
		}
	}

	if len(chars) > 0 {
		slog.Debug("sync batch complete", "requested", len(ids), "found", len(chars))
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

// fetchCharacters fetches characters via the Executor (rate-limited + retry) when available,
// falling back to a direct call when Executor is nil (e.g. in tests).
func (s *Service) fetchCharacters(ctx context.Context, ids []int64) ([]collection.MediaCharacter, error) {
	if s.executor != nil {
		return s.executor.WithContext(ctx).Get(func() ([]collection.MediaCharacter, error) {
			return s.Anilist.CharactersByIDs(ctx, ids)
		})
	}
	return s.Anilist.CharactersByIDs(ctx, ids)
}

// nextBatch returns the next batch of IDs to process, rebuilding the pool
// when exhausted. Returns nil when the context is cancelled.
func (s *Service) nextBatch(ctx context.Context, provider *idprovider.Provider) []int64 {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		batch := provider.NextBatch()
		if batch != nil {
			return batch
		}
		// Pool exhausted — rebuild.
		if s.MaxID != nil {
			s.refreshMaxID(ctx)
		}
		if pool := s.buildDiscoveryPool(ctx); pool != nil {
			provider.SetPool(pool)
		} else {
			provider.SetMaxID(s.maxID.Load())
		}
		batch = provider.NextBatch()
		if batch != nil {
			return batch
		}
		// Still nothing — wait for new characters or cancellation.
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(time.Minute):
		}
	}
}

// backoffSleep sleeps for d or returns false if ctx is cancelled.
func backoffSleep(ctx context.Context, d time.Duration) bool {
	select {
	case <-ctx.Done():
		return false
	case <-time.After(d):
		return true
	}
}

// nextBackoff computes exponential backoff capped at 1 minute.
func nextBackoff(current time.Duration) time.Duration {
	if current == 0 {
		return time.Second
	}
	current *= 2
	if current > time.Minute {
		current = time.Minute
	}
	return current
}

// Run starts the background sync loop. It uses the IDProvider to supply
// batches of character IDs, refreshes the max ID bound when the pool is
// exhausted, and processes each batch. The loop continues until the context
// is cancelled.
func (s *Service) Run(ctx context.Context) {
	slog.Info("starting character sync worker")
	defer slog.Info("character sync worker stopped")

	if s.MaxID != nil {
		s.refreshMaxID(ctx) // first refresh at startup
	}

	provider := idprovider.New(1, idprovider.Config{BatchSize: 50})
	if pool := s.buildDiscoveryPool(ctx); pool != nil {
		provider.SetPool(pool)
	}

	var bo time.Duration
	for {
		batch := s.nextBatch(ctx, provider)
		if batch == nil {
			return
		}

		if err := s.processBatch(ctx, batch); err != nil {
			slog.Error("sync failed", "error", err)
			bo = nextBackoff(bo)
			if !backoffSleep(ctx, bo) {
				return
			}
			continue
		}
		bo = 0
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
