package idprovider

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew_DefaultBatchSize(t *testing.T) {
	p := New(100, Config{}) // zero config
	// Exhaust the pool, verify we get batches of 50.
	batches := 0
	for {
		batch := p.NextBatch()
		if batch == nil {
			break
		}
		batches++
		assert.Len(t, batch, 50, "each batch (except possibly last) should be 50")
	}
	assert.Equal(t, 2, batches, "100 IDs / 50 batchSize = 2 batches")
}

func TestNew_ZeroMaxID(t *testing.T) {
	p := New(0, Config{BatchSize: 10})
	assert.NotNil(t, p)
	batch := p.NextBatch()
	require.NotNil(t, batch)
	assert.Len(t, batch, 1, "maxID clamped to 1, so only ID 1 should exist")
	assert.Equal(t, int64(1), batch[0])
	// Pool should be exhausted after one batch.
	assert.Nil(t, p.NextBatch())
}

func TestNextBatch_ReturnsFullBatches(t *testing.T) {
	p := New(100, Config{BatchSize: 10})
	batches := 0
	for {
		batch := p.NextBatch()
		if batch == nil {
			break
		}
		batches++
		// All 10 batches should be full size 10.
		assert.Len(t, batch, 10)
	}
	assert.Equal(t, 10, batches)
}

func TestNextBatch_LastBatchPartial(t *testing.T) {
	p := New(5, Config{BatchSize: 50})
	batch := p.NextBatch()
	require.NotNil(t, batch)
	assert.Len(t, batch, 5, "only 5 IDs exist, batch should be partial")
	// Exhausted.
	assert.Nil(t, p.NextBatch())
}

func TestNextBatch_ReturnsNilWhenExhausted(t *testing.T) {
	p := New(3, Config{BatchSize: 10})
	assert.NotNil(t, p.NextBatch()) // batch of 3
	assert.Nil(t, p.NextBatch())    // exhausted
	assert.Nil(t, p.NextBatch())    // still exhausted
}

func TestNextBatch_AllIDsInRange(t *testing.T) {
	const maxID int64 = 100
	p := New(maxID, Config{BatchSize: 10})
	seen := make(map[int64]struct{})
	for {
		batch := p.NextBatch()
		if batch == nil {
			break
		}
		for _, id := range batch {
			seen[id] = struct{}{}
		}
	}
	assert.Len(t, seen, int(maxID))
	for i := int64(1); i <= maxID; i++ {
		assert.Contains(t, seen, i)
	}
}

func TestSetMaxID_RebuildsAfterExhaustion(t *testing.T) {
	p := New(5, Config{BatchSize: 10})
	p.NextBatch()                // consume
	assert.Nil(t, p.NextBatch()) // exhausted

	p.SetMaxID(10)
	seen := make(map[int64]struct{})
	for {
		batch := p.NextBatch()
		if batch == nil {
			break
		}
		for _, id := range batch {
			seen[id] = struct{}{}
		}
	}
	assert.Len(t, seen, 10)
	for i := int64(1); i <= 10; i++ {
		assert.Contains(t, seen, i)
	}
}

func TestSetMaxID_Reshuffles(t *testing.T) {
	p := New(100, Config{BatchSize: 10})

	// Collect order after first build.
	first := collectAll(p)

	// Rebuild.
	p.SetMaxID(100)
	second := collectAll(p)

	// Verify both contain all IDs.
	assert.Len(t, first, 100)
	assert.Len(t, second, 100)

	// They should be different orderings with high probability (Fisher-Yates).
	// Compare element-by-element; if every position matches, shuffle is broken.
	same := true
	for i := range first {
		if first[i] != second[i] {
			same = false
			break
		}
	}
	assert.False(t, same, "two SetMaxID(100) calls should produce different orderings")
}

func TestSetPool(t *testing.T) {
	t.Run("replaces pool with given IDs", func(t *testing.T) {
		p := New(100, Config{BatchSize: 10})
		// Exhaust the original pool
		for p.NextBatch() != nil {
		}

		// Set a new pool
		ids := []int64{10, 20, 30, 40, 50}
		p.SetPool(ids)

		// Collect all batches
		var got []int64
		for {
			batch := p.NextBatch()
			if batch == nil {
				break
			}
			got = append(got, batch...)
		}

		assert.ElementsMatch(t, ids, got)
		assert.Len(t, got, 5)
	})

	t.Run("shuffles on every call", func(t *testing.T) {
		p := New(100, Config{BatchSize: 10})
		for p.NextBatch() != nil {
		}

		ids := []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
		p.SetPool(ids)
		order1 := collectAll(p)

		p.SetPool(ids)
		order2 := collectAll(p)

		// Over 50 runs, the order should differ at least once
		// (statistically certain with Fisher-Yates)
		different := false
		for i := range order1 {
			if order1[i] != order2[i] {
				different = true
				break
			}
		}
		assert.True(t, different, "two consecutive SetPool calls should produce different orderings")
	})
}

func TestConcurrent(t *testing.T) {
	p := New(1000, Config{BatchSize: 50})
	var wg sync.WaitGroup

	// Multiple goroutines reading batches concurrently.
	for range 10 {
		wg.Go(func() {
			for {
				batch := p.NextBatch()
				if batch == nil {
					return
				}
				_ = batch
			}
		})
	}

	// Concurrently rebuild the pool.
	wg.Go(func() {
		p.SetMaxID(500)
	})

	wg.Wait()
	// No panic, no data race — verified with -race.
}

// collectAll drains p into a slice of all IDs returned.
func collectAll(p *Provider) []int64 {
	var all []int64
	for {
		batch := p.NextBatch()
		if batch == nil {
			break
		}
		all = append(all, batch...)
	}
	return all
}
