#!/bin/bash
# While Loop Benchmark Commands

echo "=========================================="
echo "While Loop Benchmark Test Suite"
echo "=========================================="
echo ""

# Basic benchmark run
echo "1. Running basic benchmarks..."
go test -bench=BenchmarkEmpty -benchmem

echo ""
echo "2. Running latency analysis..."
go test -bench=BenchmarkLoopLatency -benchmem

echo ""
echo "3. Running real-world patterns..."
go test -bench="BenchmarkGameLoop|BenchmarkMessageQueue|BenchmarkEvent|BenchmarkConnection" -benchmem

echo ""
echo "4. Running polling vs blocking comparison..."
go test -bench=BenchmarkPollingVsBlocking -benchmem

echo ""
echo "=========================================="
echo "Additional commands you can run:"
echo "=========================================="
echo ""
echo "# Run all benchmarks:"
echo "go test -bench=. -benchmem"
echo ""
echo "# Run with CPU profiling:"
echo "go test -bench=BenchmarkSpinLock -cpuprofile=cpu.prof"
echo "go tool pprof cpu.prof"
echo ""
echo "# Run longer for more accurate results:"
echo "go test -bench=. -benchmem -benchtime=5s"
echo ""
echo "# Run multiple times and compare:"
echo "go test -bench=. -benchmem -count=5"
echo ""
