package mybenchmark

import (
	"sync"
	"testing"
)

// BenchmarkMutexLockUnlockOnly measures just the cost of Lock() + Unlock()
// This isolates the fundamental synchronization cost
func BenchmarkMutexLockUnlockOnly(b *testing.B) {
	var mu sync.Mutex
	b.ResetTimer()
	for b.Loop() {
		mu.Lock()
		mu.Unlock()
	}
}

// BenchmarkMutexLockUnlockWithMemOp measures Lock + trivial memory op + Unlock
// This simulates protecting a simple variable access
func BenchmarkMutexLockUnlockWithMemOp(b *testing.B) {
	var mu sync.Mutex
	var x int
	b.ResetTimer()
	for b.Loop() {
		mu.Lock()
		x++ // trivial memory operation
		mu.Unlock()
	}
	_ = x
}

// BenchmarkNoMutexMemOp measures baseline: just the memory operation without lock
// Shows what we're comparing against
func BenchmarkNoMutexMemOp(b *testing.B) {
	var x int
	b.ResetTimer()
	for b.Loop() {
		x++
	}
	_ = x
}

// BenchmarkAtomicVsMutex compares atomic.AddInt32 vs mutex-protected increment
// Atomic operations are lock-free primitives
func BenchmarkAtomicAddInt32(b *testing.B) {
	var x int32
	var a sync.Mutex
	b.ResetTimer()
	for b.Loop() {
		a.Lock()
		x++
		a.Unlock()
	}
	_ = x
}

// BenchmarkMutexContentionSimulated simulates 2 goroutines fighting over same lock
// Shows contention cost
func BenchmarkMutexContentionSimulated(b *testing.B) {
	var mu sync.Mutex
	var x int

	opsPerGoroutine := b.N / 2
	b.ResetTimer()

	var wg sync.WaitGroup
	for g := 0; g < 2; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < opsPerGoroutine; i++ {
				mu.Lock()
				x++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	_ = x
}
