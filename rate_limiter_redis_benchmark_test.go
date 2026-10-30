package mybenchmark

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

// ============================================================================
// requireRedis
//
// Skip guard — dials Redis and calls tb.Skip if unreachable.
// Use at the top of every benchmark and test in this file so CI without a
// local Redis instance does not break.
// ============================================================================

func requireRedis(tb testing.TB) *redis.Client {
	tb.Helper()
	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		tb.Skip("Redis not available:", err)
	}
	return rdb
}

// ============================================================================
// redisStore
//
// Real Redis-backed counter store.  Mirrors the centralStore interface used
// by the channel-based benchmarks so the rate-limiter types above it are
// interchangeable.
//
// add() uses a Lua script to atomically INCRBY + EXPIRE in a single round-
// trip, avoiding the race where the key could expire between two separate
// commands.
// ============================================================================

// luaIncrByExpire atomically increments a key by delta and (re)sets its TTL.
// Returns the new value so the caller can compare against the limit.
var luaIncrByExpire = redis.NewScript(`
local v = redis.call("INCRBY", KEYS[1], ARGV[1])
if v == tonumber(ARGV[1]) then
    redis.call("EXPIRE", KEYS[1], ARGV[2])
end
return v
`)

type redisStore struct {
	client *redis.Client
	limit  int64
	ttlSec int64 // TTL in whole seconds passed to Lua
	hops   atomic.Int64
}

func newRedisStore(rdb *redis.Client, limit int64, ttl time.Duration) *redisStore {
	return &redisStore{
		client: rdb,
		limit:  limit,
		ttlSec: int64(ttl.Seconds()),
	}
}

// add increments key by delta in Redis and returns whether the new total is
// within the limit.  One call == one Redis round-trip.
func (s *redisStore) add(ctx context.Context, key string, delta int64) bool {
	s.hops.Add(1)
	val, err := luaIncrByExpire.Run(ctx, s.client, []string{key}, delta, s.ttlSec).Int64()
	if err != nil {
		// Treat Redis errors as over-limit (fail-closed) to avoid bypassing
		// the rate limit on store outages.
		return false
	}
	return val <= s.limit
}

// close releases the Redis client connection.
func (s *redisStore) close() { _ = s.client.Close() }

// ============================================================================
// RedisExactRateLimiter
//
// Baseline: one Redis round-trip on every Allow() call.
// Equivalent to the classic single INCR-per-request pattern.
// ============================================================================

type RedisExactRateLimiter struct {
	store *redisStore
}

func newRedisExactRateLimiter(rdb *redis.Client, limit int64) *RedisExactRateLimiter {
	return &RedisExactRateLimiter{store: newRedisStore(rdb, limit, 60*time.Second)}
}

func (r *RedisExactRateLimiter) Allow(ctx context.Context, key string) bool {
	return r.store.add(ctx, key, 1)
}

// ============================================================================
// RedisApproxRateLimiter
//
// Optimised: accumulates increments locally and flushes to Redis only every
// approxSyncThreshold (100) calls, reducing Redis round-trips by ~100×.
// Reuses approxLocalState from rate_limiter_benchmark_test.go (same package).
//
// Not safe for concurrent use — each goroutine / logical node owns its own
// instance, matching the approxNode pattern.
// ============================================================================

type RedisApproxRateLimiter struct {
	store  *redisStore
	locals map[string]*approxLocalState
	thresh int64
}

func newRedisApproxRateLimiter(rdb *redis.Client, limit int64) *RedisApproxRateLimiter {
	return &RedisApproxRateLimiter{
		store:  newRedisStore(rdb, limit, 60*time.Second),
		locals: make(map[string]*approxLocalState),
		thresh: approxSyncThreshold, // 100, from rate_limiter_benchmark_test.go
	}
}

func (r *RedisApproxRateLimiter) Allow(ctx context.Context, key string) bool {
	st, ok := r.locals[key]
	if !ok {
		st = &approxLocalState{}
		r.locals[key] = st
	}
	if st.exceeded {
		return false
	}
	st.count++
	if st.count%r.thresh == 0 {
		if !r.store.add(ctx, key, r.thresh) {
			st.exceeded = true
			return false
		}
	}
	return true
}

// ============================================================================
// Phase 2 — Throughput benchmarks
//
// All benchmarks use noLimit so admission decisions never block measurement;
// we measure pure per-request overhead of each design.
//
// Key cleanup via b.Cleanup(DEL) keeps runs idempotent — the counter never
// overflows across -count=3 repetitions.
// ============================================================================

// BenchmarkRedisExact_Allow — one Redis round-trip per Allow() call.
// Establishes the real-network baseline: every request pays full RTT cost.
func BenchmarkRedisExact_Allow(b *testing.B) {
	rdb := requireRedis(b)
	ctx := context.Background()
	b.Cleanup(func() { rdb.Del(ctx, "bench:exact") })

	rl := newRedisExactRateLimiter(rdb, noLimit)
	b.ResetTimer()
	for b.Loop() {
		rateLimiterSinkBool = rl.Allow(ctx, "bench:exact")
	}
	b.ReportMetric(float64(rl.store.hops.Load())/b.Elapsed().Seconds(), "hops/s")
}

// BenchmarkRedisApprox_Allow — batched INCRBY; one Redis hop every 100 calls.
// Measures the per-request overhead of the approximate design on real Redis.
func BenchmarkRedisApprox_Allow(b *testing.B) {
	rdb := requireRedis(b)
	ctx := context.Background()
	b.Cleanup(func() { rdb.Del(ctx, "bench:approx") })

	rl := newRedisApproxRateLimiter(rdb, noLimit)
	b.ResetTimer()
	for b.Loop() {
		rateLimiterSinkBool = rl.Allow(ctx, "bench:approx")
	}
	b.ReportMetric(float64(rl.store.hops.Load())/b.Elapsed().Seconds(), "hops/s")
}

// BenchmarkRedisApprox_Allow_Parallel — parallel approximate limiter.
// Each goroutine owns its own RedisApproxRateLimiter (no shared local state),
// matching the one-node-per-replica deployment model.
// Reports req/s for direct throughput comparison against the Exact baseline.
func BenchmarkRedisApprox_Allow_Parallel(b *testing.B) {
	rdb := requireRedis(b)
	ctx := context.Background()
	b.Cleanup(func() { rdb.Del(ctx, "bench:approx:parallel") })

	// Shared store so all goroutines flush to the same Redis key; each
	// goroutine gets its own limiter instance for independent local state.
	sharedStore := newRedisStore(rdb, noLimit, 60*time.Second)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		rl := &RedisApproxRateLimiter{
			store:  sharedStore,
			locals: make(map[string]*approxLocalState),
			thresh: approxSyncThreshold,
		}
		for pb.Next() {
			rateLimiterSinkBool = rl.Allow(ctx, "bench:approx:parallel")
		}
	})
	b.ReportMetric(float64(b.N)/b.Elapsed().Seconds(), "req/s")
}

// ============================================================================
// Phase 3 — Accuracy measurement
//
// TestRedisApprox_OverAdmission mirrors TestApproximateRateLimiter_ErrorRate
// from rate_limiter_benchmark_test.go but uses real Redis as the backing store.
//
// Setup:
//   - 20 nodes (goroutines), limit = 10 000, sync threshold = 100
//   - Each node sends 1 000 requests → 20 000 total (2× limit)
//
// Expected behaviour:
//   - At most (numNodes × syncThreshold) = 2 000 requests above the limit
//     can be incorrectly allowed (worst case: every node holds a full batch
//     of 100 in flight when the limit is crossed).
//   - Error rate ≤ 20% of the above-limit window (2 000 / 10 000).
//   - Redis hops ≈ totalRequests / syncThreshold → ~200; exact would be 20 000.
//
// The test key is deleted in t.Cleanup so it is idempotent across runs.
// ============================================================================

func TestRedisApprox_OverAdmission(t *testing.T) {
	const (
		numNodes = 20
		limit    = 10_000
		thresh   = 100
		perNode  = 1_000 // total = 20 000, 2× limit
	)

	rdb := requireRedis(t)
	ctx := context.Background()
	t.Cleanup(func() { rdb.Del(ctx, "test:approx:over") })

	// Shared Redis store; each goroutine gets its own limiter (independent
	// local state) but they all flush increments to the same Redis key.
	sharedStore := newRedisStore(rdb, limit, 60*time.Second)

	var totalAllowed atomic.Int64
	var wg sync.WaitGroup
	wg.Add(numNodes)
	for range numNodes {
		go func() {
			defer wg.Done()
			rl := &RedisApproxRateLimiter{
				store:  sharedStore,
				locals: make(map[string]*approxLocalState),
				thresh: thresh,
			}
			var allowed int64
			for range perNode {
				if rl.Allow(ctx, "test:approx:over") {
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

	hops := sharedStore.hops.Load()
	hopReduction := float64(totalRequests) / float64(hops)

	t.Logf("allowed: %d / %d  (limit: %d)", allowed, totalRequests, limit)
	t.Logf("incorrectly allowed above limit: %d / %d  →  error rate: %.1f%%", incorrect, aboveLimit, errorRate)
	t.Logf("Redis hops: %d  (exact would be %d  →  %.0fx reduction)", hops, totalRequests, hopReduction)

	// Worst-case over-admission: every node holds a full batch in flight.
	// maxIncorrect = numNodes * thresh = 20 * 100 = 2 000
	// → errorRate ≤ 2 000 / 10 000 = 20%
	if errorRate > 20.0 {
		t.Errorf("error rate %.1f%% exceeds 20%% worst-case bound", errorRate)
	}
	if hopReduction < 50 || hopReduction > 200 {
		t.Errorf("hop reduction %.0fx outside expected range [50x, 200x]", hopReduction)
	}
}

// ============================================================================
// TestRedisApprox_ThresholdSweep
//
// Sweeps sync thresholds [10, 25, 50] to find the operating point where the
// measured error rate falls below 5% and 10%, and reports the resulting hop
// reduction at each threshold.
//
// Theory: errorRate ≤ numNodes×thresh / limit
//   thresh=50 → ≤10%   thresh=25 → ≤5%   thresh=10 → ≤2%
// ============================================================================

func TestRedisApprox_ThresholdSweep(t *testing.T) {
	const (
		numNodes = 20
		limit    = 10_000
		perNode  = 1_000 // 20 000 total, 2× limit
	)

	rdb := requireRedis(t)
	ctx := context.Background()

	for _, thresh := range []int64{10, 25, 50} {
		thresh := thresh
		redisKey := fmt.Sprintf("test:sweep:%d", thresh)
		t.Cleanup(func() { rdb.Del(ctx, redisKey) })

		sharedStore := newRedisStore(rdb, limit, 60*time.Second)

		var totalAllowed atomic.Int64
		var wg sync.WaitGroup
		wg.Add(numNodes)
		for range numNodes {
			go func() {
				defer wg.Done()
				rl := &RedisApproxRateLimiter{
					store:  sharedStore,
					locals: make(map[string]*approxLocalState),
					thresh: thresh,
				}
				var allowed int64
				for range perNode {
					if rl.Allow(ctx, redisKey) {
						allowed++
					}
				}
				totalAllowed.Add(allowed)
			}()
		}
		wg.Wait()

		allowed := totalAllowed.Load()
		totalRequests := int64(numNodes * perNode)
		aboveLimit := totalRequests - limit
		incorrect := allowed - limit
		if incorrect < 0 {
			incorrect = 0
		}
		errorRate := float64(incorrect) / float64(aboveLimit) * 100
		hops := sharedStore.hops.Load()
		hopReduction := float64(totalRequests) / float64(hops)

		t.Logf("thresh=%3d | error rate: %5.1f%% | hops: %4d | hop reduction: %5.0fx | allowed: %d",
			thresh, errorRate, hops, hopReduction, allowed)
	}
}

// ============================================================================
// BenchmarkRedisApprox_ThresholdSweep
//
// Measures single-goroutine throughput (ns/op) at thresholds 10, 25, 50 and
// 100 so the throughput cost of tightening the error bound is visible.
// ============================================================================

func BenchmarkRedisApprox_ThresholdSweep(b *testing.B) {
	rdb := requireRedis(b)
	ctx := context.Background()

	for _, thresh := range []int64{10, 25, 50, 100} {
		thresh := thresh
		name := fmt.Sprintf("thresh=%d", thresh)
		b.Run(name, func(b *testing.B) {
			key := fmt.Sprintf("bench:sweep:%d", thresh)
			b.Cleanup(func() { rdb.Del(ctx, key) })
			rl := &RedisApproxRateLimiter{
				store:  newRedisStore(rdb, noLimit, 60*time.Second),
				locals: make(map[string]*approxLocalState),
				thresh: thresh,
			}
			b.ResetTimer()
			for b.Loop() {
				rateLimiterSinkBool = rl.Allow(ctx, key)
			}
			b.ReportMetric(float64(rl.store.hops.Load())/b.Elapsed().Seconds(), "hops/s")
		})
	}
}
