#!/bin/bash

echo "=========================================="
echo "CPU Usage Visualization - Quick Start"
echo "=========================================="
echo ""
echo "Choose a method:"
echo ""
echo "1. Simple Demo (Recommended - Easiest)"
echo "2. Benchmark with CPU Profiling"
echo "3. Custom monitor script"
echo ""
read -p "Enter choice (1-3): " choice

case $choice in
  1)
    echo ""
    echo "Starting simple demo..."
    echo "Open another terminal and run: htop"
    echo "or: watch -n 0.5 'ps aux | head -1; ps aux | grep simple_cpu_demo'"
    echo ""
    sleep 3
    cd cpu_demo && go run simple_cpu_demo.go
    ;;
  2)
    echo ""
    echo "Running benchmark with CPU profiling..."
    echo "This will take about 30 seconds..."
    go test -bench=BenchmarkPollingVsBlocking -cpuprofile=cpu.prof -benchtime=3s
    echo ""
    echo "Profile saved! Now analyze it:"
    echo "  go tool pprof -http=:8080 cpu.prof"
    echo "Then open http://localhost:8080 in your browser"
    ;;
  3)
    echo ""
    echo "Creating monitor script..."
    cat > /tmp/cpu_monitor.go << 'MONITOR'
package main

import (
"fmt"
"sync/atomic"
"time"
)

func main() {
fmt.Println("Monitor CPU in another terminal!")
fmt.Println("Run: watch -n 0.5 'ps aux | grep cpu_monitor'")
time.Sleep(3 * time.Second)

// 5 seconds of busy-wait
fmt.Println("\n🔥 BUSY-WAIT starting (100% CPU)...")
var flag atomic.Bool
done := make(chan bool)
go func() {
for i := 0; i < 5000000000; i++ {
flag.Load()
}
done <- true
}()
<-done

fmt.Println("\n✅ BLOCKING starting (0% CPU)...")
time.Sleep(5 * time.Second)

fmt.Println("\nDone!")
}
MONITOR
    go run /tmp/cpu_monitor.go
    ;;
esac
