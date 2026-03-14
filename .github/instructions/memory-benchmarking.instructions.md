---
applyTo: "**/*_benchmark_test.go"
description: "Use when writing or modifying Go benchmarks that report memory; prevents confusion between build-time allocations and retained heap footprint."
---

# Memory Benchmarking Guardrails

When benchmarking memory, do not treat benchmark `B/op` and `allocs/op` as retained memory footprint.

Required methodology:

1. Separate benchmark intent into explicit groups:
- Throughput and latency: operation speed only
- Build cost: construction-time allocations and CPU
- Retained footprint: post-build heap retained after GC

2. For lookup benchmarks:
- Build all fixtures before `b.ResetTimer()`
- Keep probe data deterministic
- Avoid allocations inside timed loops unless that allocation is intentionally measured

3. For build benchmarks:
- Include construction in the timed loop
- Use `b.ReportAllocs()`
- Document that results are build-time allocation behavior, not resident size

4. For retained footprint benchmarks:
- Use `runtime.GC()` before and after build
- Capture `runtime.MemStats.HeapAlloc` delta
- Keep the built object alive with `runtime.KeepAlive`
- Report explicit custom metrics such as:
  - `map_heap_B`
  - `mph_heap_B`
  - ratios like `mph_to_map_heap_ratio`

5. Avoid false conclusions:
- If MPH needs auxiliary key verification for misses, report both:
  - MPH table only
  - MPH plus required auxiliary structures
- State clearly which structures are included in each metric

6. Benchmark result interpretation:
- `B/op` and `allocs/op` answer: "How expensive is construction or operation per iteration?"
- retained heap metrics answer: "How much memory remains after build and GC?"
- Never compare one of these to the other as if they are the same quantity

Checklist before finalizing benchmark changes:

- [ ] Lookup benchmark setup is outside timer
- [ ] Build benchmark is separately named and documented
- [ ] Retained heap benchmark exists for memory-footprint claims
- [ ] Metrics names distinguish build vs retained memory
- [ ] Summary text explains what each memory metric means
