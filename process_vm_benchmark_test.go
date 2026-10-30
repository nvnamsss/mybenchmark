//go:build linux

package mybenchmark

import (
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"
	"unsafe"

	"golang.org/x/sys/unix"
)

// BenchmarkProcessVmReadvSmall benchmarks process_vm_readv for small data (5 bytes)
func BenchmarkProcessVmReadvSmall(b *testing.B) {
	// Start a helper process that holds data in memory
	cmd := exec.Command(os.Args[0], "-test.run=TestVmReadvHelper")
	cmd.Env = append(os.Environ(), "VM_READV_HELPER=1")

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		b.Fatal(err)
	}

	if err := cmd.Start(); err != nil {
		b.Fatal(err)
	}
	defer cmd.Process.Kill()

	// Read the remote address from the helper process
	var remoteAddr uintptr
	var dataSize int
	_, err = fmt.Fscanf(stdout, "%d %d\n", &remoteAddr, &dataSize)
	if err != nil {
		b.Fatal(err)
	}

	pid := cmd.Process.Pid
	localBuf := make([]byte, 5)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		localIov := []unix.Iovec{{
			Base: &localBuf[0],
			Len:  uint64(len(localBuf)),
		}}

		remoteIov := []unix.RemoteIovec{{
			Base: remoteAddr,
			Len:  len(localBuf),
		}}

		_, err := unix.ProcessVMReadv(pid, localIov, remoteIov, 0)
		if err != nil {
			b.Fatal(err)
		}
	}
	b.StopTimer()

	cmd.Process.Kill()
	cmd.Wait()
}

// BenchmarkProcessVmReadvLarge benchmarks process_vm_readv for large data (4KB)
func BenchmarkProcessVmReadvLarge(b *testing.B) {
	// Start a helper process that holds data in memory
	cmd := exec.Command(os.Args[0], "-test.run=TestVmReadvHelper")
	cmd.Env = append(os.Environ(), "VM_READV_HELPER=1")

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		b.Fatal(err)
	}

	if err := cmd.Start(); err != nil {
		b.Fatal(err)
	}
	defer cmd.Process.Kill()

	// Read the remote address from the helper process
	var remoteAddr uintptr
	var dataSize int
	_, err = fmt.Fscanf(stdout, "%d %d\n", &remoteAddr, &dataSize)
	if err != nil {
		b.Fatal(err)
	}

	pid := cmd.Process.Pid
	localBuf := make([]byte, 4096)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		localIov := []unix.Iovec{{
			Base: &localBuf[0],
			Len:  uint64(len(localBuf)),
		}}

		remoteIov := []unix.RemoteIovec{{
			Base: remoteAddr,
			Len:  len(localBuf),
		}}

		_, err := unix.ProcessVMReadv(pid, localIov, remoteIov, 0)
		if err != nil {
			b.Fatal(err)
		}
	}
	b.StopTimer()

	cmd.Process.Kill()
	cmd.Wait()
}

// BenchmarkProcessVmWritevSmall benchmarks process_vm_writev for small data (5 bytes)
func BenchmarkProcessVmWritevSmall(b *testing.B) {
	// Start a helper process that holds data in memory
	cmd := exec.Command(os.Args[0], "-test.run=TestVmWritevHelper")
	cmd.Env = append(os.Environ(), "VM_WRITEV_HELPER=1")

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		b.Fatal(err)
	}

	if err := cmd.Start(); err != nil {
		b.Fatal(err)
	}
	defer cmd.Process.Kill()

	// Read the remote address from the helper process
	var remoteAddr uintptr
	var dataSize int
	_, err = fmt.Fscanf(stdout, "%d %d\n", &remoteAddr, &dataSize)
	if err != nil {
		b.Fatal(err)
	}

	pid := cmd.Process.Pid
	localBuf := []byte("hello")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		localIov := []unix.Iovec{{
			Base: &localBuf[0],
			Len:  uint64(len(localBuf)),
		}}

		remoteIov := []unix.RemoteIovec{{
			Base: remoteAddr,
			Len:  len(localBuf),
		}}

		_, err := unix.ProcessVMWritev(pid, localIov, remoteIov, 0)
		if err != nil {
			b.Fatal(err)
		}
	}
	b.StopTimer()

	cmd.Process.Kill()
	cmd.Wait()
}

// BenchmarkProcessVmRoundTrip benchmarks write + read round-trip
func BenchmarkProcessVmRoundTrip(b *testing.B) {
	// Start a helper process that holds data in memory
	cmd := exec.Command(os.Args[0], "-test.run=TestVmWritevHelper")
	cmd.Env = append(os.Environ(), "VM_WRITEV_HELPER=1")

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		b.Fatal(err)
	}

	if err := cmd.Start(); err != nil {
		b.Fatal(err)
	}
	defer cmd.Process.Kill()

	// Read the remote address from the helper process
	var remoteAddr uintptr
	var dataSize int
	_, err = fmt.Fscanf(stdout, "%d %d\n", &remoteAddr, &dataSize)
	if err != nil {
		b.Fatal(err)
	}

	pid := cmd.Process.Pid
	writeBuf := []byte("hello")
	readBuf := make([]byte, 5)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Write
		localIovWrite := []unix.Iovec{{
			Base: &writeBuf[0],
			Len:  uint64(len(writeBuf)),
		}}
		remoteIovWrite := []unix.RemoteIovec{{
			Base: remoteAddr,
			Len:  len(writeBuf),
		}}
		_, err := unix.ProcessVMWritev(pid, localIovWrite, remoteIovWrite, 0)
		if err != nil {
			b.Fatal(err)
		}

		// Read
		localIovRead := []unix.Iovec{{
			Base: &readBuf[0],
			Len:  uint64(len(readBuf)),
		}}
		remoteIovRead := []unix.RemoteIovec{{
			Base: remoteAddr,
			Len:  len(readBuf),
		}}
		_, err = unix.ProcessVMReadv(pid, localIovRead, remoteIovRead, 0)
		if err != nil {
			b.Fatal(err)
		}
	}
	b.StopTimer()

	cmd.Process.Kill()
	cmd.Wait()
}

// TestVmReadvHelper runs as a separate process holding data in memory
func TestVmReadvHelper(t *testing.T) {
	if os.Getenv("VM_READV_HELPER") != "1" {
		t.Skip("Not in VM readv helper mode")
	}

	// Allocate data buffer
	data := []byte("hello world this is test data")
	dataAddr := uintptr(unsafe.Pointer(&data[0]))

	// Print the address and size to stdout
	fmt.Printf("%d %d\n", dataAddr, len(data))

	// Keep the process alive
	time.Sleep(10 * time.Minute)
}

// TestVmWritevHelper runs as a separate process with writable memory
func TestVmWritevHelper(t *testing.T) {
	if os.Getenv("VM_WRITEV_HELPER") != "1" {
		t.Skip("Not in VM writev helper mode")
	}

	// Allocate writable data buffer
	data := make([]byte, 8192)
	dataAddr := uintptr(unsafe.Pointer(&data[0]))

	// Print the address and size to stdout
	fmt.Printf("%d %d\n", dataAddr, len(data))

	// Keep the process alive
	time.Sleep(10 * time.Minute)
}
