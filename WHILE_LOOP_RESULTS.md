# While Loop Benchmark Results - Real Data

## Executive Summary

**The bottom line**: While loops themselves cost almost nothing (~0.28 ns/op). The real cost comes from what you DO inside the loop.

## Detailed Results

### 1. Pure Loop Overhead (Baseline)
```
BenchmarkEmptyWhileLoop                 0.28 ns/op      0 B/op    0 allocs/op
```
**Analysis**: An empty while loop is essentially free - modern CPUs can execute billions per second.

### 2. Latency Analysis (Per-Iteration Costs)
```
NoContention (simple counter)           0.41 ns/iteration
WithAtomicIncrement                     6.50 ns/iteration  (16x slower)
WithMutex                              13.14 ns/iteration  (32x slower)
```
**Key Insight**: Synchronization primitives multiply latency by 15-30x!

### 3. Real-World Patterns
```
BenchmarkGameLoopSimulation            110.1 ns/op      0 B/op    0 allocs/op
BenchmarkConnectionAcceptLoop           61.5 ns/op      0 B/op    0 allocs/op
BenchmarkEventLoopSimulation            72.2 ns/op      0 B/op    0 allocs/op
BenchmarkMessageQueueProcessing         86.2 ns/op     16 B/op    1 allocs/op
```
**Analysis**: 
- Game loops (updating 100 objects): ~110 ns/iteration
- Event/message processing: 60-90 ns/iteration
- All these are FAST - millions of iterations per second!

### 4. Polling vs Blocking (CRITICAL COMPARISON)
```
Blocking (channel receive)              59.4 ns/op      0 B/op    0 allocs/op
SelectWithDefault (polling)            598.9 ns/op      0 B/op    0 allocs/op
```
**VIVID INSIGHT**: Polling with `select default` is **10x slower** than blocking!
- Blocking: Efficient, Go scheduler manages it
- Polling: Burns CPU cycles checking repeatedly

## CPU Consumption Patterns

### Busy-Wait Loop (While checking a flag)
```go
for !ready.Load() {
    // No sleep - spins continuously
}
```
- **CPU Usage**: 100% of one core
- **Latency**: Microseconds (very responsive)
- **When to use**: Ultra low-latency requirements (<1μs)

### Sleep-Based Loop
```go
for !ready.Load() {
    time.Sleep(1 * time.Millisecond)
}
```
- **CPU Usage**: <1%
- **Latency**: Milliseconds (sleep duration + overhead)
- **When to use**: Background tasks, not time-critical

### Channel-Based Loop (RECOMMENDED)
```go
for msg := range channel {
    process(msg)
}
```
- **CPU Usage**: Near zero when idle, scales with load
- **Latency**: ~60 ns per message
- **When to use**: Almost always! Let Go scheduler handle it

## Performance Characteristics

| Pattern | ns/op | CPU Usage | Best For |
|---------|-------|-----------|----------|
| Empty loop | 0.28 | Minimal | Baseline reference |
| Simple counter | 0.41 | Minimal | Pure computation |
| Atomic ops | 6.50 | Low | Shared state without locks |
| Mutex locks | 13.14 | Low | Protected critical sections |
| Channel blocking | 59.4 | Very Low | Message passing (RECOMMENDED) |
| Event loop | 72.2 | Low | Event-driven apps |
| Message queue | 86.2 | Low | Queue consumers |
| Game loop | 110.1 | Medium | Real-time updates |
| Select polling | 598.9 | HIGH | Avoid this! Use blocking instead |

## Practical Recommendations

### ✅ DO:
1. **Use channels and blocking** - 10x faster than polling
2. **Keep loop bodies simple** - complexity adds cost
3. **Use atomic operations** for simple counters (6.5 ns vs 13 ns for mutex)
4. **Profile before optimizing** - measure real costs

### ❌ DON'T:
1. **Don't busy-wait** unless you need sub-microsecond latency
2. **Don't use select with default** for waiting - use blocking receives
3. **Don't add unnecessary mutexes** - use atomics when possible
4. **Don't assume loops are slow** - they're not! (0.28 ns is nothing)

## Real-World Translation

At 0.28 ns per empty loop iteration:
- **3.6 billion iterations/second** (on one core!)
- Running a game loop at 60 FPS with 100 objects: 110 ns = **no problem**
- Processing 1 million messages: 86 ms total
- Event loop handling 10K events/sec: 0.7% of CPU time

## The Verdict

**While loops don't cost much. What you do inside them does.**

Focus your optimization efforts on:
1. What synchronization you use (channels > atomics > mutexes)
2. How much work each iteration does
3. Whether you're busy-waiting vs blocking

The loop itself? It's basically free! 🚀
