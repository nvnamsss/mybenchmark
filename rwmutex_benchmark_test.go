package mybenchmark

import (
	"sync"
	"testing"
)

// Phase 6: RWMutex (Read-Write Locking)
// Separate read and write locks. Multiple readers can proceed concurrently.
// Writers are exclusive (readers and writers block each other).
// Optimal for read-heavy workloads.

// BenchmarkScalarRWMutexRead measures baseline single-threaded read operation.
// Single reader, no contention.
func BenchmarkScalarRWMutexRead(b *testing.B) {
	var rwmu sync.RWMutex
	var counter int64
	b.ResetTimer()
	for b.Loop() {
		rwmu.RLock()
		_ = counter
		rwmu.RUnlock()
	}
}

// BenchmarkRWMutexRead_8G measures pure read workload with 8 goroutines.
// All 8 goroutines perform reads concurrently (no writes).
// Expected: excellent scaling since RLock allows concurrent readers.
func BenchmarkRWMutexRead_8G(b *testing.B) {
	var rwmu sync.RWMutex
	var counter int64
	opsPerGoroutine := b.N / 8

	b.ResetTimer()
	var wg sync.WaitGroup
	for g := 0; g < 8; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < opsPerGoroutine; i++ {
				rwmu.RLock()
				_ = counter
				rwmu.RUnlock()
			}
		}()
	}
	wg.Wait()
	_ = counter
}

// BenchmarkRWMutexReadWrite_8G_1W measures mixed workload: 7 readers + 1 writer.
// One writer, seven readers. Writers serialize access.
func BenchmarkRWMutexReadWrite_8G_1W(b *testing.B) {
	var rwmu sync.RWMutex
	var counter int64
	opsPerGoroutine := b.N / 8

	b.ResetTimer()
	var wg sync.WaitGroup

	// 7 readers
	for g := 0; g < 7; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < opsPerGoroutine; i++ {
				rwmu.RLock()
				_ = counter
				rwmu.RUnlock()
			}
		}()
	}

	// 1 writer
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < opsPerGoroutine; i++ {
			rwmu.Lock()
			counter++
			rwmu.Unlock()
		}
	}()

	wg.Wait()
	_ = counter
}

// BenchmarkRWMutexReadWrite_8G_4W measures balanced workload: 4 readers + 4 writers.
// Four writers, four readers. More contention between readers and writers.
func BenchmarkRWMutexReadWrite_8G_4W(b *testing.B) {
	var rwmu sync.RWMutex
	var counter int64
	opsPerGoroutine := b.N / 8

	b.ResetTimer()
	var wg sync.WaitGroup

	// 4 readers
	for g := 0; g < 4; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < opsPerGoroutine; i++ {
				rwmu.RLock()
				_ = counter
				rwmu.RUnlock()
			}
		}()
	}

	// 4 writers
	for g := 0; g < 4; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < opsPerGoroutine; i++ {
				rwmu.Lock()
				counter++
				rwmu.Unlock()
			}
		}()
	}

	wg.Wait()
	_ = counter
}

// BenchmarkMutexVsRWMutex_ReadHeavy compares sync.Mutex vs sync.RWMutex
// with a read-heavy workload (90% reads, 10% writes).
// This demonstrates when RWMutex provides significant benefit.
// Implementation: 18 readers + 2 writers (out of 20 goroutines, scaled to 8 with proportional ratios)
// Simplified for 8 goroutines: 7 readers + 1 writer = 87.5% reads
func BenchmarkMutexReadHeavy_8G(b *testing.B) {
	var mu sync.Mutex
	var counter int64
	opsPerGoroutine := b.N / 8

	b.ResetTimer()
	var wg sync.WaitGroup

	// 7 readers (using mutex)
	for g := 0; g < 7; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < opsPerGoroutine; i++ {
				mu.Lock()
				_ = counter
				mu.Unlock()
			}
		}()
	}

	// 1 writer (using mutex)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < opsPerGoroutine; i++ {
			mu.Lock()
			counter++
			mu.Unlock()
		}
	}()

	wg.Wait()
	_ = counter
}

// BenchmarkRWMutexReadHeavy_8G measures the same workload with RWMutex.
// Readers use RLock (can be concurrent), writer uses Lock (exclusive).
// Expected: faster than Mutex version due to concurrent readers.
func BenchmarkRWMutexReadHeavy_8G(b *testing.B) {
	var rwmu sync.RWMutex
	var counter int64
	opsPerGoroutine := b.N / 8

	b.ResetTimer()
	var wg sync.WaitGroup

	// 7 readers (using rwmutex RLock)
	for g := 0; g < 7; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < opsPerGoroutine; i++ {
				rwmu.RLock()
				_ = counter
				rwmu.RUnlock()
			}
		}()
	}

	// 1 writer (using rwmutex Lock)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < opsPerGoroutine; i++ {
			rwmu.Lock()
			counter++
			rwmu.Unlock()
		}
	}()

	wg.Wait()
	_ = counter
}
