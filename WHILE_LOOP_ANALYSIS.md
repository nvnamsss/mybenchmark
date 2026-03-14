# While Loop Cost Analysis

## Quick Summary of Key Findings

### 1. Basic Loop Overhead (Extremely Low)
- **Empty while loop**: ~0.30 ns/op
- **Simple condition check**: ~0.66 ns/op  
- **Atomic read per iteration**: ~0.35 ns/op

**Insight**: The pure cost of a while loop is negligible - less than 1 nanosecond per iteration.

### 2. Latency Per Operation
- **No contention**: 0.35 ns/iteration
- **Atomic increment**: 6.35 ns/iteration (18x slower)
- **Mutex lock/unlock**: 14.50 ns/iteration (42x slower)

**Insight**: The loop itself is cheap. The COST comes from synchronization primitives inside the loop.

### 3. Real-World Patterns
- **Game loop** (100 objects): ~121 ns/op
- **Message queue processing**: ~98 ns/op + 16 B/op
- **SpinLock acquisition** (contended): ~183 ns/op

**Insight**: Real applications show moderate costs due to actual work being done, not the loop structure.

## Running the Benchmarks

### Basic run:
```bash
go test -bench=. -benchmem
```

### With CPU profiling:
```bash
go test -bench=BenchmarkSpinLock -cpuprofile=cpu.prof
go tool pprof cpu.prof
```

### Detailed analysis:
```bash
go test -bench=. -benchmem -benchtime=5s -count=5
```

## What Each Benchmark Reveals

1. **BenchmarkEmptyWhileLoop**: Pure loop overhead
2. **BenchmarkWhileLoopBusyWait**: CPU cost of busy-waiting (reports spins/op)
3. **BenchmarkWhileLoopWithSleep**: Trade-off between CPU usage and latency
4. **BenchmarkGameLoopSimulation**: Real-time game loop pattern
5. **BenchmarkMessageQueueProcessing**: Message queue consumer pattern
6. **BenchmarkPollingVsBlocking**: Direct comparison of approaches
7. **BenchmarkLoopLatency**: Detailed latency analysis with different sync methods

## Key Takeaways

1. **While loops are NOT expensive** - the overhead is < 1 ns/op
2. **The work inside the loop matters** - synchronization, I/O, computation
3. **Busy-waiting wastes CPU** - use channels/blocking when possible
4. **Atomic operations add ~6ns** overhead per operation
5. **Mutexes add ~15ns** overhead per lock/unlock pair
6. **Real applications**: 100-200 ns/op is typical for event/message processing

## CPU Usage Patterns

- **Busy wait loop**: 100% CPU usage, lowest latency
- **Sleep-based loop**: <1% CPU usage, higher latency
- **Channel-based loop**: Balanced CPU usage, Go scheduler managed
- **Blocking I/O loop**: Near-zero CPU when idle

The choice depends on your latency vs CPU usage trade-off!
