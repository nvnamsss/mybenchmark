//go:build linux

package mybenchmark

import (
	"os"
	"sync"
	"syscall"
	"testing"
	"unsafe"
)

const (
	shmSize = 4096 // 4KB shared memory
)

// BenchmarkSharedMemoryWrite benchmarks writing to shared memory
func BenchmarkSharedMemoryWrite(b *testing.B) {
	// Create shared memory using mmap
	tmpFile, err := os.CreateTemp("", "shm_bench_*.dat")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Resize file to desired size
	if err := tmpFile.Truncate(shmSize); err != nil {
		b.Fatal(err)
	}

	// Map the file into memory
	data, err := syscall.Mmap(int(tmpFile.Fd()), 0, shmSize,
		syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		b.Fatal(err)
	}
	defer syscall.Munmap(data)

	msg := []byte("hello")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		copy(data, msg)
	}
}

// BenchmarkSharedMemoryRead benchmarks reading from shared memory
func BenchmarkSharedMemoryRead(b *testing.B) {
	// Create shared memory using mmap
	tmpFile, err := os.CreateTemp("", "shm_bench_*.dat")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Resize file to desired size
	if err := tmpFile.Truncate(shmSize); err != nil {
		b.Fatal(err)
	}

	// Map the file into memory
	data, err := syscall.Mmap(int(tmpFile.Fd()), 0, shmSize,
		syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		b.Fatal(err)
	}
	defer syscall.Munmap(data)

	// Write some data
	msg := []byte("hello")
	copy(data, msg)

	buf := make([]byte, len(msg))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		copy(buf, data[:len(msg)])
	}
}

// BenchmarkSharedMemoryRoundTrip benchmarks write + read from shared memory
func BenchmarkSharedMemoryRoundTrip(b *testing.B) {
	// Create shared memory using mmap
	tmpFile, err := os.CreateTemp("", "shm_bench_*.dat")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Resize file to desired size
	if err := tmpFile.Truncate(shmSize); err != nil {
		b.Fatal(err)
	}

	// Map the file into memory
	data, err := syscall.Mmap(int(tmpFile.Fd()), 0, shmSize,
		syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		b.Fatal(err)
	}
	defer syscall.Munmap(data)

	msg := []byte("hello")
	buf := make([]byte, len(msg))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		copy(data, msg)
		copy(buf, data[:len(msg)])
	}
}

// BenchmarkSharedMemoryWithSync benchmarks shared memory with mutex synchronization
func BenchmarkSharedMemoryWithSync(b *testing.B) {
	// Create shared memory using mmap
	tmpFile, err := os.CreateTemp("", "shm_bench_*.dat")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Resize file to desired size
	if err := tmpFile.Truncate(shmSize); err != nil {
		b.Fatal(err)
	}

	// Map the file into memory
	data, err := syscall.Mmap(int(tmpFile.Fd()), 0, shmSize,
		syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		b.Fatal(err)
	}
	defer syscall.Munmap(data)

	var mu sync.Mutex
	msg := []byte("hello")
	buf := make([]byte, len(msg))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mu.Lock()
		copy(data, msg)
		copy(buf, data[:len(msg)])
		mu.Unlock()
	}
}

// BenchmarkSharedMemoryAtomic benchmarks shared memory with atomic operations
func BenchmarkSharedMemoryAtomic(b *testing.B) {
	// Create shared memory using mmap
	tmpFile, err := os.CreateTemp("", "shm_bench_*.dat")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Resize file to desired size
	if err := tmpFile.Truncate(shmSize); err != nil {
		b.Fatal(err)
	}

	// Map the file into memory
	data, err := syscall.Mmap(int(tmpFile.Fd()), 0, shmSize,
		syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		b.Fatal(err)
	}
	defer syscall.Munmap(data)

	// Use first 8 bytes as atomic counter
	counter := (*int64)(unsafe.Pointer(&data[0]))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate atomic write and read
		*counter = int64(i)
		_ = *counter
	}
}

// BenchmarkChannelCommunication benchmarks Go channels for comparison
func BenchmarkChannelCommunication(b *testing.B) {
	ch := make(chan []byte, 1)
	msg := []byte("hello")

	go func() {
		for {
			data := <-ch
			if data == nil {
				return
			}
			ch <- data // Echo back
		}
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ch <- msg
		<-ch
	}
	ch <- nil // Signal goroutine to exit
}

// BenchmarkBufferedChannelCommunication benchmarks buffered Go channels
func BenchmarkBufferedChannelCommunication(b *testing.B) {
	ch := make(chan []byte, 100)
	msg := []byte("hello")

	go func() {
		for {
			data := <-ch
			if data == nil {
				return
			}
			ch <- data // Echo back
		}
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ch <- msg
		<-ch
	}
	ch <- nil // Signal goroutine to exit
}

// BenchmarkMemoryMapLargeWrite benchmarks writing larger data to shared memory
func BenchmarkMemoryMapLargeWrite(b *testing.B) {
	// Create shared memory using mmap
	tmpFile, err := os.CreateTemp("", "shm_bench_*.dat")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Resize file to desired size
	if err := tmpFile.Truncate(shmSize); err != nil {
		b.Fatal(err)
	}

	// Map the file into memory
	data, err := syscall.Mmap(int(tmpFile.Fd()), 0, shmSize,
		syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		b.Fatal(err)
	}
	defer syscall.Munmap(data)

	msg := make([]byte, 1024) // 1KB message
	for i := range msg {
		msg[i] = byte(i % 256)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		copy(data, msg)
	}
}
