package mybenchmark

import (
	"io"
	"net"
	"sync"
	"testing"
)

// BenchmarkNoMutexTCPEcho — baseline TCP echo without mutex protection.
// Measures raw loopback socket latency.
func BenchmarkNoMutexTCPEcho(b *testing.B) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		b.Fatal(err)
	}
	defer listener.Close()

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				io.Copy(c, c)
			}(conn)
		}
	}()

	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		b.Fatal(err)
	}
	defer conn.Close()

	var data [5]byte
	var buf [5]byte
	copy(data[:], []byte("hello"))

	b.ResetTimer()
	for b.Loop() {
		_, err := conn.Write(data[:])
		if err != nil {
			b.Fatal(err)
		}
		_, err = io.ReadFull(conn, buf[:])
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkWithMutexTCPEcho — TCP echo with uncontended mutex protection.
// Isolates Lock/Unlock overhead on each operation.
func BenchmarkWithMutexTCPEcho(b *testing.B) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		b.Fatal(err)
	}
	defer listener.Close()

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				io.Copy(c, c)
			}(conn)
		}
	}()

	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		b.Fatal(err)
	}
	defer conn.Close()

	var mu sync.Mutex
	var data [5]byte
	var buf [5]byte
	copy(data[:], []byte("hello"))

	b.ResetTimer()
	for b.Loop() {
		mu.Lock()
		_, err := conn.Write(data[:])
		if err != nil {
			b.Fatal(err)
		}
		_, err = io.ReadFull(conn, buf[:])
		if err != nil {
			b.Fatal(err)
		}
		mu.Unlock()
	}
}

// BenchmarkNoMutexUDPEcho — baseline UDP echo without mutex protection.
// Measures raw loopback datagram latency.
func BenchmarkNoMutexUDPEcho(b *testing.B) {
	serverAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		b.Fatal(err)
	}

	serverConn, err := net.ListenUDP("udp", serverAddr)
	if err != nil {
		b.Fatal(err)
	}
	defer serverConn.Close()

	go func() {
		var buf [1024]byte
		for {
			n, addr, err := serverConn.ReadFromUDP(buf[:])
			if err != nil {
				return
			}
			serverConn.WriteToUDP(buf[:n], addr)
		}
	}()

	clientConn, err := net.DialUDP("udp", nil, serverConn.LocalAddr().(*net.UDPAddr))
	if err != nil {
		b.Fatal(err)
	}
	defer clientConn.Close()

	var data [5]byte
	var buf [5]byte
	copy(data[:], []byte("hello"))

	b.ResetTimer()
	for b.Loop() {
		_, err := clientConn.Write(data[:])
		if err != nil {
			b.Fatal(err)
		}
		_, err = clientConn.Read(buf[:])
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkWithMutexUDPEcho — UDP echo with uncontended mutex protection.
// Isolates Lock/Unlock overhead on each operation.
func BenchmarkWithMutexUDPEcho(b *testing.B) {
	serverAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		b.Fatal(err)
	}

	serverConn, err := net.ListenUDP("udp", serverAddr)
	if err != nil {
		b.Fatal(err)
	}
	defer serverConn.Close()

	go func() {
		var buf [1024]byte
		for {
			n, addr, err := serverConn.ReadFromUDP(buf[:])
			if err != nil {
				return
			}
			serverConn.WriteToUDP(buf[:n], addr)
		}
	}()

	clientConn, err := net.DialUDP("udp", nil, serverConn.LocalAddr().(*net.UDPAddr))
	if err != nil {
		b.Fatal(err)
	}
	defer clientConn.Close()

	var mu sync.Mutex
	var data [5]byte
	var buf [5]byte
	copy(data[:], []byte("hello"))

	b.ResetTimer()
	for b.Loop() {
		mu.Lock()
		_, err := clientConn.Write(data[:])
		if err != nil {
			b.Fatal(err)
		}
		_, err = clientConn.Read(buf[:])
		if err != nil {
			b.Fatal(err)
		}
		mu.Unlock()
	}
}

// Phase 4: Contended mutex benchmarks
// Measure performance when multiple goroutines compete for the same lock.

// BenchmarkMutexContentionTCP_2Goroutines — two goroutines contending for TCP connection.
func BenchmarkMutexContentionTCP_2Goroutines(b *testing.B) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		b.Fatal(err)
	}
	defer listener.Close()

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				io.Copy(c, c)
			}(conn)
		}
	}()

	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		b.Fatal(err)
	}
	defer conn.Close()

	var mu sync.Mutex
	opsPerGoroutine := b.N / 2

	b.ResetTimer()
	var wg sync.WaitGroup
	for g := 0; g < 2; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var data [5]byte
			var buf [5]byte
			copy(data[:], []byte("hello"))
			for i := 0; i < opsPerGoroutine; i++ {
				mu.Lock()
				_, _ = conn.Write(data[:])
				_, _ = io.ReadFull(conn, buf[:])
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	b.ReportMetric(float64(b.N)/b.Elapsed().Seconds(), "ops/sec")
}

// BenchmarkMutexContentionTCP_4Goroutines — four goroutines contending for TCP connection.
func BenchmarkMutexContentionTCP_4Goroutines(b *testing.B) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		b.Fatal(err)
	}
	defer listener.Close()

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				io.Copy(c, c)
			}(conn)
		}
	}()

	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		b.Fatal(err)
	}
	defer conn.Close()

	var mu sync.Mutex
	opsPerGoroutine := b.N / 4

	b.ResetTimer()
	var wg sync.WaitGroup
	for g := 0; g < 4; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var data [5]byte
			var buf [5]byte
			copy(data[:], []byte("hello"))
			for i := 0; i < opsPerGoroutine; i++ {
				mu.Lock()
				_, _ = conn.Write(data[:])
				_, _ = io.ReadFull(conn, buf[:])
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	b.ReportMetric(float64(b.N)/b.Elapsed().Seconds(), "ops/sec")
}

// BenchmarkMutexContentionUDP_2Goroutines — two goroutines contending for UDP connection.
func BenchmarkMutexContentionUDP_2Goroutines(b *testing.B) {
	serverAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		b.Fatal(err)
	}

	serverConn, err := net.ListenUDP("udp", serverAddr)
	if err != nil {
		b.Fatal(err)
	}
	defer serverConn.Close()

	go func() {
		var buf [1024]byte
		for {
			n, addr, err := serverConn.ReadFromUDP(buf[:])
			if err != nil {
				return
			}
			serverConn.WriteToUDP(buf[:n], addr)
		}
	}()

	clientConn, err := net.DialUDP("udp", nil, serverConn.LocalAddr().(*net.UDPAddr))
	if err != nil {
		b.Fatal(err)
	}
	defer clientConn.Close()

	var mu sync.Mutex
	opsPerGoroutine := b.N / 2

	b.ResetTimer()
	var wg sync.WaitGroup
	for g := 0; g < 2; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var data [5]byte
			var buf [5]byte
			copy(data[:], []byte("hello"))
			for i := 0; i < opsPerGoroutine; i++ {
				mu.Lock()
				_, _ = clientConn.Write(data[:])
				_, _ = clientConn.Read(buf[:])
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	b.ReportMetric(float64(b.N)/b.Elapsed().Seconds(), "ops/sec")
}

// BenchmarkMutexContentionUDP_4Goroutines — four goroutines contending for UDP connection.
func BenchmarkMutexContentionUDP_4Goroutines(b *testing.B) {
	serverAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		b.Fatal(err)
	}

	serverConn, err := net.ListenUDP("udp", serverAddr)
	if err != nil {
		b.Fatal(err)
	}
	defer serverConn.Close()

	go func() {
		var buf [1024]byte
		for {
			n, addr, err := serverConn.ReadFromUDP(buf[:])
			if err != nil {
				return
			}
			serverConn.WriteToUDP(buf[:n], addr)
		}
	}()

	clientConn, err := net.DialUDP("udp", nil, serverConn.LocalAddr().(*net.UDPAddr))
	if err != nil {
		b.Fatal(err)
	}
	defer clientConn.Close()

	var mu sync.Mutex
	opsPerGoroutine := b.N / 4

	b.ResetTimer()
	var wg sync.WaitGroup
	for g := 0; g < 4; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var data [5]byte
			var buf [5]byte
			copy(data[:], []byte("hello"))
			for i := 0; i < opsPerGoroutine; i++ {
				mu.Lock()
				_, _ = clientConn.Write(data[:])
				_, _ = clientConn.Read(buf[:])
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	b.ReportMetric(float64(b.N)/b.Elapsed().Seconds(), "ops/sec")
}

// Phase 5: Parallel throughput with uncontended mutex
// Measure throughput when each goroutine owns its own mutex and connection.
// This is the most common real-world pattern for concurrent socket handlers.

// BenchmarkMutexPerGoroutineTCP_Parallel — measure TCP throughput with per-goroutine mutex.
// Each goroutine spawned by b.RunParallel gets its own connection and uncontended mutex.
func BenchmarkMutexPerGoroutineTCP_Parallel(b *testing.B) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		b.Fatal(err)
	}
	defer listener.Close()

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				io.Copy(c, c)
			}(conn)
		}
	}()

	listenerAddr := listener.Addr().String()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		conn, err := net.Dial("tcp", listenerAddr)
		if err != nil {
			b.Fatal(err)
		}
		defer conn.Close()

		var mu sync.Mutex
		var data [5]byte
		var buf [5]byte
		copy(data[:], []byte("hello"))

		for pb.Next() {
			mu.Lock()
			_, _ = conn.Write(data[:])
			_, _ = io.ReadFull(conn, buf[:])
			mu.Unlock()
		}
	})
}

// BenchmarkMutexPerGoroutineUDP_Parallel — measure UDP throughput with per-goroutine mutex.
// Each goroutine spawned by b.RunParallel gets its own connection and uncontended mutex.
func BenchmarkMutexPerGoroutineUDP_Parallel(b *testing.B) {
	serverAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		b.Fatal(err)
	}

	serverConn, err := net.ListenUDP("udp", serverAddr)
	if err != nil {
		b.Fatal(err)
	}
	defer serverConn.Close()

	go func() {
		var buf [1024]byte
		for {
			n, addr, err := serverConn.ReadFromUDP(buf[:])
			if err != nil {
				return
			}
			serverConn.WriteToUDP(buf[:n], addr)
		}
	}()

	serverUDPAddr := serverConn.LocalAddr().(*net.UDPAddr)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		clientConn, err := net.DialUDP("udp", nil, serverUDPAddr)
		if err != nil {
			b.Fatal(err)
		}
		defer clientConn.Close()

		var mu sync.Mutex
		var data [5]byte
		var buf [5]byte
		copy(data[:], []byte("hello"))

		for pb.Next() {
			mu.Lock()
			_, _ = clientConn.Write(data[:])
			_, _ = clientConn.Read(buf[:])
			mu.Unlock()
		}
	})
}
