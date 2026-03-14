// Simple CPU Demo - shows busy-wait vs blocking
// Run: go run simple_cpu_demo.go
// Monitor in another terminal: top -pid $(pgrep simple_cpu_demo)

package main

import (
	"fmt"
	"sync/atomic"
	"time"
)

func main() {
	fmt.Println("=== CPU Usage Demo ===")
	fmt.Println("Run 'top' or 'htop' in another terminal to see CPU usage!")
	fmt.Println()

	// Demo 1: Busy-Wait (Burns CPU)
	fmt.Println("🔥 Starting BUSY-WAIT loop for 10 seconds...")
	fmt.Println("   Check CPU: One core should be at ~100%")
	demoBusyWait()

	fmt.Println()
	fmt.Println("Cooling down for 3 seconds...")
	time.Sleep(3 * time.Second)

	// Demo 2: Channel Blocking (Efficient)
	fmt.Println()
	fmt.Println("✅ Starting CHANNEL BLOCKING for 10 seconds...")
	fmt.Println("   Check CPU: Should be near 0%")
	demoChannelBlocking()

	fmt.Println()
	fmt.Println("Demo complete!")
}

func demoBusyWait() {
	var ready atomic.Bool
	done := make(chan bool)
	
	// Goroutine that sets ready flag occasionally
	go func() {
		for i := 0; i < 10; i++ {
			time.Sleep(1 * time.Second)
			ready.Store(true)
			time.Sleep(10 * time.Millisecond)
			ready.Store(false)
		}
	}()
	
	// BUSY-WAIT LOOP - This burns 100% of one CPU core!
	go func() {
		spins := uint64(0)
		start := time.Now()
		
		for time.Since(start) < 10*time.Second {
			if ready.Load() {
				// Found it!
			}
			spins++
			// NO SLEEP HERE - keeps spinning at full speed!
		}
		
		fmt.Printf("   Busy-wait completed: %d spins in 10 seconds\n", spins)
		done <- true
	}()
	
	<-done
}

func demoChannelBlocking() {
	messages := make(chan int)
	done := make(chan bool)
	
	// Sender: Sends messages occasionally
	go func() {
		for i := 0; i < 10; i++ {
			time.Sleep(1 * time.Second)
			messages <- i
		}
		close(messages)
	}()
	
	// BLOCKING RECEIVER - CPU stays idle!
	go func() {
		count := 0
		for msg := range messages {
			// Goroutine SLEEPS here when channel is empty
			// Go scheduler parks it - zero CPU usage!
			count++
			_ = msg
		}
		
		fmt.Printf("   Blocking completed: %d messages received\n", count)
		done <- true
	}()
	
	<-done
}
