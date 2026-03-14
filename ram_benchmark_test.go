package mybenchmark

import (
	"sync/atomic"
	"testing"
	"unsafe"
)

// BenchmarkRAMReadSequential benchmarks sequential RAM reads
func BenchmarkRAMReadSequential(b *testing.B) {
	size := 1024 * 1024 // 1MB
	data := make([]byte, size)
	for i := range data {
		data[i] = byte(i % 256)
	}

	b.ResetTimer()
	var sum byte
	for i := 0; i < b.N; i++ {
		for j := 0; j < len(data); j++ {
			sum += data[j]
		}
	}
	_ = sum
}

// BenchmarkRAMWriteSequential benchmarks sequential RAM writes
func BenchmarkRAMWriteSequential(b *testing.B) {
	size := 1024 * 1024 // 1MB
	data := make([]byte, size)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < len(data); j++ {
			data[j] = byte(j % 256)
		}
	}
}

// BenchmarkRAMReadRandom benchmarks random RAM reads
func BenchmarkRAMReadRandom(b *testing.B) {
	size := 1024 * 1024 // 1MB
	data := make([]byte, size)
	for i := range data {
		data[i] = byte(i % 256)
	}

	// Create random access pattern
	indices := make([]int, 10000)
	for i := range indices {
		indices[i] = (i * 4096) % size
	}

	b.ResetTimer()
	var sum byte
	for i := 0; i < b.N; i++ {
		idx := indices[i%len(indices)]
		sum += data[idx]
	}
	_ = sum
}

// BenchmarkRAMAtomicRead benchmarks atomic read operations
func BenchmarkRAMAtomicRead(b *testing.B) {
	var value int64 = 42

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = atomic.LoadInt64(&value)
	}
}

// BenchmarkRAMAtomicWrite benchmarks atomic write operations
func BenchmarkRAMAtomicWrite(b *testing.B) {
	var value int64

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		atomic.StoreInt64(&value, int64(i))
	}
}

// BenchmarkRAMAllocation benchmarks memory allocation
func BenchmarkRAMAllocation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		data := make([]byte, 4096)
		_ = data
	}
}

// BenchmarkRAMBandwidthTest benchmarks memory bandwidth
func BenchmarkRAMBandwidthTest(b *testing.B) {
	size := 64 * 1024 * 1024 // 64MB - larger than cache
	src := make([]byte, size)
	dst := make([]byte, size)

	for i := range src {
		src[i] = byte(i % 256)
	}

	b.SetBytes(int64(size))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		copy(dst, src)
	}
}

// BenchmarkRAMAlignedAccess benchmarks aligned memory access
func BenchmarkRAMAlignedAccess(b *testing.B) {
	size := 1024 * 1024
	data := make([]int64, size)
	for i := range data {
		data[i] = int64(i)
	}

	b.ResetTimer()
	var sum int64
	for i := 0; i < b.N; i++ {
		for j := 0; j < len(data); j++ {
			sum += data[j]
		}
	}
	_ = sum
}

// BenchmarkRAMUnalignedAccess benchmarks unaligned memory access
func BenchmarkRAMUnalignedAccess(b *testing.B) {
	size := 1024*1024 + 1
	data := make([]byte, size)
	for i := range data {
		data[i] = byte(i % 256)
	}

	b.ResetTimer()
	var sum int64
	for i := 0; i < b.N; i++ {
		for j := 0; j < len(data)-8; j++ {
			// Unaligned read of int64
			value := *(*int64)(unsafe.Pointer(&data[j]))
			sum += value
		}
	}
	_ = sum
}
