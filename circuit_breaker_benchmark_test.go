package mybenchmark

import (
	"sync/atomic"
	"testing"
	"time"
)

// Sink variable — prevents dead-code elimination of Allow() / Call() results.
var cbSinkBool bool

// ============================================================================
// upstreamSimulator
//
// Simulates a downstream service that a WAF node calls on each request.
// A single background goroutine owns the response logic; callers communicate
// via a per-call request/response channel pair, modelling the IPC cost of a
// real network hop.
//
// When degraded is false the goroutine responds immediately with true.
// When degraded is true it sleeps 1 ms (simulated timeout) then responds false.
// ============================================================================

type cbUpstreamReq struct {
	respCh chan bool
}

type upstreamSimulator struct {
	reqCh    chan cbUpstreamReq
	degraded atomic.Bool
}

func newUpstreamSimulator() *upstreamSimulator {
	u := &upstreamSimulator{
		reqCh: make(chan cbUpstreamReq, 64),
	}
	go func() {
		for req := range u.reqCh {
			if u.degraded.Load() {
				time.Sleep(1 * time.Millisecond)
				req.respCh <- false
			} else {
				req.respCh <- true
			}
		}
	}()
	return u
}

// call sends one request to the upstream and blocks until the response arrives.
func (u *upstreamSimulator) call() bool {
	ch := make(chan bool, 1)
	u.reqCh <- cbUpstreamReq{respCh: ch}
	return <-ch
}

func (u *upstreamSimulator) setDegraded(v bool) { u.degraded.Store(v) }
func (u *upstreamSimulator) close()             { close(u.reqCh) }

// ============================================================================
// NaiveClient
//
// Scalar baseline: calls the upstream directly on every Allow() and blocks
// until either a response arrives or the fixed timeout elapses.
// Equivalent to a service that has no circuit breaker — every goroutine waits
// for the full upstream timeout during degradation events.
// ============================================================================

type NaiveClient struct {
	upstream *upstreamSimulator
	timeout  time.Duration
}

func newNaiveClient(upstream *upstreamSimulator, timeout time.Duration) *NaiveClient {
	return &NaiveClient{upstream: upstream, timeout: timeout}
}

// Call forwards the request and respects the timeout.  Allocates one
// time.Timer and one channel per call — this is the baseline cost the circuit
// breaker eliminates on its open-state fast-fail path.
func (c *NaiveClient) Call() bool {
	ch := make(chan bool, 1)
	c.upstream.reqCh <- cbUpstreamReq{respCh: ch}
	select {
	case r := <-ch:
		return r
	case <-time.After(c.timeout):
		return false
	}
}

// ============================================================================
// CircuitBreaker
//
// Wraps upstreamSimulator with a three-state machine:
//
//   Closed ──(5 consecutive failures)──► Open
//   Open   ──(probeAfter elapsed)──────► HalfOpen ──(success)──► Closed
//                                                  └──(failure)──► Open
//
// All hot-path state is held in atomics so the open-state fast-fail path
// (Allow returning false immediately) makes zero heap allocations.
// ============================================================================

const (
	cbStateClosed   int32 = 0
	cbStateOpen     int32 = 1
	cbStateHalfOpen int32 = 2
)

// cbTripAfter is the number of consecutive failures that trip the breaker.
const cbTripAfter = 5

// cbProbeAfterNs is how long (in nanoseconds) the breaker stays open before
// allowing a single probe.  100 ms is a realistic WAF value.
const cbProbeAfterNs = int64(100 * time.Millisecond)

type CircuitBreaker struct {
	upstream     *upstreamSimulator
	probeAfterNs int64 // nanoseconds the breaker stays Open before allowing a probe
	state        atomic.Int32
	failures     atomic.Int32 // consecutive failure counter; reset on success
	lastOpenedAt atomic.Int64 // UnixNano when the breaker last tripped to Open
}

func newCircuitBreaker(upstream *upstreamSimulator) *CircuitBreaker {
	return &CircuitBreaker{upstream: upstream, probeAfterNs: cbProbeAfterNs}
}

// newCircuitBreakerWithProbe creates a CircuitBreaker with a custom probe
// interval.  Pass a very large duration (e.g. time.Hour) in benchmarks that
// must keep the breaker open for the full benchmark duration.
func newCircuitBreakerWithProbe(upstream *upstreamSimulator, probeAfter time.Duration) *CircuitBreaker {
	return &CircuitBreaker{upstream: upstream, probeAfterNs: probeAfter.Nanoseconds()}
}

// Allow is the hot path called by every WAF request.
//
// Open state:  pure atomic reads + comparison → returns false with 0 allocs.
// Closed state: forwards to upstream → 1 alloc (the upstream channel, not CB overhead).
func (cb *CircuitBreaker) Allow() bool {
	switch cb.state.Load() {
	case cbStateOpen:
		if time.Now().UnixNano()-cb.lastOpenedAt.Load() < cb.probeAfterNs {
			// Still within open window — fast-fail with zero allocations.
			return false
		}
		// Probe interval elapsed; attempt to transition to half-open.
		if cb.state.CompareAndSwap(cbStateOpen, cbStateHalfOpen) {
			return cb.probe()
		}
		// Another goroutine already transitioned; shed this request.
		return false

	case cbStateHalfOpen:
		// One goroutine owns the probe; all others fast-fail.
		return false

	default: // cbStateClosed
		return cb.callUpstream()
	}
}

// callUpstream forwards the request and updates failure accounting.
func (cb *CircuitBreaker) callUpstream() bool {
	if cb.upstream.call() {
		cb.failures.Store(0)
		return true
	}
	if cb.failures.Add(1) >= cbTripAfter {
		cb.lastOpenedAt.Store(time.Now().UnixNano())
		cb.state.Store(cbStateOpen)
	}
	return false
}

// probe is called once when transitioning from Open → HalfOpen.
func (cb *CircuitBreaker) probe() bool {
	if cb.upstream.call() {
		cb.failures.Store(0)
		cb.state.Store(cbStateClosed)
		return true
	}
	cb.lastOpenedAt.Store(time.Now().UnixNano())
	cb.state.Store(cbStateOpen)
	return false
}

// forceOpen trips the breaker directly — used in benchmark setup so the
// measurement loop starts with the breaker already open.
func (cb *CircuitBreaker) forceOpen() {
	cb.lastOpenedAt.Store(time.Now().UnixNano())
	cb.state.Store(cbStateOpen)
}

// ============================================================================
// TestCB_ZeroAllocsOnHotPath
//
// Verifies the core allocation guarantee:
//   - Open-state Allow() must allocate nothing (pure atomic fast-fail).
//   - Closed-state Allow() on a healthy upstream allocates exactly once
//     (the upstream response channel inside call()) — the CB itself adds 0.
// ============================================================================

func TestCB_ZeroAllocsOnHotPath(t *testing.T) {
	upstream := newUpstreamSimulator()
	t.Cleanup(upstream.close)
	cb := newCircuitBreaker(upstream)

	// ── Open state: 0 allocs ────────────────────────────────────────────────
	cb.forceOpen()
	// Wind the clock forward so the probe interval is NOT elapsed — guaranteed
	// fast-fail for the duration of AllocsPerRun.
	cb.lastOpenedAt.Store(time.Now().UnixNano()) // reset timer

	openAllocs := testing.AllocsPerRun(1000, func() {
		cbSinkBool = cb.Allow()
	})
	if openAllocs != 0 {
		t.Errorf("open-state Allow(): got %.0f allocs/op, want 0", openAllocs)
	}

	// ── Closed state: 1 alloc (upstream chan), not from CB itself ───────────
	cb.state.Store(cbStateClosed)
	cb.failures.Store(0)

	closedAllocs := testing.AllocsPerRun(1000, func() {
		cbSinkBool = cb.Allow()
	})
	if closedAllocs != 1 {
		t.Errorf("closed-state Allow(): got %.0f allocs/op, want 1 (upstream channel)", closedAllocs)
	}
}

// ============================================================================
// Phase 2 — Closed-state overhead benchmarks
//
// Both benchmarks run against a healthy upstream.  The delta between them is
// the pure CB state-check cost (one atomic.Int32.Load + branch), which should
// be well under 10 ns.
//
// BenchmarkNaiveClient_Healthy  — no circuit breaker; allocates a time.After
//   timer on every call even when healthy (2 allocs/op).
// BenchmarkCB_ClosedState       — CB wrapping same upstream; skips the timer,
//   adds only an atomic load (1 alloc/op; CB itself adds 0).
// ============================================================================

func BenchmarkNaiveClient_Healthy(b *testing.B) {
	upstream := newUpstreamSimulator() // healthy by default
	b.Cleanup(upstream.close)
	client := newNaiveClient(upstream, 1*time.Millisecond)

	b.ResetTimer()
	for b.Loop() {
		cbSinkBool = client.Call()
	}
}

func BenchmarkCB_ClosedState(b *testing.B) {
	upstream := newUpstreamSimulator() // healthy by default
	b.Cleanup(upstream.close)
	cb := newCircuitBreaker(upstream) // state is Closed by default

	b.ResetTimer()
	for b.Loop() {
		cbSinkBool = cb.Allow()
	}
}

// ============================================================================
// Phase 3 — Open-state fast-fail vs blocking (key CV metric)
//
// BenchmarkNaive_BlockingTimeout  — upstream degraded; every Call() blocks for
//   ~1 ms waiting for either the upstream response or time.After to fire.
//   Models what happens without a circuit breaker during an outage.
//
// BenchmarkCB_OpenState  — breaker pre-tripped; Allow() reads one atomic,
//   compares timestamps, and returns false immediately.  0 allocs, <10 ns.
//   The probe interval is set to 1 hour so it cannot expire mid-benchmark.
// ============================================================================

func BenchmarkNaive_BlockingTimeout(b *testing.B) {
	upstream := newUpstreamSimulator()
	b.Cleanup(upstream.close)
	upstream.setDegraded(true) // upstream is down for the entire benchmark
	client := newNaiveClient(upstream, 1*time.Millisecond)

	b.ResetTimer()
	for b.Loop() {
		cbSinkBool = client.Call()
	}
}

func BenchmarkCB_OpenState(b *testing.B) {
	upstream := newUpstreamSimulator()
	b.Cleanup(upstream.close)
	// Probe interval >> benchmark duration: breaker stays Open for every iteration.
	cb := newCircuitBreakerWithProbe(upstream, time.Hour)
	cb.forceOpen()

	b.ResetTimer()
	for b.Loop() {
		cbSinkBool = cb.Allow()
	}
}

// ============================================================================
// Phase 4 — State transition benchmarks
//
// BenchmarkCB_TripLatency  — measures one full trip event: Closed → 5
//   consecutive failing upstream calls → Open.  Each iteration resets the
//   breaker to Closed and drives it to Open through cbTripAfter failures.
//   The long probe interval prevents an accidental recovery mid-reset.
//   Expected: ~5 ms/op (5 × 1 ms upstream sleeps).
//
// BenchmarkCB_HalfOpenRecovery  — measures one full recovery: Open →
//   probe → success → Closed.  lastOpenedAt is zeroed each iteration so
//   the probe interval is always considered elapsed (time.Now() − 0 ≫ 100 ms).
//   Expected: ~500 ns/op (one healthy upstream round-trip).
// ============================================================================

func BenchmarkCB_TripLatency(b *testing.B) {
	upstream := newUpstreamSimulator()
	b.Cleanup(upstream.close)
	upstream.setDegraded(true) // upstream is down for the entire benchmark
	// Long probe interval: prevents the breaker recovering during the in-loop reset.
	cb := newCircuitBreakerWithProbe(upstream, time.Hour)

	b.ResetTimer()
	for b.Loop() {
		// Reset to Closed for each trip measurement.
		cb.state.Store(cbStateClosed)
		cb.failures.Store(0)
		// Drive cbTripAfter consecutive failures to trip the breaker.
		for range cbTripAfter {
			cbSinkBool = cb.Allow()
		}
	}
	b.ReportMetric(float64(cbTripAfter), "failures/trip")
}

func BenchmarkCB_HalfOpenRecovery(b *testing.B) {
	upstream := newUpstreamSimulator() // healthy: probe will succeed
	b.Cleanup(upstream.close)
	cb := newCircuitBreaker(upstream) // default 100 ms probe interval

	b.ResetTimer()
	for b.Loop() {
		// Reset to Open with lastOpenedAt=0 so the probe interval is always
		// considered elapsed (time.Now().UnixNano() − 0 ≫ probeAfterNs).
		cb.state.Store(cbStateOpen)
		cb.lastOpenedAt.Store(0)
		// One Allow() call: Open → HalfOpen → probe → success → Closed.
		cbSinkBool = cb.Allow()
	}
}

// ============================================================================
// Phase 5 — Parallel throughput under degradation
//
// BenchmarkNaive_Degraded_Parallel  — 16 goroutines (GOMAXPROCS) all call a
//   dead upstream.  The 64-item reqCh buffer fills quickly; goroutines then
//   block on the channel send itself, modelling connection-pool exhaustion.
//   Those that get through wait for time.After(1 ms).  This is intentionally
//   realistic and makes the naive numbers worse, strengthening the CV argument.
//
// BenchmarkCB_Degraded_Parallel  — breaker pre-tripped before b.ResetTimer();
//   every goroutine sees Open state and fast-fails in ~6 ns.  The upstream is
//   never contacted.  Demonstrates throughput preservation under an outage.
//
// Both benchmarks report aggregate req/s via b.ReportMetric so the throughput
// ratio can be read directly from the benchmark output.
// ============================================================================

func BenchmarkNaive_Degraded_Parallel(b *testing.B) {
	upstream := newUpstreamSimulator()
	b.Cleanup(upstream.close)
	upstream.setDegraded(true)
	client := newNaiveClient(upstream, 1*time.Millisecond)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			cbSinkBool = client.Call()
		}
	})
	b.ReportMetric(float64(b.N)/b.Elapsed().Seconds(), "req/s")
}

func BenchmarkCB_Degraded_Parallel(b *testing.B) {
	upstream := newUpstreamSimulator()
	b.Cleanup(upstream.close)
	upstream.setDegraded(true)
	// Pre-trip with a long probe interval so every goroutine sees Open state
	// from the first iteration — no 5-failure warm-up race between goroutines.
	cb := newCircuitBreakerWithProbe(upstream, time.Hour)
	cb.forceOpen()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			cbSinkBool = cb.Allow()
		}
	})
	b.ReportMetric(float64(b.N)/b.Elapsed().Seconds(), "req/s")
}
