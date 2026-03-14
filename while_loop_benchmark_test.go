package mybenchmark

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// BenchmarkEmptyWhileLoop benchmarks the overhead of an empty while loop
// This shows the pure cost of the loop condition check
func BenchmarkEmptyWhileLoop(b *testing.B) {
	i := 0
	b.ResetTimer()
	for i < b.N {
		i++
	}
}

// BenchmarkWhileLoopWithSimpleCondition benchmarks while loop with simple condition
func BenchmarkWhileLoopWithSimpleCondition(b *testing.B) {
	i := 0
	running := true
	count := 0
	b.ResetTimer()
	for i < b.N && running {
		count++
		if count >= b.N {
			running = false
		}
		i++
	}
}

// BenchmarkWhileLoopWithAtomicRead benchmarks while loop checking atomic variable
// Simulates checking a shared state flag (common in real-time apps)
func BenchmarkWhileLoopWithAtomicRead(b *testing.B) {
	var running atomic.Bool
	running.Store(true)

	i := 0
	b.ResetTimer()
	for i < b.N && running.Load() {
		i++
	}
}

// BenchmarkWhileLoopWithChannelCheck benchmarks polling a channel in a loop
// Simulates message queue polling pattern
func BenchmarkWhileLoopWithChannelCheck(b *testing.B) {
	ch := make(chan int, 100)
	done := make(chan bool)

	// Producer
	go func() {
		for i := 0; i < b.N; i++ {
			ch <- i
		}
		close(ch)
	}()

	count := 0
	b.ResetTimer()

	// Consumer with while-like pattern
	for {
		select {
		case _, ok := <-ch:
			if !ok {
				done <- true
				return
			}
			count++
		}
	}
}

// BenchmarkWhileLoopBusyWait benchmarks busy-waiting pattern
// Shows the CPU cost of spinning without yielding
func BenchmarkWhileLoopBusyWait(b *testing.B) {
	var ready atomic.Bool
	ready.Store(false)

	// Simulate work being done in background
	go func() {
		time.Sleep(1 * time.Microsecond)
		ready.Store(true)
	}()

	b.ResetTimer()
	iterations := 0
	for i := 0; i < b.N; i++ {
		ready.Store(false)

		// Start background task
		go func() {
			time.Sleep(1 * time.Microsecond)
			ready.Store(true)
		}()

		// Busy wait
		for !ready.Load() {
			iterations++
		}
	}
	b.ReportMetric(float64(iterations)/float64(b.N), "spins/op")
}

// BenchmarkWhileLoopWithSleep benchmarks loop with sleep (yielding CPU)
// Compares to busy wait - shows trade-off between CPU usage and latency
func BenchmarkWhileLoopWithSleep(b *testing.B) {
	var ready atomic.Bool
	ready.Store(false)

	sleepDuration := 1 * time.Microsecond

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ready.Store(false)

		// Start background task
		go func() {
			time.Sleep(10 * time.Microsecond)
			ready.Store(true)
		}()

		// Wait with sleep
		for !ready.Load() {
			time.Sleep(sleepDuration)
		}
	}
}

// BenchmarkGameLoopSimulation benchmarks a typical game loop pattern
// Fixed time step with state updates
func BenchmarkGameLoopSimulation(b *testing.B) {
	type GameObject struct {
		x, y   float64
		vx, vy float64
	}

	objects := make([]GameObject, 100)
	for i := range objects {
		objects[i] = GameObject{
			x: float64(i), y: float64(i),
			vx: 1.0, vy: 1.0,
		}
	}

	deltaTime := 0.016 // 60 FPS

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Update all objects
		for j := range objects {
			objects[j].x += objects[j].vx * deltaTime
			objects[j].y += objects[j].vy * deltaTime
		}
	}
}

// BenchmarkConnectionAcceptLoop benchmarks server accept loop pattern
// Simulates waiting for incoming connections
func BenchmarkConnectionAcceptLoop(b *testing.B) {
	connections := make(chan int, b.N)

	// Simulate incoming connections
	go func() {
		for i := 0; i < b.N; i++ {
			connections <- i
		}
		close(connections)
	}()

	accepted := 0
	b.ResetTimer()

	// Accept loop
	for conn := range connections {
		accepted = conn
		_ = accepted
	}
}

// BenchmarkEventLoopSimulation benchmarks event loop pattern
// Simulates processing events from multiple sources
func BenchmarkEventLoopSimulation(b *testing.B) {
	events := make(chan string, 1000)
	done := make(chan bool)

	eventTypes := []string{"click", "keypress", "mouse_move", "timer"}

	// Event producer
	go func() {
		for i := 0; i < b.N; i++ {
			events <- eventTypes[i%len(eventTypes)]
		}
		close(events)
	}()

	eventCounts := make(map[string]int)

	b.ResetTimer()

	// Event loop
	go func() {
		for event := range events {
			eventCounts[event]++
		}
		done <- true
	}()

	<-done
}

// BenchmarkMessageQueueProcessing benchmarks message queue consumer pattern
// Shows cost of continuously polling and processing messages
func BenchmarkMessageQueueProcessing(b *testing.B) {
	type Message struct {
		ID      int
		Payload []byte
	}

	queue := make(chan Message, 1000)

	// Producer
	go func() {
		for i := 0; i < b.N; i++ {
			queue <- Message{
				ID:      i,
				Payload: []byte("test message"),
			}
		}
		close(queue)
	}()

	processed := 0
	b.ResetTimer()

	// Consumer loop
	for msg := range queue {
		// Simulate processing
		_ = msg.ID
		_ = len(msg.Payload)
		processed++
	}

	b.ReportMetric(float64(processed), "messages")
}

// BenchmarkWorkStealingLoop benchmarks work-stealing pattern
// Multiple workers competing for tasks
func BenchmarkWorkStealingLoop(b *testing.B) {
	tasks := make(chan int, b.N)
	numWorkers := 4
	var wg sync.WaitGroup

	// Fill task queue
	go func() {
		for i := 0; i < b.N; i++ {
			tasks <- i
		}
		close(tasks)
	}()

	b.ResetTimer()

	// Workers in while-loop pattern
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			localCount := 0
			for task := range tasks {
				// Simulate work
				_ = task * 2
				localCount++
			}
		}()
	}

	wg.Wait()
}

// BenchmarkRateLimitedLoop benchmarks loop with rate limiting
// Common in API servers, packet processing, etc.
func BenchmarkRateLimitedLoop(b *testing.B) {
	// Process 1000 items per second
	ticker := time.NewTicker(time.Second / 1000)
	defer ticker.Stop()

	b.ResetTimer()

	processed := 0
	for processed < b.N {
		<-ticker.C
		// Simulate processing one item
		processed++
	}
}

// BenchmarkSpinLockAcquisition benchmarks spinlock pattern
// Shows cost of lock contention with spinning
func BenchmarkSpinLockAcquisition(b *testing.B) {
	var locked atomic.Bool

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// Try to acquire lock with spinning
			for !locked.CompareAndSwap(false, true) {
				// Spin
			}

			// Critical section (minimal work)
			_ = 1 + 1

			// Release lock
			locked.Store(false)
		}
	})
}

// BenchmarkPollingVsBlocking compares polling vs blocking approaches
func BenchmarkPollingVsBlocking(b *testing.B) {
	b.Run("Blocking", func(b *testing.B) {
		ch := make(chan int, 100)

		go func() {
			for i := 0; i < b.N; i++ {
				ch <- i
			}
			close(ch)
		}()

		b.ResetTimer()
		for range ch {
			// Process message
		}
	})

	b.Run("SelectWithDefault", func(b *testing.B) {
		ch := make(chan int, 100)

		go func() {
			for i := 0; i < b.N; i++ {
				ch <- i
			}
			close(ch)
		}()

		b.ResetTimer()
		count := 0
		for count < b.N {
			select {
			case _, ok := <-ch:
				if ok {
					count++
				} else {
					return
				}
			default:
				// Polling - check again
			}
		}
	})
} // BenchmarkLoopLatency measures per-iteration latency in different scenarios
func BenchmarkLoopLatency(b *testing.B) {
	b.Run("NoContention", func(b *testing.B) {
		counter := 0
		b.ResetTimer()
		start := time.Now()

		for i := 0; i < b.N; i++ {
			counter++
		}

		elapsed := time.Since(start)
		b.ReportMetric(float64(elapsed.Nanoseconds())/float64(b.N), "ns/iteration")
	})

	b.Run("WithAtomicIncrement", func(b *testing.B) {
		var counter atomic.Int64
		b.ResetTimer()
		start := time.Now()

		for i := 0; i < b.N; i++ {
			counter.Add(1)
		}

		elapsed := time.Since(start)
		b.ReportMetric(float64(elapsed.Nanoseconds())/float64(b.N), "ns/iteration")
	})

	b.Run("WithMutex", func(b *testing.B) {
		var mu sync.Mutex
		counter := 0
		b.ResetTimer()
		start := time.Now()

		for i := 0; i < b.N; i++ {
			mu.Lock()
			counter++
			mu.Unlock()
		}

		elapsed := time.Since(start)
		b.ReportMetric(float64(elapsed.Nanoseconds())/float64(b.N), "ns/iteration")
	})
}

// BenchmarkCPUUsageComparison - Run with CPU profiling to see the difference!
// go test -bench=BenchmarkCPUUsageComparison -cpuprofile=cpu.prof -benchtime=3s
func BenchmarkCPUUsageComparison(b *testing.B) {
	b.Run("BusyWaitPolling_BurnsCPU", func(b *testing.B) {
		// This will show HIGH CPU usage in profile
		var ready atomic.Bool
		messages := make(chan int, 1000)
		done := make(chan bool)

		// Slow producer - sends occasionally
		go func() {
			for i := 0; i < 100; i++ {
				time.Sleep(10 * time.Millisecond) // Simulate slow arrival
				messages <- i
				ready.Store(true)
			}
			close(messages)
		}()

		b.ResetTimer()
		processed := 0

		// BUSY WAIT - spins continuously checking the flag
		go func() {
			for {
				if ready.Load() {
					select {
					case msg, ok := <-messages:
						if !ok {
							done <- true
							return
						}
						_ = msg
						processed++
						ready.Store(false)
					default:
					}
				}
				// This loop NEVER STOPS - burns CPU even when idle!
			}
		}()

		<-done
	})

	b.Run("ChannelBlocking_IdleCPU", func(b *testing.B) {
		// This will show LOW CPU usage in profile
		messages := make(chan int, 1000)
		done := make(chan bool)

		// Same slow producer
		go func() {
			for i := 0; i < 100; i++ {
				time.Sleep(10 * time.Millisecond) // Simulate slow arrival
				messages <- i
			}
			close(messages)
		}()

		b.ResetTimer()
		processed := 0

		// BLOCKING - goroutine sleeps when channel is empty
		go func() {
			for msg := range messages {
				_ = msg
				processed++
				// Goroutine automatically sleeps here when no messages!
			}
			done <- true
		}()

		<-done
	})
}
