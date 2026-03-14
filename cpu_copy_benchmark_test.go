package mybenchmark

import (
	"testing"
)

// BenchmarkCPUCopySmall benchmarks CPU copy for small data (5 bytes)
func BenchmarkCPUCopySmall(b *testing.B) {
	src := []byte("hello")
	dst := make([]byte, 5)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		copy(dst, src)
	}
}

// BenchmarkCPUCopyMedium benchmarks CPU copy for medium data (1KB)
func BenchmarkCPUCopyMedium(b *testing.B) {
	src := make([]byte, 1024)
	dst := make([]byte, 1024)
	for i := range src {
		src[i] = byte(i % 256)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		copy(dst, src)
	}
}

// BenchmarkCPUCopyLarge benchmarks CPU copy for large data (4KB)
func BenchmarkCPUCopyLarge(b *testing.B) {
	src := make([]byte, 4096)
	dst := make([]byte, 4096)
	for i := range src {
		src[i] = byte(i % 256)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		copy(dst, src)
	}
}

// BenchmarkCPUCopyVeryLarge benchmarks CPU copy for very large data (64KB)
func BenchmarkCPUCopyVeryLarge(b *testing.B) {
	src := make([]byte, 65536)
	dst := make([]byte, 65536)
	for i := range src {
		src[i] = byte(i % 256)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		copy(dst, src)
	}
}

// BenchmarkCPUCopyHuge benchmarks CPU copy for huge data (1MB)
func BenchmarkCPUCopyHuge(b *testing.B) {
	src := make([]byte, 1024*1024)
	dst := make([]byte, 1024*1024)
	for i := range src {
		src[i] = byte(i % 256)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		copy(dst, src)
	}
}

// BenchmarkCPUManualCopySmall benchmarks manual byte-by-byte copy for small data
func BenchmarkCPUManualCopySmall(b *testing.B) {
	src := []byte("hello")
	dst := make([]byte, 5)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := range src {
			dst[j] = src[j]
		}
	}
}

// BenchmarkCPUManualCopyLarge benchmarks manual byte-by-byte copy for large data
func BenchmarkCPUManualCopyLarge(b *testing.B) {
	src := make([]byte, 4096)
	dst := make([]byte, 4096)
	for i := range src {
		src[i] = byte(i % 256)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := range src {
			dst[j] = src[j]
		}
	}
}

// BenchmarkCPUAppend benchmarks using append to copy data
func BenchmarkCPUAppend(b *testing.B) {
	src := []byte("hello")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dst := make([]byte, 0, len(src))
		dst = append(dst, src...)
		_ = dst
	}
}

// BenchmarkCPUCopyWithAllocation benchmarks copy with allocation overhead
func BenchmarkCPUCopyWithAllocation(b *testing.B) {
	src := make([]byte, 4096)
	for i := range src {
		src[i] = byte(i % 256)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dst := make([]byte, len(src))
		copy(dst, src)
		_ = dst
	}
}

// BenchmarkCPUCopyPreallocated benchmarks copy without allocation overhead
func BenchmarkCPUCopyPreallocated(b *testing.B) {
	src := make([]byte, 4096)
	dst := make([]byte, 4096)
	for i := range src {
		src[i] = byte(i % 256)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		copy(dst, src)
	}
}

// BenchmarkCPUCopyMisaligned benchmarks copy with misaligned data
func BenchmarkCPUCopyMisaligned(b *testing.B) {
	buffer := make([]byte, 4097) // Intentionally odd size
	src := buffer[1:4097]        // Misaligned by 1 byte
	dst := make([]byte, 4096)
	for i := range src {
		src[i] = byte(i % 256)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		copy(dst, src)
	}
}
