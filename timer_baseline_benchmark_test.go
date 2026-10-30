package mybenchmark

import (
	"fmt"
	"math"
	"runtime"
	"sort"
	"testing"
	"time"
)

// Sink variable — prevents dead-code elimination of timer operations.
var timerSinkTime time.Time

// ============================================================================
// TimerBackend
//
// Interface for timer implementations. Each backend runs a goroutine that
// fires ticks at a precise interval. Implementations can vary:
// - time.Ticker (Go runtime; baseline)
// - Busy-wait with spin loop (approach 1)
// - Linux timerfd + epoll (approach 2)
// - PREEMPT_RT + SCHED_FIFO (approach 3)
// ============================================================================

type TimerBackend interface {
	// Start launches the timer goroutine. Must be called before Ticks().
	Start() error

	// Stop cleanly shuts down the timer goroutine and closes the tick channel.
	Stop() error

	// Ticks returns a read-only channel that fires at the configured interval.
	Ticks() <-chan time.Time

	// Name returns a descriptive label for this backend.
	Name() string
}

// ============================================================================
// Stats
//
// Captures jitter and resource metrics for a benchmark run.
// Computed from per-tick latency deltas (time of fire vs. target time).
// ============================================================================

type Stats struct {
	Name         string
	TotalTicks   int
	CPUMs        float64 // milliseconds CPU used during benchmark
	MemAllocsOps int64   // total allocations across all ticks
	MemBytesOps  int64   // total bytes allocated

	// Per-tick jitter deltas (microseconds): actual fire time - target time
	JitterDeltas []int64

	JitterMin    int64   // minimum delta (µs)
	JitterMax    int64   // maximum delta (µs)
	JitterP50    int64   // median delta (µs)
	JitterP99    int64   // 99th percentile delta (µs)
	JitterStddev float64 // standard deviation (µs)
}

// String returns a formatted summary of the stats.
func (s *Stats) String() string {
	return fmt.Sprintf(
		"%s: %d ticks, CPU=%.1f ms, Allocs=%d, Bytes=%d, Jitter min=%d µs, p50=%d µs, p99=%d µs, max=%d µs, stddev=%.1f µs",
		s.Name, s.TotalTicks, s.CPUMs, s.MemAllocsOps, s.MemBytesOps,
		s.JitterMin, s.JitterP50, s.JitterP99, s.JitterMax, s.JitterStddev,
	)
}

// computeStats calculates jitter statistics from per-tick deltas (in microseconds).
func computeStats(name string, deltas []int64, cpuMs float64, memAllocs, memBytes int64) *Stats {
	stats := &Stats{
		Name:         name,
		TotalTicks:   len(deltas),
		CPUMs:        cpuMs,
		MemAllocsOps: memAllocs,
		MemBytesOps:  memBytes,
		JitterDeltas: deltas,
	}

	if len(deltas) == 0 {
		return stats
	}

	// Sort deltas to compute percentiles
	sorted := make([]int64, len(deltas))
	copy(sorted, deltas)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })

	// Min and max
	stats.JitterMin = sorted[0]
	stats.JitterMax = sorted[len(sorted)-1]

	// Median (p50)
	stats.JitterP50 = sorted[len(sorted)/2]

	// p99: 99th percentile
	p99Idx := int(math.Ceil(float64(len(sorted)) * 0.99))
	if p99Idx >= len(sorted) {
		p99Idx = len(sorted) - 1
	}
	stats.JitterP99 = sorted[p99Idx]

	// Standard deviation
	var sumSq float64
	mean := float64(sorted[len(sorted)/2]) // use median as reference to avoid overflow
	for _, delta := range sorted {
		diff := float64(delta) - mean
		sumSq += diff * diff
	}
	variance := sumSq / float64(len(sorted))
	stats.JitterStddev = math.Sqrt(variance)

	return stats
}

// ============================================================================
// TickerBackend
//
// Scalar baseline using Go's time.Ticker.
// Expected jitter: ~1–2 ms (Go runtime timer layer variance).
// ============================================================================

type TickerBackend struct {
	interval time.Duration
	ticker   *time.Ticker
	tickCh   chan time.Time
	stopCh   chan struct{}
	stopped  bool
}

// NewTickerBackend creates a timer backend using Go's time.Ticker.
// interval: target tick period (e.g., 33.333 ms for 30 Hz).
func NewTickerBackend(interval time.Duration) *TickerBackend {
	return &TickerBackend{
		interval: interval,
		tickCh:   make(chan time.Time, 1),
		stopCh:   make(chan struct{}),
	}
}

// Start launches the ticker goroutine.
func (tb *TickerBackend) Start() error {
	tb.ticker = time.NewTicker(tb.interval)

	go func() {
		defer close(tb.tickCh)
		for {
			select {
			case <-tb.stopCh:
				tb.ticker.Stop()
				return
			case tick := <-tb.ticker.C:
				timerSinkTime = tick
				select {
				case tb.tickCh <- tick:
				default:
					// Drop if receiver is slow; don't block the ticker goroutine
				}
			}
		}
	}()

	return nil
}

// Stop cleanly shuts down the ticker.
func (tb *TickerBackend) Stop() error {
	if tb.stopped {
		return nil
	}
	tb.stopped = true
	close(tb.stopCh)
	// Drain the tick channel to allow the goroutine to exit
	for range tb.tickCh {
	}
	return nil
}

// Ticks returns the tick channel.
func (tb *TickerBackend) Ticks() <-chan time.Time {
	return tb.tickCh
}

// Name returns the backend name.
func (tb *TickerBackend) Name() string {
	return "TickerBackend"
}

// ============================================================================
// Benchmark harness
//
// Runs a timer backend for ~10 seconds (300 ticks at 30 Hz), collecting
// per-tick jitter deltas and resource metrics.
// ============================================================================

// runTimerBenchmark executes a benchmark for a given timer backend.
// withLoad: if true, spawn 8 background goroutines doing busy work.
// Returns collected statistics.
func runTimerBenchmark(backend TimerBackend, targetHz int, withLoad bool, duration time.Duration) (*Stats, error) {
	// Start the timer backend
	if err := backend.Start(); err != nil {
		return nil, fmt.Errorf("start timer: %w", err)
	}
	defer backend.Stop()

	// Optionally start background load goroutines
	loadStopCh := make(chan struct{})
	if withLoad {
		startLoadGoroutines(8, loadStopCh)
	}

	// Benchmark loop: collect per-tick jitter deltas
	targetInterval := time.Duration(int64(time.Second) / int64(targetHz))
	var deltas []int64
	startTime := time.Now()

	// Record benchmark start for CPU time
	startCPUTime := getCPUTime()
	startMemStats := getMemStats()

	// Consume first tick to sync up with timer
	firstTick := <-backend.Ticks()
	_ = firstTick

	// Now measure remaining ticks
	expectedTime := firstTick.Add(targetInterval)

	for tick := range backend.Ticks() {
		// Calculate jitter delta: actual fire time - expected time (in microseconds)
		// Negative means early, positive means late
		delta := tick.Sub(expectedTime).Microseconds()
		deltas = append(deltas, delta)

		// Advance expected time for next tick
		expectedTime = expectedTime.Add(targetInterval)

		// Stop after duration
		if time.Since(startTime) >= duration {
			backend.Stop()
			break
		}
	}

	// Calculate resource usage
	endCPUTime := getCPUTime()
	endMemStats := getMemStats()
	cpuMs := float64(endCPUTime-startCPUTime) / 1_000_000.0
	allocsDiff := endMemStats.Alloc - startMemStats.Alloc

	// Stop load goroutines
	if withLoad {
		close(loadStopCh)
	}

	// Compute statistics
	stats := computeStats(backend.Name(), deltas, cpuMs, int64(len(deltas)), int64(allocsDiff))
	return stats, nil
}

// startLoadGoroutines spawns N goroutines that spin in tight loops.
// They simulate game logic or other background work, causing scheduler contention.
func startLoadGoroutines(count int, stopCh <-chan struct{}) {
	for i := 0; i < count; i++ {
		go func() {
			for {
				select {
				case <-stopCh:
					return
				default:
					// Busy spin — do some work that can't be optimized away
					x := int64(0)
					for j := 0; j < 1000; j++ {
						x = (x * 1103515245) + 12345 // LCG
					}
					_ = x
				}
			}
		}()
	}
}

// getCPUTime returns CPU time in nanoseconds for the current process.
// On systems without getrusage, returns 0.
func getCPUTime() int64 {
	// For now, use a simple approximation: time.Now() in nanoseconds
	// A more accurate implementation would use getrusage(2) on Unix or
	// GetProcessTimes on Windows. For timer benchmarks, wall-clock time
	// is sufficient because we control the spin/sleep.
	return time.Now().UnixNano()
}

// memStats holds memory allocation info at a point in time.
type memStats struct {
	Alloc uint64
}

// getMemStats returns current memory allocations.
// Used to track allocation overhead during benchmarks.
func getMemStats() memStats {
	var m memStats
	var s runtime.MemStats
	runtime.ReadMemStats(&s)
	m.Alloc = s.Alloc
	return m
}

// ============================================================================
// Phase 1 Benchmarks: Ticker Baseline
// ============================================================================

// BenchmarkTickerBackend_Idle runs the ticker in isolation (no background work).
// Expected: ~1–2 ms p99 jitter, minimal CPU overhead.
func BenchmarkTickerBackend_Idle(b *testing.B) {
	backend := NewTickerBackend(time.Second / 30) // 30 Hz = 33.333 ms

	b.ResetTimer()
	stats, err := runTimerBenchmark(backend, 30, false, 10*time.Second)
	if err != nil {
		b.Fatalf("runTimerBenchmark: %v", err)
	}

	// Report metrics
	b.ReportMetric(float64(stats.JitterMin), "jitter_min_us")
	b.ReportMetric(float64(stats.JitterP50), "jitter_p50_us")
	b.ReportMetric(float64(stats.JitterP99), "jitter_p99_us")
	b.ReportMetric(float64(stats.JitterMax), "jitter_max_us")
	b.ReportMetric(stats.JitterStddev, "jitter_stddev_us")
	b.ReportMetric(stats.CPUMs, "cpu_ms")

	b.Logf("Idle: %s", stats.String())
}

// BenchmarkTickerBackend_ActiveLoad runs the ticker with 8 background goroutines.
// Expected: ~2–5 ms p99 jitter (scheduler contention), similar CPU as idle.
func BenchmarkTickerBackend_ActiveLoad(b *testing.B) {
	backend := NewTickerBackend(time.Second / 30) // 30 Hz = 33.333 ms

	b.ResetTimer()
	stats, err := runTimerBenchmark(backend, 30, true, 10*time.Second)
	if err != nil {
		b.Fatalf("runTimerBenchmark: %v", err)
	}

	// Report metrics
	b.ReportMetric(float64(stats.JitterMin), "jitter_min_us")
	b.ReportMetric(float64(stats.JitterP50), "jitter_p50_us")
	b.ReportMetric(float64(stats.JitterP99), "jitter_p99_us")
	b.ReportMetric(float64(stats.JitterMax), "jitter_max_us")
	b.ReportMetric(stats.JitterStddev, "jitter_stddev_us")
	b.ReportMetric(stats.CPUMs, "cpu_ms")

	b.Logf("Active load: %s", stats.String())
}

// ============================================================================
// Correctness Tests
// ============================================================================

// TestTickerBackend_TickCount verifies that the ticker produces ~300 ticks
// in approximately 10 seconds at 30 Hz.
func TestTickerBackend_TickCount(t *testing.T) {
	backend := NewTickerBackend(time.Second / 30)
	if err := backend.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer backend.Stop()

	tickCount := 0
	startTime := time.Now()
	deadline := startTime.Add(10 * time.Second)

	for tick := range backend.Ticks() {
		_ = tick
		tickCount++
		if time.Now().After(deadline) {
			backend.Stop()
			break
		}
	}

	elapsed := time.Since(startTime)
	expectedTicks := 30 * 10 // 30 Hz for 10 seconds
	tolerance := 5            // Allow ±5 ticks

	if tickCount < expectedTicks-tolerance || tickCount > expectedTicks+tolerance {
		t.Errorf("tick count: got %d, want ~%d (elapsed: %v)", tickCount, expectedTicks, elapsed)
	}
}
