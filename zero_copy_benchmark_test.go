//go:build linux
// +build linux

package mybenchmark

import (
	"io"
	"net"
	"os"
	"testing"

	"golang.org/x/sys/unix"
)

// Helper function to create a temporary file with data
func createTempFileWithData(b *testing.B, size int) (*os.File, string) {
	tmpFile, err := os.CreateTemp("", "zerocopy_bench_*.dat")
	if err != nil {
		b.Fatal(err)
	}

	data := make([]byte, size)
	for i := range data {
		data[i] = byte(i % 256)
	}

	if _, err := tmpFile.Write(data); err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		b.Fatal(err)
	}

	if _, err := tmpFile.Seek(0, 0); err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		b.Fatal(err)
	}

	return tmpFile, tmpFile.Name()
}

// Helper function to setup TCP connection
func setupTCPConnection(b *testing.B) (net.Listener, net.Conn, net.Conn) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		b.Fatal(err)
	}

	// Accept connection in goroutine
	serverConnCh := make(chan net.Conn, 1)
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		serverConnCh <- conn
	}()

	// Connect client
	clientConn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		listener.Close()
		b.Fatal(err)
	}

	serverConn := <-serverConnCh

	return listener, serverConn, clientConn
}

// BenchmarkSendfileFileToSocket benchmarks sendfile syscall for file-to-socket transfer
func BenchmarkSendfileFileToSocket(b *testing.B) {
	// Create temp file with 64KB data
	tmpFile, tmpPath := createTempFileWithData(b, 64*1024)
	defer os.Remove(tmpPath)
	defer tmpFile.Close()

	// Setup TCP connection
	listener, serverConn, clientConn := setupTCPConnection(b)
	defer listener.Close()
	defer serverConn.Close()
	defer clientConn.Close()

	// Server goroutine - discard received data
	go func() {
		io.Copy(io.Discard, serverConn)
	}()

	// Get client socket file descriptor
	clientFile, err := clientConn.(*net.TCPConn).File()
	if err != nil {
		b.Fatal(err)
	}
	defer clientFile.Close()

	clientFd := int(clientFile.Fd())
	fileFd := int(tmpFile.Fd())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Reset file position
		if _, err := tmpFile.Seek(0, 0); err != nil {
			b.Fatal(err)
		}

		// Use sendfile to transfer data
		_, err := unix.Sendfile(clientFd, fileFd, nil, 64*1024)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkSendfileSmallFile benchmarks sendfile with small file (4KB)
func BenchmarkSendfileSmallFile(b *testing.B) {
	tmpFile, tmpPath := createTempFileWithData(b, 4*1024)
	defer os.Remove(tmpPath)
	defer tmpFile.Close()

	listener, serverConn, clientConn := setupTCPConnection(b)
	defer listener.Close()
	defer serverConn.Close()
	defer clientConn.Close()

	go func() {
		io.Copy(io.Discard, serverConn)
	}()

	clientFile, err := clientConn.(*net.TCPConn).File()
	if err != nil {
		b.Fatal(err)
	}
	defer clientFile.Close()

	clientFd := int(clientFile.Fd())
	fileFd := int(tmpFile.Fd())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tmpFile.Seek(0, 0)
		unix.Sendfile(clientFd, fileFd, nil, 4*1024)
	}
}

// BenchmarkSendfileLargeFile benchmarks sendfile with large file (1MB)
func BenchmarkSendfileLargeFile(b *testing.B) {
	tmpFile, tmpPath := createTempFileWithData(b, 1024*1024)
	defer os.Remove(tmpPath)
	defer tmpFile.Close()

	listener, serverConn, clientConn := setupTCPConnection(b)
	defer listener.Close()
	defer serverConn.Close()
	defer clientConn.Close()

	go func() {
		io.Copy(io.Discard, serverConn)
	}()

	clientFile, err := clientConn.(*net.TCPConn).File()
	if err != nil {
		b.Fatal(err)
	}
	defer clientFile.Close()

	clientFd := int(clientFile.Fd())
	fileFd := int(tmpFile.Fd())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tmpFile.Seek(0, 0)
		unix.Sendfile(clientFd, fileFd, nil, 1024*1024)
	}
}

// BenchmarkTraditionalCopyFileToSocket benchmarks traditional read/write for comparison
func BenchmarkTraditionalCopyFileToSocket(b *testing.B) {
	tmpFile, tmpPath := createTempFileWithData(b, 64*1024)
	defer os.Remove(tmpPath)
	defer tmpFile.Close()

	listener, serverConn, clientConn := setupTCPConnection(b)
	defer listener.Close()
	defer serverConn.Close()
	defer clientConn.Close()

	go func() {
		io.Copy(io.Discard, serverConn)
	}()

	buf := make([]byte, 32*1024)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tmpFile.Seek(0, 0)
		io.CopyBuffer(clientConn, tmpFile, buf)
	}
}

// BenchmarkSplicePipeToSocket benchmarks splice syscall for pipe-to-socket transfer
func BenchmarkSplicePipeToSocket(b *testing.B) {
	// Create pipe
	pipeFds := make([]int, 2)
	if err := unix.Pipe(pipeFds); err != nil {
		b.Fatal(err)
	}
	defer unix.Close(pipeFds[0])
	defer unix.Close(pipeFds[1])

	// Setup TCP connection
	listener, serverConn, clientConn := setupTCPConnection(b)
	defer listener.Close()
	defer serverConn.Close()
	defer clientConn.Close()

	// Server goroutine - discard received data
	go func() {
		io.Copy(io.Discard, serverConn)
	}()

	// Get client socket file descriptor
	clientFile, err := clientConn.(*net.TCPConn).File()
	if err != nil {
		b.Fatal(err)
	}
	defer clientFile.Close()

	clientFd := int(clientFile.Fd())
	data := make([]byte, 4096)
	for i := range data {
		data[i] = byte(i % 256)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Write data to pipe
		unix.Write(pipeFds[1], data)

		// Splice from pipe to socket
		_, err := unix.Splice(pipeFds[0], nil, clientFd, nil, len(data), 0)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkSpliceFileToSocket benchmarks splice for file-to-socket via pipe
func BenchmarkSpliceFileToSocket(b *testing.B) {
	tmpFile, tmpPath := createTempFileWithData(b, 64*1024)
	defer os.Remove(tmpPath)
	defer tmpFile.Close()

	// Create pipe
	pipeFds := make([]int, 2)
	if err := unix.Pipe(pipeFds); err != nil {
		b.Fatal(err)
	}
	defer unix.Close(pipeFds[0])
	defer unix.Close(pipeFds[1])

	listener, serverConn, clientConn := setupTCPConnection(b)
	defer listener.Close()
	defer serverConn.Close()
	defer clientConn.Close()

	go func() {
		io.Copy(io.Discard, serverConn)
	}()

	clientFile, err := clientConn.(*net.TCPConn).File()
	if err != nil {
		b.Fatal(err)
	}
	defer clientFile.Close()

	clientFd := int(clientFile.Fd())
	fileFd := int(tmpFile.Fd())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tmpFile.Seek(0, 0)

		// Splice file to pipe, then pipe to socket
		remaining := 64 * 1024
		for remaining > 0 {
			// Splice from file to pipe
			n, err := unix.Splice(fileFd, nil, pipeFds[1], nil, remaining, 0)
			if err != nil {
				b.Fatal(err)
			}
			if n == 0 {
				break
			}

			// Splice from pipe to socket
			_, err = unix.Splice(pipeFds[0], nil, clientFd, nil, int(n), 0)
			if err != nil {
				b.Fatal(err)
			}

			remaining -= int(n)
		}
	}
}

// BenchmarkTraditionalPipeToSocket benchmarks traditional pipe read/write
func BenchmarkTraditionalPipeToSocket(b *testing.B) {
	// Create pipe
	pipeFds := make([]int, 2)
	if err := unix.Pipe(pipeFds); err != nil {
		b.Fatal(err)
	}
	defer unix.Close(pipeFds[0])
	defer unix.Close(pipeFds[1])

	listener, serverConn, clientConn := setupTCPConnection(b)
	defer listener.Close()
	defer serverConn.Close()
	defer clientConn.Close()

	go func() {
		io.Copy(io.Discard, serverConn)
	}()

	data := make([]byte, 4096)
	buf := make([]byte, 4096)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Write to pipe
		unix.Write(pipeFds[1], data)

		// Read from pipe and write to socket
		n, _ := unix.Read(pipeFds[0], buf)
		clientConn.Write(buf[:n])
	}
}

// BenchmarkVmspliceUserToKernel benchmarks vmsplice for zero-copy user buffer to pipe
func BenchmarkVmspliceUserToKernel(b *testing.B) {
	// Create pipe
	pipeFds := make([]int, 2)
	if err := unix.Pipe(pipeFds); err != nil {
		b.Fatal(err)
	}
	defer unix.Close(pipeFds[0])
	defer unix.Close(pipeFds[1])

	// Goroutine to drain pipe
	go func() {
		buf := make([]byte, 4096)
		for {
			_, err := unix.Read(pipeFds[0], buf)
			if err != nil {
				return
			}
		}
	}()

	data := make([]byte, 4096)
	for i := range data {
		data[i] = byte(i % 256)
	}

	// Create iovec for vmsplice
	iovec := []unix.Iovec{{
		Base: &data[0],
		Len:  uint64(len(data)),
	}}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := unix.Vmsplice(pipeFds[1], iovec, 0)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkVmsplicePipeTransfer benchmarks full vmsplice + splice chain
func BenchmarkVmsplicePipeTransfer(b *testing.B) {
	// Create pipe
	pipeFds := make([]int, 2)
	if err := unix.Pipe(pipeFds); err != nil {
		b.Fatal(err)
	}
	defer unix.Close(pipeFds[0])
	defer unix.Close(pipeFds[1])

	// Create second pipe for transfer
	pipe2Fds := make([]int, 2)
	if err := unix.Pipe(pipe2Fds); err != nil {
		b.Fatal(err)
	}
	defer unix.Close(pipe2Fds[0])
	defer unix.Close(pipe2Fds[1])

	// Goroutine to drain second pipe
	go func() {
		buf := make([]byte, 4096)
		for {
			_, err := unix.Read(pipe2Fds[0], buf)
			if err != nil {
				return
			}
		}
	}()

	data := make([]byte, 4096)
	for i := range data {
		data[i] = byte(i % 256)
	}

	iovec := []unix.Iovec{{
		Base: &data[0],
		Len:  uint64(len(data)),
	}}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Vmsplice user buffer to pipe
		n, err := unix.Vmsplice(pipeFds[1], iovec, 0)
		if err != nil {
			b.Fatal(err)
		}

		// Splice from first pipe to second pipe
		_, err = unix.Splice(pipeFds[0], nil, pipe2Fds[1], nil, n, 0)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkTraditionalWriteToPipe benchmarks traditional write to pipe
func BenchmarkTraditionalWriteToPipe(b *testing.B) {
	// Create pipe
	pipeFds := make([]int, 2)
	if err := unix.Pipe(pipeFds); err != nil {
		b.Fatal(err)
	}
	defer unix.Close(pipeFds[0])
	defer unix.Close(pipeFds[1])

	// Goroutine to drain pipe
	go func() {
		buf := make([]byte, 4096)
		for {
			_, err := unix.Read(pipeFds[0], buf)
			if err != nil {
				return
			}
		}
	}()

	data := make([]byte, 4096)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := unix.Write(pipeFds[1], data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkSendfileVsTraditional compares sendfile and traditional copy side-by-side
func BenchmarkSendfileVsTraditional(b *testing.B) {
	b.Run("Sendfile", func(b *testing.B) {
		tmpFile, tmpPath := createTempFileWithData(b, 64*1024)
		defer os.Remove(tmpPath)
		defer tmpFile.Close()

		listener, serverConn, clientConn := setupTCPConnection(b)
		defer listener.Close()
		defer serverConn.Close()
		defer clientConn.Close()

		go func() {
			io.Copy(io.Discard, serverConn)
		}()

		clientFile, _ := clientConn.(*net.TCPConn).File()
		defer clientFile.Close()

		clientFd := int(clientFile.Fd())
		fileFd := int(tmpFile.Fd())

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			tmpFile.Seek(0, 0)
			unix.Sendfile(clientFd, fileFd, nil, 64*1024)
		}
	})

	b.Run("Traditional", func(b *testing.B) {
		tmpFile, tmpPath := createTempFileWithData(b, 64*1024)
		defer os.Remove(tmpPath)
		defer tmpFile.Close()

		listener, serverConn, clientConn := setupTCPConnection(b)
		defer listener.Close()
		defer serverConn.Close()
		defer clientConn.Close()

		go func() {
			io.Copy(io.Discard, serverConn)
		}()

		buf := make([]byte, 32*1024)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			tmpFile.Seek(0, 0)
			io.CopyBuffer(clientConn, tmpFile, buf)
		}
	})
}

// BenchmarkSpliceVsTraditional compares splice and traditional pipe operations
func BenchmarkSpliceVsTraditional(b *testing.B) {
	b.Run("Splice", func(b *testing.B) {
		pipeFds := make([]int, 2)
		unix.Pipe(pipeFds)
		defer unix.Close(pipeFds[0])
		defer unix.Close(pipeFds[1])

		listener, serverConn, clientConn := setupTCPConnection(b)
		defer listener.Close()
		defer serverConn.Close()
		defer clientConn.Close()

		go func() {
			io.Copy(io.Discard, serverConn)
		}()

		clientFile, _ := clientConn.(*net.TCPConn).File()
		defer clientFile.Close()

		clientFd := int(clientFile.Fd())
		data := make([]byte, 4096)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			unix.Write(pipeFds[1], data)
			unix.Splice(pipeFds[0], nil, clientFd, nil, len(data), 0)
		}
	})

	b.Run("Traditional", func(b *testing.B) {
		pipeFds := make([]int, 2)
		unix.Pipe(pipeFds)
		defer unix.Close(pipeFds[0])
		defer unix.Close(pipeFds[1])

		listener, serverConn, clientConn := setupTCPConnection(b)
		defer listener.Close()
		defer serverConn.Close()
		defer clientConn.Close()

		go func() {
			io.Copy(io.Discard, serverConn)
		}()

		data := make([]byte, 4096)
		buf := make([]byte, 4096)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			unix.Write(pipeFds[1], data)
			n, _ := unix.Read(pipeFds[0], buf)
			clientConn.Write(buf[:n])
		}
	})
}

// BenchmarkVmspliceVsTraditional compares vmsplice and traditional write
func BenchmarkVmspliceVsTraditional(b *testing.B) {
	b.Run("Vmsplice", func(b *testing.B) {
		pipeFds := make([]int, 2)
		unix.Pipe(pipeFds)
		defer unix.Close(pipeFds[0])
		defer unix.Close(pipeFds[1])

		go func() {
			buf := make([]byte, 4096)
			for {
				_, err := unix.Read(pipeFds[0], buf)
				if err != nil {
					return
				}
			}
		}()

		data := make([]byte, 4096)
		iovec := []unix.Iovec{{
			Base: &data[0],
			Len:  uint64(len(data)),
		}}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			unix.Vmsplice(pipeFds[1], iovec, 0)
		}
	})

	b.Run("Traditional", func(b *testing.B) {
		pipeFds := make([]int, 2)
		unix.Pipe(pipeFds)
		defer unix.Close(pipeFds[0])
		defer unix.Close(pipeFds[1])

		go func() {
			buf := make([]byte, 4096)
			for {
				_, err := unix.Read(pipeFds[0], buf)
				if err != nil {
					return
				}
			}
		}()

		data := make([]byte, 4096)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			unix.Write(pipeFds[1], data)
		}
	})
}
