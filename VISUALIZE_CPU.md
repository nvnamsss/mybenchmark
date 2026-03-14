# How to Visualize CPU Usage: Busy-Wait vs Channel Blocking

## Method 1: Run the Benchmark with CPU Profiling

### Step 1: Run benchmark and generate CPU profile
```bash
cd /home/namnv/git/mybenchmark
go test -bench=BenchmarkCPUUsageComparison -cpuprofile=cpu.prof -benchtime=3s
```

### Step 2: Analyze with pprof
```bash
# Interactive mode
go tool pprof cpu.prof

# In pprof, type these commands:
# top10          - See top CPU consumers
# list BusyWait  - See hot spots in busy wait code
# list Blocking  - See hot spots in blocking code
# web            - Open graphical view (requires graphviz)
```

### Step 3: Visual web interface
```bash
# Generate SVG graph
go tool pprof -svg cpu.prof > cpu_profile.svg

# Or start web server
go tool pprof -http=:8080 cpu.prof
# Open browser to http://localhost:8080
```

You'll see **BusyWaitPolling** consuming massive CPU time vs **ChannelBlocking** showing minimal CPU!

---

## Method 2: Live CPU Monitoring with Demo Program

### Build and run the demo
```bash
cd /home/namnv/git/mybenchmark/cpu_demo
go build -o cpu_demo
```

### In Terminal 1: Run the demo
```bash
./cpu_demo
```

### In Terminal 2: Monitor CPU in real-time
```bash
# Option A: Using top (press '1' to see per-core usage)
top -pid $(pgrep cpu_demo)

# Option B: Using htop (more visual)
htop -p $(pgrep cpu_demo)

# Option C: Watch ps output
watch -n 0.5 'ps aux | grep cpu_demo | grep -v grep'
```

### What you'll see:

**During BUSY-WAIT test:**
```
PID   USER  %CPU  %MEM  COMMAND
12345 user  100.0  0.1   ./cpu_demo    <-- One core at 100%!
```

**During CHANNEL BLOCKING test:**
```
PID   USER  %CPU  %MEM  COMMAND
12345 user   0.3   0.1   ./cpu_demo    <-- Near 0% CPU!
```

---

## Method 3: Simple Go Program with Runtime Stats

Create this file and run it:

```bash
cat > cpu_monitor.go << 'GOEOF'
package main

import (
    "fmt"
    "runtime"
    "sync/atomic"
    "time"
)

func main() {
    fmt.Println("Starting CPU comparison...")
    
    // Test 1: Busy wait for 5 seconds
    fmt.Println("\n🔥 BUSY WAIT - Check CPU usage now!")
    busyWait := make(chan bool)
    var flag atomic.Bool
    go func() {
        for i := 0; i < 5; i++ {
            time.Sleep(1 * time.Second)
            flag.Store(true)
        }
        busyWait <- true
    }()
    
    go func() {
        count := 0
        for {
            select {
            case <-busyWait:
                fmt.Printf("Busy-wait spins: %d\n", count)
                return
            default:
                if flag.Load() {
                    flag.Store(false)
                }
                count++
            }
        }
    }()
    
    time.Sleep(6 * time.Second)
    
    // Test 2: Channel blocking for 5 seconds
    fmt.Println("\n✅ CHANNEL BLOCKING - Check CPU usage now!")
    ch := make(chan int)
    go func() {
        for i := 0; i < 5; i++ {
            time.Sleep(1 * time.Second)
            ch <- i
        }
        close(ch)
    }()
    
    count := 0
    for range ch {
        count++
    }
    fmt.Printf("Blocking received: %d messages\n", count)
    
    fmt.Println("\nDone!")
}
GOEOF

go run cpu_monitor.go
```

While it runs, in another terminal run:
```bash
watch -n 0.5 'ps aux | grep "cpu_monitor\|%CPU" | grep -v grep'
```

---

## Method 4: Using runtime/trace

```bash
# Run test with trace
go test -bench=BenchmarkCPUUsageComparison -trace=trace.out -benchtime=3s

# View trace in browser
go tool trace trace.out
```

Click on "View trace" - you'll see:
- **Busy-wait**: Goroutine constantly running (green bar)
- **Blocking**: Goroutine mostly sleeping (gray periods)

---

## Expected Results

### Busy-Wait Polling:
- **CPU**: 100% of one core continuously
- **Why**: Tight loop never yields, keeps checking flag
- **Goroutine state**: Always "Running"

### Channel Blocking:
- **CPU**: 0-2% (only spikes when message arrives)
- **Why**: Goroutine sleeps, scheduler wakes it only when needed
- **Goroutine state**: "Waiting" most of the time

---

## Quick Commands Summary

```bash
# Build and run demo
cd /home/namnv/git/mybenchmark/cpu_demo && go build && ./cpu_demo

# In another terminal - monitor CPU
htop -p $(pgrep cpu_demo)

# Or benchmark with profiling
cd /home/namnv/git/mybenchmark
go test -bench=BenchmarkCPUUsageComparison -cpuprofile=cpu.prof -benchtime=5s
go tool pprof -http=:8080 cpu.prof
```

The difference will be **VIVID** - you'll see one approach burning CPU at 100% while the other sits at 0%! 🔥 vs ✅
