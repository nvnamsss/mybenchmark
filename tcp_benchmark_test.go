package mybenchmark

import (
	"io"
	"net"
	"testing"
)

// BenchmarkTCPEcho benchmarks a TCP echo operation (client sends, server echoes back)
func BenchmarkTCPEcho(b *testing.B) {
	// Start a TCP server
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		b.Fatal(err)
	}
	defer listener.Close()

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
	conn, err := net.Dial("tcp", listener.Addr().String())
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

// BenchmarkTCPConnect benchmarks just the TCP connection establishment
func BenchmarkTCPConnect(b *testing.B) {
	// Start a TCP server
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		b.Fatal(err)
	}
	defer listener.Close()

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
		conn, err := net.Dial("tcp", listener.Addr().String())
		if err != nil {
			b.Fatal(err)
		}
		conn.Close()
	}
}

// BenchmarkTCPWrite benchmarks TCP write operations with persistent connection
func BenchmarkTCPWrite(b *testing.B) {
	// Start a TCP server
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		b.Fatal(err)
	}
	defer listener.Close()

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

	// Write == Send
	//
	// Connect client
	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		b.Fatal(err)
	}
	defer conn.Close()

	data := []byte("hello")
	
	b.ResetTimer()
	// loop to benchmark
	for i := 0; i < b.N; i++ {
		_, err := conn.Write(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkTCPRead benchmarks TCP read operations with persistent connection
func BenchmarkTCPRead(b *testing.B) {
	// Start a TCP server
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		b.Fatal(err)
	}
	defer listener.Close()

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
	conn, err := net.Dial("tcp", listener.Addr().String())
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
