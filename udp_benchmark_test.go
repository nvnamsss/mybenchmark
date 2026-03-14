package mybenchmark

import (
	"net"
	"testing"
)

// BenchmarkUDPEcho benchmarks a UDP echo operation (client sends, server echoes back)
func BenchmarkUDPEcho(b *testing.B) {
	// Start a UDP server
	serverAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		b.Fatal(err)
	}

	serverConn, err := net.ListenUDP("udp", serverAddr)
	if err != nil {
		b.Fatal(err)
	}
	defer serverConn.Close()

	// Server goroutine - echo back
	go func() {
		buf := make([]byte, 1024)
		for {
			n, addr, err := serverConn.ReadFromUDP(buf)
			if err != nil {
				return
			}
			serverConn.WriteToUDP(buf[:n], addr)
		}
	}()

	// Connect client
	clientConn, err := net.DialUDP("udp", nil, serverConn.LocalAddr().(*net.UDPAddr))
	if err != nil {
		b.Fatal(err)
	}
	defer clientConn.Close()

	data := []byte("hello")
	buf := make([]byte, len(data))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := clientConn.Write(data)
		if err != nil {
			b.Fatal(err)
		}
		_, err = clientConn.Read(buf)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkUDPWrite benchmarks UDP write (send) operations
func BenchmarkUDPWrite(b *testing.B) {
	// Start a UDP server
	serverAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		b.Fatal(err)
	}

	serverConn, err := net.ListenUDP("udp", serverAddr)
	if err != nil {
		b.Fatal(err)
	}
	defer serverConn.Close()

	// Server goroutine - just receive and discard
	go func() {
		buf := make([]byte, 1024)
		for {
			_, _, err := serverConn.ReadFromUDP(buf)
			if err != nil {
				return
			}
		}
	}()

	// Connect client
	clientConn, err := net.DialUDP("udp", nil, serverConn.LocalAddr().(*net.UDPAddr))
	if err != nil {
		b.Fatal(err)
	}
	defer clientConn.Close()

	data := []byte("hello")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := clientConn.Write(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkUDPRead benchmarks UDP read (receive) operations
func BenchmarkUDPRead(b *testing.B) {
	// Client receives
	clientAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		b.Fatal(err)
	}

	clientConn, err := net.ListenUDP("udp", clientAddr)
	if err != nil {
		b.Fatal(err)
	}
	defer clientConn.Close()

	data := []byte("hello")

	// Server goroutine - continuously send data to client
	go func() {
		sendConn, err := net.DialUDP("udp", nil, clientConn.LocalAddr().(*net.UDPAddr))
		if err != nil {
			return
		}
		defer sendConn.Close()

		for {
			_, err := sendConn.Write(data)
			if err != nil {
				return
			}
		}
	}()

	buf := make([]byte, len(data))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := clientConn.ReadFromUDP(buf)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkUDPSendTo benchmarks UDP sendto operations (without "connected" socket)
func BenchmarkUDPSendTo(b *testing.B) {
	// Start a UDP server
	serverAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		b.Fatal(err)
	}

	serverConn, err := net.ListenUDP("udp", serverAddr)
	if err != nil {
		b.Fatal(err)
	}
	defer serverConn.Close()

	// Server goroutine - just receive and discard
	go func() {
		buf := make([]byte, 1024)
		for {
			_, _, err := serverConn.ReadFromUDP(buf)
			if err != nil {
				return
			}
		}
	}()

	// Client socket (not connected)
	clientAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		b.Fatal(err)
	}

	clientConn, err := net.ListenUDP("udp", clientAddr)
	if err != nil {
		b.Fatal(err)
	}
	defer clientConn.Close()

	data := []byte("hello")
	targetAddr := serverConn.LocalAddr().(*net.UDPAddr)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := clientConn.WriteToUDP(data, targetAddr)
		if err != nil {
			b.Fatal(err)
		}
	}
}
