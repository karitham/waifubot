// Package idprovider supplies batches of character IDs to the sync worker.
//
// It builds a full shuffled pool of IDs [1, maxID], then serves them in
// batches. When the pool is exhausted, the caller must call SetMaxID to
// rebuild. The type is a pure data structure: no I/O, no errors, simple
// interface, testable without mocks.
package idprovider

import (
	"math/rand"
	"sync"
	"time"
)

// Config configures the Provider. All fields have safe defaults.
type Config struct {
	// BatchSize is the number of IDs per batch (default 50).
	BatchSize int
}

// Provider supplies batches of character IDs to fetch.
//
// It builds a full shuffled pool of IDs [1, maxID], then serves them in
// batches. When exhausted, the caller must call SetMaxID to rebuild.
//
// Thread-safe: NextBatch and SetMaxID can be called concurrently.
type Provider struct {
	mu        sync.Mutex
	maxID     int64
	batchSize int
	pool      []int64 // remaining shuffled IDs; nil when exhausted
	index     int
	rng       *rand.Rand
}

// New creates a Provider for the given maxID.
func New(maxID int64, cfg Config) *Provider {
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = 50
	}
	if maxID <= 0 {
		maxID = 1
	}
	p := &Provider{
		maxID:     maxID,
		batchSize: cfg.BatchSize,
		rng:       rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	p.buildShuffledPool()
	return p
}

// NextBatch returns the next batch of IDs, or nil if the current pool is
// exhausted (call SetMaxID to rebuild).
func (p *Provider) NextBatch() []int64 {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.pool == nil {
		return nil
	}

	remaining := len(p.pool) - p.index
	if remaining <= 0 {
		p.pool = nil
		return nil
	}

	n := min(p.batchSize, remaining)

	batch := p.pool[p.index : p.index+n]
	p.index += n
	return batch
}

// SetMaxID rebuilds the shuffled pool for the new range [1, maxID].
func (p *Provider) SetMaxID(maxID int64) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if maxID <= 0 {
		maxID = 1
	}
	p.maxID = maxID
	p.buildShuffledPool()
}

// SetPool replaces the pool with a shuffled copy of the given IDs.
// The caller controls what goes in (e.g. known active IDs + discovery window).
func (p *Provider) SetPool(ids []int64) {
	p.mu.Lock()
	defer p.mu.Unlock()

	pool := make([]int64, len(ids))
	copy(pool, ids)
	p.rng.Shuffle(len(pool), func(i, j int) {
		pool[i], pool[j] = pool[j], pool[i]
	})
	p.pool = pool
	p.index = 0
}

// buildShuffledPool creates a shuffled slice of all IDs [1, maxID] and resets
// the index. Must be called with p.mu held.
func (p *Provider) buildShuffledPool() {
	n := int(p.maxID)
	pool := make([]int64, n)
	for i := range pool {
		pool[i] = int64(i + 1)
	}
	p.rng.Shuffle(n, func(i, j int) {
		pool[i], pool[j] = pool[j], pool[i]
	})
	p.pool = pool
	p.index = 0
}
