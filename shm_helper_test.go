//go:build linux

package mybenchmark

import (
	"flag"
	"os"
	"sync/atomic"
	"syscall"
	"testing"
	"unsafe"
)

var shmPathFlag = flag.String("shmpath", "", "path to shared memory file")

// TestShmReader runs as a separate process to read from shared memory
func TestShmReader(t *testing.T) {
	if os.Getenv("SHM_READER_MODE") != "1" {
		t.Skip("Not in reader mode")
	}

	if *shmPathFlag == "" {
		t.Fatal("shmpath flag is required")
	}

	// Open shared memory file
	file, err := os.OpenFile(*shmPathFlag, os.O_RDWR, 0666)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	// Get file size
	stat, err := file.Stat()
	if err != nil {
		t.Fatal(err)
	}

	// Map the file into memory
	data, err := syscall.Mmap(int(file.Fd()), 0, int(stat.Size()),
		syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		t.Fatal(err)
	}
	defer syscall.Munmap(data)

	shm := (*ShmIPC)(unsafe.Pointer(&data[0]))
	buf := make([]byte, 4096)

	// Reader loop
	for {
		// Wait for data to be ready
		for atomic.LoadInt64(&shm.readyFlag) != 1 {
			// Spin wait
			// Check if flag is -1 to exit
			if atomic.LoadInt64(&shm.readyFlag) == -1 {
				return
			}
		}

		// Read data
		dataLen := atomic.LoadInt64(&shm.dataLen)
		copy(buf[:dataLen], shm.data[:dataLen])

		// Signal read complete
		atomic.StoreInt64(&shm.readyFlag, 2)
	}
}

// TestShmEcho runs as a separate process to echo data back
func TestShmEcho(t *testing.T) {
	if os.Getenv("SHM_ECHO_MODE") != "1" {
		t.Skip("Not in echo mode")
	}

	if *shmPathFlag == "" {
		t.Fatal("shmpath flag is required")
	}

	// Open shared memory file
	file, err := os.OpenFile(*shmPathFlag, os.O_RDWR, 0666)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	// Get file size
	stat, err := file.Stat()
	if err != nil {
		t.Fatal(err)
	}

	// Map the file into memory
	data, err := syscall.Mmap(int(file.Fd()), 0, int(stat.Size()),
		syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		t.Fatal(err)
	}
	defer syscall.Munmap(data)

	shm := (*ShmIPC)(unsafe.Pointer(&data[0]))
	buf := make([]byte, 4096)

	// Echo loop
	for {
		// Wait for data to be ready
		for atomic.LoadInt64(&shm.readyFlag) != 1 {
			// Spin wait
			// Check if flag is -1 to exit
			if atomic.LoadInt64(&shm.readyFlag) == -1 {
				return
			}
		}

		// Read data
		dataLen := atomic.LoadInt64(&shm.dataLen)
		copy(buf[:dataLen], shm.data[:dataLen])

		// Echo back (data already in place, just signal)
		atomic.StoreInt64(&shm.readyFlag, 2)
	}
}
