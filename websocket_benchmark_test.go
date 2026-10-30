package mybenchmark

import (
	"fmt"
	"net"
	"net/http"
	"testing"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// BenchmarkWebSocketHandshake benchmarks the WebSocket upgrade handshake
func BenchmarkWebSocketHandshake(b *testing.B) {
	// Create a new HTTP mux for this benchmark
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		conn.Close()
	})

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		b.Fatal(err)
	}
	defer listener.Close()

	go http.Serve(listener, mux)

	wsURL := fmt.Sprintf("ws://%s/ws", listener.Addr().String())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		if err != nil {
			b.Fatal(err)
		}
		conn.Close()
	}
}

// BenchmarkWebSocketEcho benchmarks WebSocket echo operation (client sends, server echoes back)
func BenchmarkWebSocketEcho(b *testing.B) {
	// Create a new HTTP mux for this benchmark
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		// Echo server
		for {
			messageType, data, err := conn.ReadMessage()
			if err != nil {
				return
			}
			err = conn.WriteMessage(messageType, data)
			if err != nil {
				return
			}
		}
	})

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		b.Fatal(err)
	}
	defer listener.Close()

	go http.Serve(listener, mux)

	wsURL := fmt.Sprintf("ws://%s/ws", listener.Addr().String())

	// Connect client
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		b.Fatal(err)
	}
	defer conn.Close()

	data := []byte("hello")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := conn.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			b.Fatal(err)
		}
		_, _, err = conn.ReadMessage()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkWebSocketWrite benchmarks WebSocket write operations with persistent connection
func BenchmarkWebSocketWrite(b *testing.B) {
	// Create a new HTTP mux for this benchmark
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		// Server just receives and discards
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				return
			}
		}
	})

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		b.Fatal(err)
	}
	defer listener.Close()

	go http.Serve(listener, mux)

	wsURL := fmt.Sprintf("ws://%s/ws", listener.Addr().String())

	// Connect client
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		b.Fatal(err)
	}
	defer conn.Close()

	data := []byte("hello")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := conn.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkWebSocketRead benchmarks WebSocket read operations
func BenchmarkWebSocketRead(b *testing.B) {
	// Create a new HTTP mux for this benchmark
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		// Server continuously sends messages
		data := []byte("hello")
		for {
			err := conn.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				return
			}
		}
	})

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		b.Fatal(err)
	}
	defer listener.Close()

	go http.Serve(listener, mux)

	wsURL := fmt.Sprintf("ws://%s/ws", listener.Addr().String())

	// Connect client
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		b.Fatal(err)
	}
	defer conn.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := conn.ReadMessage()
		if err != nil {
			b.Fatal(err)
		}
	}
}
