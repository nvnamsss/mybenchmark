package mybenchmark

import (
	"io"
	"net"
	"os"
	"testing"
)

// BenchmarkUnixSocketEcho benchmarks a Unix socket echo operation
func BenchmarkUnixSocketEcho(b *testing.B) {
	sockPath := "/tmp/bench_unix_echo.sock"
	os.Remove(sockPath) // Clean up any existing socket

	// Start a Unix socket server
	listener, err := net.Listen("unix", sockPath)
	if err != nil {
		b.Fatal(err)
	}
	defer listener.Close()
	defer os.Remove(sockPath)

	// Server goroutine
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				io.Copy(c, c) // Echo back
			}(conn)
		}
	}()

	// Connect client
	conn, err := net.Dial("unix", sockPath)
	if err != nil {
		b.Fatal(err)
	}
	defer conn.Close()

	data := []byte("hello")
	buf := make([]byte, len(data))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := conn.Write(data)
		if err != nil {
			b.Fatal(err)
		}
		_, err = io.ReadFull(conn, buf)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkUnixSocketConnect benchmarks Unix socket connection establishment
func BenchmarkUnixSocketConnect(b *testing.B) {
	sockPath := "/tmp/bench_unix_connect.sock"
	os.Remove(sockPath)

	// Start a Unix socket server
	listener, err := net.Listen("unix", sockPath)
	if err != nil {
		b.Fatal(err)
	}
	defer listener.Close()
	defer os.Remove(sockPath)

	// Server goroutine
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			conn.Close()
		}
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		conn, err := net.Dial("unix", sockPath)
		if err != nil {
			b.Fatal(err)
		}
		conn.Close()
	}
}

// BenchmarkUnixSocketWrite benchmarks Unix socket write operations
func BenchmarkUnixSocketWrite(b *testing.B) {
	sockPath := "/tmp/bench_unix_write.sock"
	os.Remove(sockPath)

	// Start a Unix socket server
	listener, err := net.Listen("unix", sockPath)
	if err != nil {
		b.Fatal(err)
	}
	defer listener.Close()
	defer os.Remove(sockPath)

	// Server goroutine - just discard data
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				io.Copy(io.Discard, c)
			}(conn)
		}
	}()

	// Connect client
	conn, err := net.Dial("unix", sockPath)
	if err != nil {
		b.Fatal(err)
	}
	defer conn.Close()

	data := []byte("hello")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := conn.Write(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkUnixSocketRead benchmarks Unix socket read operations
func BenchmarkUnixSocketRead(b *testing.B) {
	sockPath := "/tmp/bench_unix_read.sock"
	os.Remove(sockPath)

	// Start a Unix socket server
	listener, err := net.Listen("unix", sockPath)
	if err != nil {
		b.Fatal(err)
	}
	defer listener.Close()
	defer os.Remove(sockPath)

	data := []byte("hello")

	// Server goroutine - continuously send data
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				for {
					_, err := c.Write(data)
					if err != nil {
						return
					}
				}
			}(conn)
		}
	}()

	// Connect client
	conn, err := net.Dial("unix", sockPath)
	if err != nil {
		b.Fatal(err)
	}
	defer conn.Close()

	buf := make([]byte, len(data))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := io.ReadFull(conn, buf)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkUnixDatagramEcho benchmarks Unix datagram socket echo operation
func BenchmarkUnixDatagramEcho(b *testing.B) {
	serverPath := "/tmp/bench_unix_dgram_server.sock"
	clientPath := "/tmp/bench_unix_dgram_client.sock"
	os.Remove(serverPath)
	os.Remove(clientPath)

	// Start a Unix datagram server
	serverAddr, err := net.ResolveUnixAddr("unixgram", serverPath)
	if err != nil {
		b.Fatal(err)
	}

	serverConn, err := net.ListenUnixgram("unixgram", serverAddr)
	if err != nil {
		b.Fatal(err)
	}
	defer serverConn.Close()
	defer os.Remove(serverPath)

	// Server goroutine - echo back
	go func() {
		buf := make([]byte, 1024)
		for {
			n, addr, err := serverConn.ReadFromUnix(buf)
			if err != nil {
				return
			}
			serverConn.WriteToUnix(buf[:n], addr)
		}
	}()

	// Connect client
	clientAddr, err := net.ResolveUnixAddr("unixgram", clientPath)
	if err != nil {
		b.Fatal(err)
	}

	clientConn, err := net.ListenUnixgram("unixgram", clientAddr)
	if err != nil {
		b.Fatal(err)
	}
	defer clientConn.Close()
	defer os.Remove(clientPath)

	data := []byte("hello")
	buf := make([]byte, len(data))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := clientConn.WriteToUnix(data, serverAddr)
		if err != nil {
			b.Fatal(err)
		}
		_, _, err = clientConn.ReadFromUnix(buf)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkUnixDatagramWrite benchmarks Unix datagram socket write operations
func BenchmarkUnixDatagramWrite(b *testing.B) {
	serverPath := "/tmp/bench_unix_dgram_write_server.sock"
	clientPath := "/tmp/bench_unix_dgram_write_client.sock"
	os.Remove(serverPath)
	os.Remove(clientPath)

	// Start a Unix datagram server
	serverAddr, err := net.ResolveUnixAddr("unixgram", serverPath)
	if err != nil {
		b.Fatal(err)
	}

	serverConn, err := net.ListenUnixgram("unixgram", serverAddr)
	if err != nil {
		b.Fatal(err)
	}
	defer serverConn.Close()
	defer os.Remove(serverPath)

	// Server goroutine - just receive and discard
	go func() {
		buf := make([]byte, 1024)
		for {
			_, _, err := serverConn.ReadFromUnix(buf)
			if err != nil {
				return
			}
		}
	}()

	// Connect client
	clientAddr, err := net.ResolveUnixAddr("unixgram", clientPath)
	if err != nil {
		b.Fatal(err)
	}

	clientConn, err := net.ListenUnixgram("unixgram", clientAddr)
	if err != nil {
		b.Fatal(err)
	}
	defer clientConn.Close()
	defer os.Remove(clientPath)

	data := []byte("hello")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := clientConn.WriteToUnix(data, serverAddr)
		if err != nil {
			b.Fatal(err)
		}
	}
}
