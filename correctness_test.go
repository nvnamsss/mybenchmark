package mybenchmark

import (
	"encoding/json"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/vmihailenco/msgpack/v5"
	"google.golang.org/protobuf/proto"

	"mybenchmark/proto/serialpb"
)

// Phase 7: Correctness Verification
// Verify that all synchronization approaches correctly count increments without lost updates.
// Each test runs N goroutines, each performing opsPerGoroutine increments.
// Expected final counter: N * opsPerGoroutine (no lost updates).

// TestCoarseGrainedCorrectness verifies that coarse-grained locking produces correct results.
func TestCoarseGrainedCorrectness(t *testing.T) {
	const numGoroutines = 8
	const opsPerGoroutine = 1000
	const expectedTotal = numGoroutines * opsPerGoroutine

	var mu sync.Mutex
	var counter int64

	var wg sync.WaitGroup
	for g := 0; g < numGoroutines; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < opsPerGoroutine; i++ {
				mu.Lock()
				counter++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	if counter != int64(expectedTotal) {
		t.Errorf("coarse-grained: expected %d, got %d (lost %d updates)",
			expectedTotal, counter, expectedTotal-int(counter))
	}
}

// TestFineGrainedCorrectness verifies fine-grained lock sharding produces correct results.
func TestFineGrainedCorrectness(t *testing.T) {
	const numGoroutines = 8
	const opsPerGoroutine = 1000
	const expectedTotal = numGoroutines * opsPerGoroutine

	locks := make([]sync.Mutex, numGoroutines)
	counters := make([]int64, numGoroutines)

	var wg sync.WaitGroup
	for g := 0; g < numGoroutines; g++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for i := 0; i < opsPerGoroutine; i++ {
				locks[id].Lock()
				counters[id]++
				locks[id].Unlock()
			}
		}(g)
	}
	wg.Wait()

	// Sum all shards
	var total int64
	for _, c := range counters {
		total += c
	}

	if total != int64(expectedTotal) {
		t.Errorf("fine-grained: expected %d, got %d (lost %d updates)",
			expectedTotal, total, expectedTotal-int(total))
	}
}

// TestOptimisticCorrectness verifies optimistic CAS-based synchronization produces correct results.
func TestOptimisticCorrectness(t *testing.T) {
	const numGoroutines = 8
	const opsPerGoroutine = 1000
	const expectedTotal = numGoroutines * opsPerGoroutine

	var counter int64

	var wg sync.WaitGroup
	for g := 0; g < numGoroutines; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < opsPerGoroutine; i++ {
				for {
					old := atomic.LoadInt64(&counter)
					if atomic.CompareAndSwapInt64(&counter, old, old+1) {
						break
					}
				}
			}
		}()
	}
	wg.Wait()

	if counter != int64(expectedTotal) {
		t.Errorf("optimistic: expected %d, got %d (lost %d updates)",
			expectedTotal, counter, expectedTotal-int(counter))
	}
}

// TestLockFreeCorrectness verifies lock-free atomic operations produce correct results.
func TestLockFreeCorrectness(t *testing.T) {
	const numGoroutines = 8
	const opsPerGoroutine = 1000
	const expectedTotal = numGoroutines * opsPerGoroutine

	var counter int64

	var wg sync.WaitGroup
	for g := 0; g < numGoroutines; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < opsPerGoroutine; i++ {
				atomic.AddInt64(&counter, 1)
			}
		}()
	}
	wg.Wait()

	if counter != int64(expectedTotal) {
		t.Errorf("lock-free: expected %d, got %d (lost %d updates)",
			expectedTotal, counter, expectedTotal-int(counter))
	}
}

// TestThreadLocalCorrectness verifies thread-local reduction with merge produces correct total.
func TestThreadLocalCorrectness(t *testing.T) {
	const numGoroutines = 8
	const opsPerGoroutine = 1000
	const expectedTotal = numGoroutines * opsPerGoroutine

	counters := make([]int64, numGoroutines)

	var wg sync.WaitGroup
	for g := 0; g < numGoroutines; g++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for i := 0; i < opsPerGoroutine; i++ {
				counters[id]++
			}
		}(g)
	}
	wg.Wait()

	// Merge results (single-threaded after sync)
	var total int64
	for _, c := range counters {
		total += c
	}

	if total != int64(expectedTotal) {
		t.Errorf("thread-local: expected %d, got %d (lost %d updates)",
			expectedTotal, total, expectedTotal-int(total))
	}
}

// TestRWMutexReadCorrectness verifies read-write mutex with pure reads produces consistent results.
func TestRWMutexReadCorrectness(t *testing.T) {
	const numReaders = 8
	const readsPerReader = 1000

	var rwmu sync.RWMutex
	var counter int64 = 42 // Constant value

	results := make([]int64, numReaders*readsPerReader)
	resultIdx := 0
	var resultMu sync.Mutex

	var wg sync.WaitGroup
	for g := 0; g < numReaders; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < readsPerReader; i++ {
				rwmu.RLock()
				val := counter
				rwmu.RUnlock()

				resultMu.Lock()
				results[resultIdx] = val
				resultIdx++
				resultMu.Unlock()
			}
		}()
	}
	wg.Wait()

	// All reads should have seen the constant value
	for i, val := range results[:numReaders*readsPerReader] {
		if val != 42 {
			t.Errorf("RWMutex read #%d: expected 42, got %d", i, val)
			return
		}
	}
}

// TestRWMutexMixedCorrectness verifies read-write mutex with mixed reads and writes.
func TestRWMutexMixedCorrectness(t *testing.T) {
	const numGoroutines = 8
	const opsPerGoroutine = 1000
	const numReaders = 6
	const numWriters = 2
	const expectedWrites = numWriters * opsPerGoroutine

	var rwmu sync.RWMutex
	var counter int64

	var wg sync.WaitGroup

	// Readers
	for g := 0; g < numReaders; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < opsPerGoroutine; i++ {
				rwmu.RLock()
				_ = counter // Just read the value
				rwmu.RUnlock()
			}
		}()
	}

	// Writers
	for g := 0; g < numWriters; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < opsPerGoroutine; i++ {
				rwmu.Lock()
				counter++
				rwmu.Unlock()
			}
		}()
	}

	wg.Wait()

	if counter != int64(expectedWrites) {
		t.Errorf("RWMutex mixed: expected %d writes, got %d (lost %d updates)",
			expectedWrites, counter, expectedWrites-int(counter))
	}
}

// TestAllApproachesConsistency runs all approaches in parallel and verifies they all
// produce the same correct result (optional extra validation).
func TestAllApproachesConsistency(t *testing.T) {
	const numGoroutines = 4
	const opsPerGoroutine = 100
	const expectedTotal = numGoroutines * opsPerGoroutine

	results := make(map[string]int64)
	var mu sync.Mutex

	// Coarse-grained
	func() {
		var coarseMu sync.Mutex
		var counter int64
		var wg sync.WaitGroup
		for g := 0; g < numGoroutines; g++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for i := 0; i < opsPerGoroutine; i++ {
					coarseMu.Lock()
					counter++
					coarseMu.Unlock()
				}
			}()
		}
		wg.Wait()
		mu.Lock()
		results["coarse-grained"] = counter
		mu.Unlock()
	}()

	// Lock-free
	func() {
		var counter int64
		var wg sync.WaitGroup
		for g := 0; g < numGoroutines; g++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for i := 0; i < opsPerGoroutine; i++ {
					atomic.AddInt64(&counter, 1)
				}
			}()
		}
		wg.Wait()
		mu.Lock()
		results["lock-free"] = counter
		mu.Unlock()
	}()

	// Thread-local
	func() {
		counters := make([]int64, numGoroutines)
		var wg sync.WaitGroup
		for g := 0; g < numGoroutines; g++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				for i := 0; i < opsPerGoroutine; i++ {
					counters[id]++
				}
			}(g)
		}
		wg.Wait()
		var total int64
		for _, c := range counters {
			total += c
		}
		mu.Lock()
		results["thread-local"] = total
		mu.Unlock()
	}()

	// Verify all approaches produced the same result
	for approach, result := range results {
		if result != int64(expectedTotal) {
			t.Errorf("%s: expected %d, got %d", approach, expectedTotal, result)
		}
	}

	// Verify all approaches agree with each other
	var firstResult int64
	for i, result := range results {
		if firstResult == 0 {
			firstResult = result
		} else if result != firstResult {
			t.Errorf("inconsistency: results differ across approaches")
		}
		t.Logf("%s: %d", i, result)
	}
}

// ===== Phase 0.0.7: Serialization Correctness Tests =====
// Verify that flat vs. nested structures serialize and deserialize correctly for both JSON and msgpack.

// TestFlatSmallJSONRoundTrip verifies JSON round-trip for flat small struct.
func TestFlatSmallJSONRoundTrip(t *testing.T) {
	original := unitSkillEventFlatSample
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var restored UnitSkillEventFlat
	err = json.Unmarshal(data, &restored)
	if err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if restored != original {
		t.Errorf("round-trip mismatch: original %v, restored %v", original, restored)
	}
}

// TestFlatSmallMsgpackRoundTrip verifies msgpack round-trip for flat small struct.
func TestFlatSmallMsgpackRoundTrip(t *testing.T) {
	original := unitSkillEventFlatSample
	data, err := msgpack.Marshal(original)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var restored UnitSkillEventFlat
	err = msgpack.Unmarshal(data, &restored)
	if err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if restored != original {
		t.Errorf("round-trip mismatch: original %v, restored %v", original, restored)
	}
}

// TestNestedSmallJSONRoundTrip verifies JSON round-trip for nested small struct.
func TestNestedSmallJSONRoundTrip(t *testing.T) {
	original := unitSkillEventNestedSample
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var restored UnitSkillEventNested
	err = json.Unmarshal(data, &restored)
	if err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if restored != original {
		t.Errorf("round-trip mismatch: original %v, restored %v", original, restored)
	}
}

// TestNestedSmallMsgpackRoundTrip verifies msgpack round-trip for nested small struct.
func TestNestedSmallMsgpackRoundTrip(t *testing.T) {
	original := unitSkillEventNestedSample
	data, err := msgpack.Marshal(original)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var restored UnitSkillEventNested
	err = msgpack.Unmarshal(data, &restored)
	if err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if restored != original {
		t.Errorf("round-trip mismatch: original %v, restored %v", original, restored)
	}
}

// TestFlatLargeJSONRoundTrip verifies JSON round-trip for flat large struct.
func TestFlatLargeJSONRoundTrip(t *testing.T) {
	original := playerStateFlatSample
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var restored PlayerStateFlat
	err = json.Unmarshal(data, &restored)
	if err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if restored != original {
		t.Errorf("round-trip mismatch: original %v, restored %v", original, restored)
	}
}

// TestFlatLargeMsgpackRoundTrip verifies msgpack round-trip for flat large struct.
func TestFlatLargeMsgpackRoundTrip(t *testing.T) {
	original := playerStateFlatSample
	data, err := msgpack.Marshal(original)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var restored PlayerStateFlat
	err = msgpack.Unmarshal(data, &restored)
	if err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if restored != original {
		t.Errorf("round-trip mismatch: original %v, restored %v", original, restored)
	}
}

// TestNestedLargeJSONRoundTrip verifies JSON round-trip for nested large struct.
func TestNestedLargeJSONRoundTrip(t *testing.T) {
	original := playerStateNestedSample
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var restored PlayerStateNested
	err = json.Unmarshal(data, &restored)
	if err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if restored != original {
		t.Errorf("round-trip mismatch: original %v, restored %v", original, restored)
	}
}

// TestNestedLargeMsgpackRoundTrip verifies msgpack round-trip for nested large struct.
func TestNestedLargeMsgpackRoundTrip(t *testing.T) {
	original := playerStateNestedSample
	data, err := msgpack.Marshal(original)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var restored PlayerStateNested
	err = msgpack.Unmarshal(data, &restored)
	if err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if restored != original {
		t.Errorf("round-trip mismatch: original %v, restored %v", original, restored)
	}
}

// ===== Protobuf Round-Trip Tests =====

// TestFlatSmallProtobufRoundTrip verifies protobuf round-trip for flat small proto message.
func TestFlatSmallProtobufRoundTrip(t *testing.T) {
	original := unitSkillEventFlatPBSample
	data, err := proto.Marshal(original)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var restored serialpb.UnitSkillEventFlatPB
	if err = proto.Unmarshal(data, &restored); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if !proto.Equal(original, &restored) {
		t.Errorf("round-trip mismatch: original %v, restored %v", original, &restored)
	}
}

// TestFlatLargeProtobufRoundTrip verifies protobuf round-trip for flat large proto message.
func TestFlatLargeProtobufRoundTrip(t *testing.T) {
	original := playerStateFlatPBSample
	data, err := proto.Marshal(original)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var restored serialpb.PlayerStateFlatPB
	if err = proto.Unmarshal(data, &restored); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if !proto.Equal(original, &restored) {
		t.Errorf("round-trip mismatch: original %v, restored %v", original, &restored)
	}
}

// TestNestedSmallProtobufRoundTrip verifies protobuf round-trip for nested small proto message.
func TestNestedSmallProtobufRoundTrip(t *testing.T) {
	original := unitSkillEventNestedPBSample
	data, err := proto.Marshal(original)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var restored serialpb.UnitSkillEventNestedPB
	if err = proto.Unmarshal(data, &restored); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if !proto.Equal(original, &restored) {
		t.Errorf("round-trip mismatch: original %v, restored %v", original, &restored)
	}
}

// TestNestedLargeProtobufRoundTrip verifies protobuf round-trip for nested large proto message.
func TestNestedLargeProtobufRoundTrip(t *testing.T) {
	original := playerStateNestedPBSample
	data, err := proto.Marshal(original)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var restored serialpb.PlayerStateNestedPB
	if err = proto.Unmarshal(data, &restored); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if !proto.Equal(original, &restored) {
		t.Errorf("round-trip mismatch: original %v, restored %v", original, &restored)
	}
}

// ===== Inventory Struct Correctness Tests =====

// TestInventoryStructGetAll verifies InventoryStruct returns correct pointers for all keys.
func TestInventoryStructGetAll(t *testing.T) {
	// Static keys A, B, C must return the exact pointers
	if got := inventoryStruct1.GetInventory("A"); got != invA {
		t.Errorf("GetInventory(\"A\"): expected %p, got %p", invA, got)
	}
	if got := inventoryStruct1.GetInventory("B"); got != invB {
		t.Errorf("GetInventory(\"B\"): expected %p, got %p", invB, got)
	}
	if got := inventoryStruct1.GetInventory("C"); got != invC {
		t.Errorf("GetInventory(\"C\"): expected %p, got %p", invC, got)
	}
	// Dynamic key D must return the fallthrough value
	if got := inventoryStruct1.GetInventory("D"); got != invD {
		t.Errorf("GetInventory(\"D\"): expected %p, got %p", invD, got)
	}
	// Unknown key returns nil from the map (no panic)
	if got := inventoryStruct1.GetInventory("Unknown"); got != nil {
		t.Errorf("GetInventory(\"Unknown\"): expected nil, got %p", got)
	}
}

// TestInventoryStruct2GetAll verifies InventoryStruct2 map returns correct pointers for all keys.
func TestInventoryStruct2GetAll(t *testing.T) {
	// All keys A, B, C, D must return the exact pointers
	if got := inventoryStruct2.GetInventory("A"); got != invA {
		t.Errorf("GetInventory(\"A\"): expected %p, got %p", invA, got)
	}
	if got := inventoryStruct2.GetInventory("B"); got != invB {
		t.Errorf("GetInventory(\"B\"): expected %p, got %p", invB, got)
	}
	if got := inventoryStruct2.GetInventory("C"); got != invC {
		t.Errorf("GetInventory(\"C\"): expected %p, got %p", invC, got)
	}
	if got := inventoryStruct2.GetInventory("D"); got != invD {
		t.Errorf("GetInventory(\"D\"): expected %p, got %p", invD, got)
	}
	// Unknown key returns nil from the map
	if got := inventoryStruct2.GetInventory("Unknown"); got != nil {
		t.Errorf("GetInventory(\"Unknown\"): expected nil, got %p", got)
	}
}

// TestInventoryStruct3GetByIndex verifies InventoryStruct3 slice index access returns correct pointers.
func TestInventoryStruct3GetByIndex(t *testing.T) {
	// Direct index access: 0→A, 1→B, 2→C
	if got := inventoryStruct3.GetInventory(0); got != invA {
		t.Errorf("GetInventory(0): expected %p, got %p", invA, got)
	}
	if got := inventoryStruct3.GetInventory(1); got != invB {
		t.Errorf("GetInventory(1): expected %p, got %p", invB, got)
	}
	if got := inventoryStruct3.GetInventory(2); got != invC {
		t.Errorf("GetInventory(2): expected %p, got %p", invC, got)
	}
}

// TestInventoryStruct3GetByName verifies InventoryStruct3 linear scan returns correct pointers.
func TestInventoryStruct3GetByName(t *testing.T) {
	// Linear scan by name field
	if got := inventoryStruct3.GetInventoryByName("A"); got != invA {
		t.Errorf("GetInventoryByName(\"A\"): expected %p, got %p", invA, got)
	}
	if got := inventoryStruct3.GetInventoryByName("B"); got != invB {
		t.Errorf("GetInventoryByName(\"B\"): expected %p, got %p", invB, got)
	}
	if got := inventoryStruct3.GetInventoryByName("C"); got != invC {
		t.Errorf("GetInventoryByName(\"C\"): expected %p, got %p", invC, got)
	}
	// Key not in slice returns nil
	if got := inventoryStruct3.GetInventoryByName("D"); got != nil {
		t.Errorf("GetInventoryByName(\"D\"): expected nil, got %p", got)
	}
}
