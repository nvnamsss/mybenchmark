package mybenchmark

import (
	"sync"
	"testing"
)

// Phase 1: Coarse-Grained Locking
// Single global mutex protects a shared counter.
// All goroutines compete for the same lock.

// BenchmarkScalarCoarseGrainedCounter measures baseline single-threaded counter increment.
// No contention, no goroutines.
func BenchmarkScalarCoarseGrainedCounter(b *testing.B) {
	var mu sync.Mutex
	var counter int64
	b.ResetTimer()
	for b.Loop() {
		mu.Lock()
		counter++
		mu.Unlock()
	}
	_ = counter
}

// BenchmarkCoarseGrainedCounter_2G measures coarse-grained locking with 2 goroutines.
// Both goroutines compete for one global mutex.
func BenchmarkCoarseGrainedCounter_2G(b *testing.B) {
	var mu sync.Mutex
	var counter int64
	opsPerGoroutine := b.N / 2

	b.ResetTimer()
	var wg sync.WaitGroup
	for g := 0; g < 2; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < opsPerGoroutine; i++ {
				mu.Lock()
				counter++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	_ = counter
}

// BenchmarkCoarseGrainedCounter_4G measures coarse-grained locking with 4 goroutines.
func BenchmarkCoarseGrainedCounter_4G(b *testing.B) {
	var mu sync.Mutex
	var counter int64
	opsPerGoroutine := b.N / 4

	b.ResetTimer()
	var wg sync.WaitGroup
	for g := 0; g < 4; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < opsPerGoroutine; i++ {
				mu.Lock()
				counter++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	_ = counter
}

// BenchmarkCoarseGrainedCounter_8G measures coarse-grained locking with 8 goroutines.
func BenchmarkCoarseGrainedCounter_8G(b *testing.B) {
	var mu sync.Mutex
	var counter int64
	opsPerGoroutine := b.N / 8

	b.ResetTimer()
	var wg sync.WaitGroup
	for g := 0; g < 8; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < opsPerGoroutine; i++ {
				mu.Lock()
				counter++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	_ = counter
}

// BenchmarkCoarseGrainedCounter_16G measures coarse-grained locking with 16 goroutines.
func BenchmarkCoarseGrainedCounter_16G(b *testing.B) {
	var mu sync.Mutex
	var counter int64
	opsPerGoroutine := b.N / 16

	b.ResetTimer()
	var wg sync.WaitGroup
	for g := 0; g < 16; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < opsPerGoroutine; i++ {
				mu.Lock()
				counter++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	_ = counter
}
