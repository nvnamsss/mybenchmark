package mybenchmark

import (
	"sync"
	"testing"
)

// Phase 2: Fine-Grained Locking
// Partition shared state into N shards, each with its own lock.
// Goroutines only contend within the same shard.

// BenchmarkScalarFineGrainedCounter measures baseline single-threaded counter increment.
func BenchmarkScalarFineGrainedCounter(b *testing.B) {
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

// BenchmarkFineGrainedCounter_2G measures fine-grained locking with 2 goroutines and 2 shards.
// Each goroutine increments its own shard counter (no contention).
func BenchmarkFineGrainedCounter_2G(b *testing.B) {
	const numShards = 2
	var locks [numShards]sync.Mutex
	var counters [numShards]int64
	opsPerGoroutine := b.N / numShards

	b.ResetTimer()
	var wg sync.WaitGroup
	for g := 0; g < numShards; g++ {
		wg.Add(1)
		shardID := g
		go func() {
			defer wg.Done()
			for i := 0; i < opsPerGoroutine; i++ {
				locks[shardID].Lock()
				counters[shardID]++
				locks[shardID].Unlock()
			}
		}()
	}
	wg.Wait()
	_ = counters
}

// BenchmarkFineGrainedCounter_4G measures fine-grained locking with 4 goroutines and 4 shards.
func BenchmarkFineGrainedCounter_4G(b *testing.B) {
	const numShards = 4
	var locks [numShards]sync.Mutex
	var counters [numShards]int64
	opsPerGoroutine := b.N / numShards

	b.ResetTimer()
	var wg sync.WaitGroup
	for g := 0; g < numShards; g++ {
		wg.Add(1)
		shardID := g
		go func() {
			defer wg.Done()
			for i := 0; i < opsPerGoroutine; i++ {
				locks[shardID].Lock()
				counters[shardID]++
				locks[shardID].Unlock()
			}
		}()
	}
	wg.Wait()
	_ = counters
}

// BenchmarkFineGrainedCounter_8G measures fine-grained locking with 8 goroutines and 8 shards.
func BenchmarkFineGrainedCounter_8G(b *testing.B) {
	const numShards = 8
	var locks [numShards]sync.Mutex
	var counters [numShards]int64
	opsPerGoroutine := b.N / numShards

	b.ResetTimer()
	var wg sync.WaitGroup
	for g := 0; g < numShards; g++ {
		wg.Add(1)
		shardID := g
		go func() {
			defer wg.Done()
			for i := 0; i < opsPerGoroutine; i++ {
				locks[shardID].Lock()
				counters[shardID]++
				locks[shardID].Unlock()
			}
		}()
	}
	wg.Wait()
	_ = counters
}

// BenchmarkFineGrainedCounter_16G measures fine-grained locking with 16 goroutines and 16 shards.
func BenchmarkFineGrainedCounter_16G(b *testing.B) {
	const numShards = 16
	var locks [numShards]sync.Mutex
	var counters [numShards]int64
	opsPerGoroutine := b.N / numShards

	b.ResetTimer()
	var wg sync.WaitGroup
	for g := 0; g < numShards; g++ {
		wg.Add(1)
		shardID := g
		go func() {
			defer wg.Done()
			for i := 0; i < opsPerGoroutine; i++ {
				locks[shardID].Lock()
				counters[shardID]++
				locks[shardID].Unlock()
			}
		}()
	}
	wg.Wait()
	_ = counters
}
