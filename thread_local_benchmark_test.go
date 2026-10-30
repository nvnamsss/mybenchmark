package mybenchmark

import (
	"sync"
	"testing"
)

// Phase 5: Thread-Local Reduction
// Each goroutine maintains a private counter; final merge at end.
// Zero contention, zero synchronization during the loop.

// BenchmarkScalarThreadLocalCounter measures baseline single-threaded counter increment.
func BenchmarkScalarThreadLocalCounter(b *testing.B) {
	var counter int64
	b.ResetTimer()
	for b.Loop() {
		counter++
	}
	_ = counter
}

// BenchmarkThreadLocalCounter_2G measures thread-local reduction with 2 goroutines.
func BenchmarkThreadLocalCounter_2G(b *testing.B) {
	opsPerGoroutine := b.N / 2
	counters := make([]int64, 2)

	b.ResetTimer()
	var wg sync.WaitGroup
	for g := 0; g < 2; g++ {
		wg.Add(1)
		localID := g
		go func() {
			defer wg.Done()
			local := int64(0)
			for i := 0; i < opsPerGoroutine; i++ {
				local++
			}
			counters[localID] = local
		}()
	}
	wg.Wait()
	_ = counters
}

// BenchmarkThreadLocalCounter_4G measures thread-local reduction with 4 goroutines.
func BenchmarkThreadLocalCounter_4G(b *testing.B) {
	opsPerGoroutine := b.N / 4
	counters := make([]int64, 4)

	b.ResetTimer()
	var wg sync.WaitGroup
	for g := 0; g < 4; g++ {
		wg.Add(1)
		localID := g
		go func() {
			defer wg.Done()
			local := int64(0)
			for i := 0; i < opsPerGoroutine; i++ {
				local++
			}
			counters[localID] = local
		}()
	}
	wg.Wait()
	_ = counters
}

// BenchmarkThreadLocalCounter_8G measures thread-local reduction with 8 goroutines.
func BenchmarkThreadLocalCounter_8G(b *testing.B) {
	opsPerGoroutine := b.N / 8
	counters := make([]int64, 8)

	b.ResetTimer()
	var wg sync.WaitGroup
	for g := 0; g < 8; g++ {
		wg.Add(1)
		localID := g
		go func() {
			defer wg.Done()
			local := int64(0)
			for i := 0; i < opsPerGoroutine; i++ {
				local++
			}
			counters[localID] = local
		}()
	}
	wg.Wait()
	_ = counters
}

// BenchmarkThreadLocalCounter_16G measures thread-local reduction with 16 goroutines.
func BenchmarkThreadLocalCounter_16G(b *testing.B) {
	opsPerGoroutine := b.N / 16
	counters := make([]int64, 16)

	b.ResetTimer()
	var wg sync.WaitGroup
	for g := 0; g < 16; g++ {
		wg.Add(1)
		localID := g
		go func() {
			defer wg.Done()
			local := int64(0)
			for i := 0; i < opsPerGoroutine; i++ {
				local++
			}
			counters[localID] = local
		}()
	}
	wg.Wait()
	_ = counters
}
