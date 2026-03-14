package mybenchmark

import "testing"

// Add performs addition of two integers
func Add(a, b int) int {
	// + - * /
	return a + b
}

// BenchmarkAdd benchmarks the addition operation
func BenchmarkAdd(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Add(5, 10)
	}
}

// BenchmarkAddParallel benchmarks the addition operation in parallel
func BenchmarkAddParallel(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			Add(5, 10)
		}
	})
}
