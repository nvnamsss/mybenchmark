package mybenchmark

import (
	"os"
	"os/exec"
	"sync/atomic"
	"syscall"
	"testing"
	"time"
	"unsafe"
)

// Shared memory structure for IPC
// Layout: [ready_flag(8B)][data_len(8B)][data(4080B)]
type ShmIPC struct {
	readyFlag int64 // 0 = not ready, 1 = ready to read, 2 = read complete
	dataLen   int64 // length of data
	data      [4080]byte
}

// BenchmarkSharedMemoryTwoProcessWrite benchmarks writing from parent process
func BenchmarkSharedMemoryTwoProcessWrite(b *testing.B) {
	shmPath := "/tmp/shm_ipc_bench.dat"
	os.Remove(shmPath)

	// Create shared memory file
	file, err := os.Create(shmPath)
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(shmPath)

	// Resize to fit our structure
	shmSize := int64(unsafe.Sizeof(ShmIPC{}))
	if err := file.Truncate(shmSize); err != nil {
		b.Fatal(err)
	}
	file.Close()

	// Open for mmap
	file, err = os.OpenFile(shmPath, os.O_RDWR, 0666)
	if err != nil {
		b.Fatal(err)
	}
	defer file.Close()

	// Map the file into memory
	data, err := syscall.Mmap(int(file.Fd()), 0, int(shmSize),
		syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		b.Fatal(err)
	}
	defer syscall.Munmap(data)

	shm := (*ShmIPC)(unsafe.Pointer(&data[0]))
	msg := []byte("hello")

	// Start reader process
	cmd := exec.Command(os.Args[0], "-test.run=TestShmReader", "-shmpath="+shmPath)
	cmd.Env = append(os.Environ(), "SHM_READER_MODE=1")
	if err := cmd.Start(); err != nil {
		b.Fatal(err)
	}
	defer cmd.Process.Kill()

	// Give reader time to start
	time.Sleep(100 * time.Millisecond)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Write data
		copy(shm.data[:], msg)
		atomic.StoreInt64(&shm.dataLen, int64(len(msg)))
		atomic.StoreInt64(&shm.readyFlag, 1) // Signal ready

		// Wait for reader to finish
		for atomic.LoadInt64(&shm.readyFlag) != 2 {
			// Spin wait
		}

		// Reset for next iteration
		atomic.StoreInt64(&shm.readyFlag, 0)
	}
	b.StopTimer()

	cmd.Process.Kill()
	cmd.Wait()
}

// BenchmarkSharedMemoryTwoProcessRoundTrip benchmarks full round-trip IPC
func BenchmarkSharedMemoryTwoProcessRoundTrip(b *testing.B) {
	shmPath := "/tmp/shm_ipc_roundtrip_bench.dat"
	os.Remove(shmPath)

	// Create shared memory file
	file, err := os.Create(shmPath)
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(shmPath)

	// Resize to fit our structure
	shmSize := int64(unsafe.Sizeof(ShmIPC{}))
	if err := file.Truncate(shmSize); err != nil {
		b.Fatal(err)
	}
	file.Close()

	// Open for mmap
	file, err = os.OpenFile(shmPath, os.O_RDWR, 0666)
	if err != nil {
		b.Fatal(err)
	}
	defer file.Close()

	// Map the file into memory
	data, err := syscall.Mmap(int(file.Fd()), 0, int(shmSize),
		syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		b.Fatal(err)
	}
	defer syscall.Munmap(data)

	shm := (*ShmIPC)(unsafe.Pointer(&data[0]))
	msg := []byte("hello")
	buf := make([]byte, len(msg))

	// Start echo process
	cmd := exec.Command(os.Args[0], "-test.run=TestShmEcho", "-shmpath="+shmPath)
	cmd.Env = append(os.Environ(), "SHM_ECHO_MODE=1")
	if err := cmd.Start(); err != nil {
		b.Fatal(err)
	}
	defer cmd.Process.Kill()

	// Give echo process time to start
	time.Sleep(100 * time.Millisecond)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Write data
		copy(shm.data[:], msg)
		atomic.StoreInt64(&shm.dataLen, int64(len(msg)))
		atomic.StoreInt64(&shm.readyFlag, 1) // Signal ready for echo

		// Wait for echo process to respond
		for atomic.LoadInt64(&shm.readyFlag) != 2 {
			// Spin wait
		}

		// Read echoed data
		dataLen := atomic.LoadInt64(&shm.dataLen)
		copy(buf, shm.data[:dataLen])

		// Reset for next iteration
		atomic.StoreInt64(&shm.readyFlag, 0)
	}
	b.StopTimer()

	cmd.Process.Kill()
	cmd.Wait()
}

// BenchmarkSharedMemoryTwoProcessLargeData benchmarks large data transfer
func BenchmarkSharedMemoryTwoProcessLargeData(b *testing.B) {
	shmPath := "/tmp/shm_ipc_large_bench.dat"
	os.Remove(shmPath)

	// Create shared memory file
	file, err := os.Create(shmPath)
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(shmPath)

	// Resize to fit our structure
	shmSize := int64(unsafe.Sizeof(ShmIPC{}))
	if err := file.Truncate(shmSize); err != nil {
		b.Fatal(err)
	}
	file.Close()

	// Open for mmap
	file, err = os.OpenFile(shmPath, os.O_RDWR, 0666)
	if err != nil {
		b.Fatal(err)
	}
	defer file.Close()

	// Map the file into memory
	data, err := syscall.Mmap(int(file.Fd()), 0, int(shmSize),
		syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		b.Fatal(err)
	}
	defer syscall.Munmap(data)

	shm := (*ShmIPC)(unsafe.Pointer(&data[0]))
	msg := make([]byte, 4000) // Large message
	for i := range msg {
		msg[i] = byte(i % 256)
	}

	// Start reader process
	cmd := exec.Command(os.Args[0], "-test.run=TestShmReader", "-shmpath="+shmPath)
	cmd.Env = append(os.Environ(), "SHM_READER_MODE=1")
	if err := cmd.Start(); err != nil {
		b.Fatal(err)
	}
	defer cmd.Process.Kill()

	// Give reader time to start
	time.Sleep(100 * time.Millisecond)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Write data
		copy(shm.data[:], msg)
		atomic.StoreInt64(&shm.dataLen, int64(len(msg)))
		atomic.StoreInt64(&shm.readyFlag, 1) // Signal ready

		// Wait for reader to finish
		for atomic.LoadInt64(&shm.readyFlag) != 2 {
			// Spin wait
		}

		// Reset for next iteration
		atomic.StoreInt64(&shm.readyFlag, 0)
	}
	b.StopTimer()

	cmd.Process.Kill()
	cmd.Wait()
}
