package mybenchmark

import (
	"sync"
	"sync/atomic"
	"testing"
)

// Phase 4: Lock-Free Synchronization
// Use atomic.AddInt64 for lock-free counter increment.
// No locks, no retries; single atomic operation per increment.

// BenchmarkScalarLockFreeCounter measures baseline single-threaded counter increment.
func BenchmarkScalarLockFreeCounter(b *testing.B) {
	var counter int64
	b.ResetTimer()
	for b.Loop() {
		atomic.AddInt64(&counter, 1)
	}
	_ = counter
}

// BenchmarkLockFreeCounter_2G measures lock-free synchronization with 2 goroutines.
func BenchmarkLockFreeCounter_2G(b *testing.B) {
	var counter int64
	opsPerGoroutine := b.N / 2

	b.ResetTimer()
	var wg sync.WaitGroup
	for g := 0; g < 2; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < opsPerGoroutine; i++ {
				atomic.AddInt64(&counter, 1)
			}
		}()
	}
	wg.Wait()
	_ = counter
}

// BenchmarkLockFreeCounter_4G measures lock-free synchronization with 4 goroutines.
func BenchmarkLockFreeCounter_4G(b *testing.B) {
	var counter int64
	opsPerGoroutine := b.N / 4

	b.ResetTimer()
	var wg sync.WaitGroup
	for g := 0; g < 4; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < opsPerGoroutine; i++ {
				atomic.AddInt64(&counter, 1)
			}
		}()
	}
	wg.Wait()
	_ = counter
}

// BenchmarkLockFreeCounter_8G measures lock-free synchronization with 8 goroutines.
func BenchmarkLockFreeCounter_8G(b *testing.B) {
	var counter int64
	opsPerGoroutine := b.N / 8

	b.ResetTimer()
	var wg sync.WaitGroup
	for g := 0; g < 8; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < opsPerGoroutine; i++ {
				atomic.AddInt64(&counter, 1)
			}
		}()
	}
	wg.Wait()
	_ = counter
}

// BenchmarkLockFreeCounter_16G measures lock-free synchronization with 16 goroutines.
func BenchmarkLockFreeCounter_16G(b *testing.B) {
	var counter int64
	opsPerGoroutine := b.N / 16

	b.ResetTimer()
	var wg sync.WaitGroup
	for g := 0; g < 16; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < opsPerGoroutine; i++ {
				atomic.AddInt64(&counter, 1)
			}
		}()
	}
	wg.Wait()
	_ = counter
}
