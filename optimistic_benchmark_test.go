package mybenchmark

import (
	"sync"
	"sync/atomic"
	"testing"
)

// Phase 3: Optimistic Synchronization
// Use compare-and-swap (CAS) in a retry loop to update shared counter.
// No explicit locks; optimistic reads and atomic updates.

// BenchmarkScalarOptimisticCounter measures baseline single-threaded counter increment.
func BenchmarkScalarOptimisticCounter(b *testing.B) {
	var counter int64
	b.ResetTimer()
	for b.Loop() {
		for {
			old := atomic.LoadInt64(&counter)
			if atomic.CompareAndSwapInt64(&counter, old, old+1) {
				break
			}
		}
	}
	_ = counter
}

// BenchmarkOptimisticCounter_2G measures optimistic synchronization with 2 goroutines.
func BenchmarkOptimisticCounter_2G(b *testing.B) {
	var counter int64
	opsPerGoroutine := b.N / 2

	b.ResetTimer()
	var wg sync.WaitGroup
	for g := 0; g < 2; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < opsPerGoroutine; i++ {
				for {
					old := atomic.LoadInt64(&counter)
					if atomic.CompareAndSwapInt64(&counter, old, old+1) {
						break
					}
				}
			}
		}()
	}
	wg.Wait()
	_ = counter
}

// BenchmarkOptimisticCounter_4G measures optimistic synchronization with 4 goroutines.
func BenchmarkOptimisticCounter_4G(b *testing.B) {
	var counter int64
	opsPerGoroutine := b.N / 4

	b.ResetTimer()
	var wg sync.WaitGroup
	for g := 0; g < 4; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < opsPerGoroutine; i++ {
				for {
					old := atomic.LoadInt64(&counter)
					if atomic.CompareAndSwapInt64(&counter, old, old+1) {
						break
					}
				}
			}
		}()
	}
	wg.Wait()
	_ = counter
}

// BenchmarkOptimisticCounter_8G measures optimistic synchronization with 8 goroutines.
func BenchmarkOptimisticCounter_8G(b *testing.B) {
	var counter int64
	opsPerGoroutine := b.N / 8

	b.ResetTimer()
	var wg sync.WaitGroup
	for g := 0; g < 8; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < opsPerGoroutine; i++ {
				for {
					old := atomic.LoadInt64(&counter)
					if atomic.CompareAndSwapInt64(&counter, old, old+1) {
						break
					}
				}
			}
		}()
	}
	wg.Wait()
	_ = counter
}

// BenchmarkOptimisticCounter_16G measures optimistic synchronization with 16 goroutines.
func BenchmarkOptimisticCounter_16G(b *testing.B) {
	var counter int64
	opsPerGoroutine := b.N / 16

	b.ResetTimer()
	var wg sync.WaitGroup
	for g := 0; g < 16; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < opsPerGoroutine; i++ {
				for {
					old := atomic.LoadInt64(&counter)
					if atomic.CompareAndSwapInt64(&counter, old, old+1) {
						break
					}
				}
			}
		}()
	}
	wg.Wait()
	_ = counter
}
