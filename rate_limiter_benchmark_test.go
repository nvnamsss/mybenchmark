package mybenchmark

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
)

// Sink variable — prevents dead-code elimination of Allow() results.
var rateLimiterSinkBool bool

// ============================================================================
// centralStore
//
// Simulates the shared distributed counter service.
// A single background goroutine owns the counters (no locks needed inside it);
// callers communicate via a request/response channel pair — one channel
// round-trip per call models the IPC cost of a real network hop.
// ============================================================================

type hopReq struct {
	key    string
	delta  int64
	respCh chan bool
}

type centralStore struct {
	reqCh chan hopReq
	hops  atomic.Int64
}

func newCentralStore(limit int64) *centralStore {
	c := &centralStore{
		// Buffer absorbs bursts; the server goroutine still processes one at a time.
		reqCh: make(chan hopReq, 64),
	}
	go func() {
		counters := make(map[string]int64)
		for req := range c.reqCh {
			counters[req.key] += req.delta
			req.respCh <- counters[req.key] <= limit
		}
	}()
	return c
}

// add sends a counter increment to the server goroutine and blocks until the
// response arrives — one round-trip == one simulated network hop.
func (c *centralStore) add(key string, delta int64) bool {
	c.hops.Add(1)
	ch := make(chan bool, 1)
	c.reqCh <- hopReq{key: key, delta: delta, respCh: ch}
	return <-ch
}

// close shuts down the server goroutine; must be called when the store is
// no longer needed (use b.Cleanup / t.Cleanup).
func (c *centralStore) close() { close(c.reqCh) }

// ============================================================================
// ExactRateLimiter
//
// Baseline: one network hop on every Allow() call.  Equivalent to a classic
// single Redis INCR per request.
// ============================================================================

type ExactRateLimiter struct {
	central *centralStore
}

func newExactRateLimiter(limit int64) *ExactRateLimiter {
	return &ExactRateLimiter{central: newCentralStore(limit)}
}

func (r *ExactRateLimiter) Allow(key string) bool {
	return r.central.add(key, 1)
}

// ============================================================================
// approxNode
//
// One logical server replica in the distributed approximate rate limiter.
// Each node maintains per-key local counters and syncs to the central store
// only every syncThreshold requests, reducing network hops by ~syncThreshold.
//
// Nodes are NOT safe for concurrent use; in a real deployment each replica
// owns its state exclusively.  In benchmarks, one goroutine == one node.
// ============================================================================

type approxLocalState struct {
	count    int64
	exceeded bool
}

type approxNode struct {
	locals  map[string]*approxLocalState
	central *centralStore
	thresh  int64
}

func newApproxNode(central *centralStore, syncThreshold int64) *approxNode {
	return &approxNode{
		locals:  make(map[string]*approxLocalState),
		central: central,
		thresh:  syncThreshold,
	}
}

func (n *approxNode) Allow(key string) bool {
	st, ok := n.locals[key]
	if !ok {
		st = &approxLocalState{}
		n.locals[key] = st
	}
	if st.exceeded {
		return false
	}
	st.count++
	if st.count%n.thresh == 0 {
		if !n.central.add(key, n.thresh) {
			st.exceeded = true
			return false
		}
	}
	return true
}

// approxSyncThreshold mirrors the real implementation's 100-hop reduction.
const approxSyncThreshold = 100

// noLimit is used for pure-throughput benchmarks where admission decisions
// are irrelevant and we only measure synchronisation overhead.
const noLimit = int64(1<<63 - 1)

// ============================================================================
// Benchmarks — Exact rate limiter
// ============================================================================

func BenchmarkRateLimiter_Exact(b *testing.B) {
	rl := newExactRateLimiter(noLimit)
	b.Cleanup(rl.central.close)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rateLimiterSinkBool = rl.Allow("req")
	}
	b.ReportMetric(float64(rl.central.hops.Load())/b.Elapsed().Seconds(), "hops/s")
}

func BenchmarkRateLimiter_Exact_Parallel(b *testing.B) {
	rl := newExactRateLimiter(noLimit)
	b.Cleanup(rl.central.close)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			rateLimiterSinkBool = rl.Allow("req")
		}
	})
	b.ReportMetric(float64(rl.central.hops.Load())/b.Elapsed().Seconds(), "hops/s")
}

// ============================================================================
// Benchmarks — Approximate rate limiter
// ============================================================================

func BenchmarkRateLimiter_Approximate(b *testing.B) {
	central := newCentralStore(noLimit)
	b.Cleanup(central.close)
	node := newApproxNode(central, approxSyncThreshold)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rateLimiterSinkBool = node.Allow("req")
	}
	b.ReportMetric(float64(central.hops.Load())/b.Elapsed().Seconds(), "hops/s")
}

func BenchmarkRateLimiter_Approximate_Parallel(b *testing.B) {
	central := newCentralStore(noLimit)
	b.Cleanup(central.close)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		// Each goroutine is an independent node with its own local counters.
		node := newApproxNode(central, approxSyncThreshold)
		for pb.Next() {
			rateLimiterSinkBool = node.Allow("req")
		}
	})
	b.ReportMetric(float64(central.hops.Load())/b.Elapsed().Seconds(), "hops/s")
}

// ============================================================================
// BenchmarkRateLimiter_Throughput — side-by-side sub-benchmarks for easy
// throughput ratio reading between the two approaches.
// ============================================================================

func BenchmarkRateLimiter_Throughput(b *testing.B) {
	b.Run("Exact", func(b *testing.B) {
		rl := newExactRateLimiter(noLimit)
		b.Cleanup(rl.central.close)
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				rateLimiterSinkBool = rl.Allow("req")
			}
		})
		b.ReportMetric(float64(rl.central.hops.Load())/b.Elapsed().Seconds(), "hops/s")
	})

	b.Run("Approximate", func(b *testing.B) {
		central := newCentralStore(noLimit)
		b.Cleanup(central.close)
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			node := newApproxNode(central, approxSyncThreshold)
			for pb.Next() {
				rateLimiterSinkBool = node.Allow("req")
			}
		})
		b.ReportMetric(float64(central.hops.Load())/b.Elapsed().Seconds(), "hops/s")
	})
}

// ============================================================================
// TestApproximateRateLimiter_ErrorRate
//
// Validates the ~10% error rate claim:
//   - 20 nodes, limit = 10 000, sync threshold = 100
//   - Each node sends 1 000 requests  →  20 000 total (2× limit)
//   - Expected: ~1 000 requests above the limit are incorrectly allowed
//     (≈ numNodes × syncThreshold/2 optimistic in-flight requests)
//   - Expected error rate: ~10% of above-limit requests
//   - Also logs network-hop reduction for the 100× claim
// ============================================================================

func TestApproximateRateLimiter_ErrorRate(t *testing.T) {
	const (
		numNodes = 20
		limit    = 10_000
		thresh   = 100
		perNode  = 1_000 // total = 20 000, 2× limit
	)

	central := newCentralStore(limit)
	t.Cleanup(central.close)

	var totalAllowed atomic.Int64
	var wg sync.WaitGroup
	wg.Add(numNodes)
	for range numNodes {
		go func() {
			defer wg.Done()
			node := newApproxNode(central, thresh)
			var allowed int64
			for range perNode {
				if node.Allow("key") {
					allowed++
				}
			}
			totalAllowed.Add(allowed)
		}()
	}
	wg.Wait()

	allowed := totalAllowed.Load()
	totalRequests := int64(numNodes * perNode) // 20 000
	aboveLimit := totalRequests - limit        // 10 000
	incorrect := allowed - limit
	if incorrect < 0 {
		incorrect = 0
	}
	errorRate := float64(incorrect) / float64(aboveLimit) * 100

	hops := central.hops.Load()
	hopReduction := float64(totalRequests) / float64(hops)

	t.Logf("allowed: %d / %d  (limit: %d)", allowed, totalRequests, limit)
	t.Logf("incorrectly allowed above limit: %d / %d  →  error rate: %.1f%%", incorrect, aboveLimit, errorRate)
	t.Logf("network hops: %d  (exact would be %d  →  %.0fx reduction)", hops, totalRequests, hopReduction)

	if errorRate < 1.0 || errorRate > 30.0 {
		t.Errorf("error rate %.1f%% outside expected range [1%%, 30%%]", errorRate)
	}
	if hopReduction < 50 || hopReduction > 200 {
		t.Errorf("hop reduction %.0fx outside expected range [50x, 200x]", hopReduction)
	}
}

// ============================================================================
// BenchmarkRateLimiter_Tradeoff — sweeps sync thresholds to show how error
// rate and throughput improvement trade off.  Run with -v to see the table.
// Each sub-benchmark is parallel (one goroutine per logical node) so the
// throughput numbers are directly comparable to Exact_Parallel.
// ============================================================================

func BenchmarkRateLimiter_Tradeoff(b *testing.B) {
	thresholds := []int64{10, 25, 50, 100}
	for _, thresh := range thresholds {
		thresh := thresh
		b.Run(fmt.Sprintf("Threshold_%d", thresh), func(b *testing.B) {
			central := newCentralStore(noLimit)
			b.Cleanup(central.close)
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				node := newApproxNode(central, thresh)
				for pb.Next() {
					rateLimiterSinkBool = node.Allow("req")
				}
			})
			b.ReportMetric(float64(central.hops.Load())/b.Elapsed().Seconds(), "hops/s")
		})
	}
}

// TestRateLimiter_TradeoffTable prints a human-readable table of
// error rate and effective overhead per request for each sync threshold.
// Run with -v to see the output:  go test -v -run=TestRateLimiter_TradeoffTable
func TestRateLimiter_TradeoffTable(t *testing.T) {
	const (
		numNodes    = 20
		limit       = 10_000
		perNode     = 1_000
		totalReqs   = numNodes * perNode // 20 000
		aboveLimit  = totalReqs - limit  // 10 000
	)

	thresholds := []int64{10, 25, 50, 100}
	t.Logf("%-12s  %-12s  %-18s", "threshold", "error rate", "hops (vs exact)")
	t.Logf("%-12s  %-12s  %-18s", "---------", "----------", "---------------")
	for _, thresh := range thresholds {
		central := newCentralStore(int64(limit))
		var totalAllowed atomic.Int64
		var wg sync.WaitGroup
		wg.Add(numNodes)
		for range numNodes {
			go func() {
				defer wg.Done()
				node := newApproxNode(central, thresh)
				var allowed int64
				for range perNode {
					if node.Allow("key") {
						allowed++
					}
				}
				totalAllowed.Add(allowed)
			}()
		}
		wg.Wait()
		central.close()

		incorrect := totalAllowed.Load() - int64(limit)
		if incorrect < 0 {
			incorrect = 0
		}
		errorRate := float64(incorrect) / float64(aboveLimit) * 100
		hopReduction := float64(totalReqs) / float64(central.hops.Load())
		t.Logf("%-12d  %-11.1f%%  ~%.0fx fewer hops", thresh, errorRate, hopReduction)
	}
}
