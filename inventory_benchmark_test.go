package mybenchmark

import (
	"testing"
)

// Minimal stub — benchmark measures lookup cost, not field access cost.
type Inventory struct {
	Name string
}

// Declarative: static fields for known keys; dynamic map for extras.
type InventoryStruct struct {
	A       *Inventory
	B       *Inventory
	C       *Inventory
	Dynamic map[string]*Inventory
}

// GetInventory returns an inventory by name: switch on A/B/C, else consult Dynamic map.
func (s *InventoryStruct) GetInventory(name string) *Inventory {
	switch name {
	case "A":
		return s.A
	case "B":
		return s.B
	case "C":
		return s.C
	default:
		return s.Dynamic[name]
	}
}

// Fully dynamic: all keys live in a single map.
type InventoryStruct2 struct {
	inventories map[string]*Inventory
}

// GetInventory returns an inventory by name via map lookup.
func (s *InventoryStruct2) GetInventory(name string) *Inventory {
	return s.inventories[name]
}

// Slice-backed: index-based O(1) access; name lookup via linear scan.
type InventoryStruct3 struct {
	inventories []*Inventory
}

// GetInventory returns an inventory by direct index.
func (s *InventoryStruct3) GetInventory(index int) *Inventory {
	return s.inventories[index]
}

// GetInventoryByName returns an inventory by name via linear scan.
func (s *InventoryStruct3) GetInventoryByName(name string) *Inventory {
	for _, inv := range s.inventories {
		if inv.Name == name {
			return inv
		}
	}
	return nil
}

// ===== Integer Key Variants =====

// Declarative: static fields for known integer keys; dynamic map for extras.
type InventoryStructInt struct {
	K1      *Inventory
	K2      *Inventory
	K3      *Inventory
	Dynamic map[int]*Inventory
}

// GetInventory returns an inventory by integer key: switch on 1/2/3, else consult Dynamic map.
func (s *InventoryStructInt) GetInventory(key int) *Inventory {
	switch key {
	case 1:
		return s.K1
	case 2:
		return s.K2
	case 3:
		return s.K3
	default:
		return s.Dynamic[key]
	}
}

// Fully dynamic: all integer keys live in a single map.
type InventoryStruct2Int struct {
	inventories map[int]*Inventory
}

// GetInventory returns an inventory by integer key via map lookup.
func (s *InventoryStruct2Int) GetInventory(key int) *Inventory {
	return s.inventories[key]
}

// Package-level fixtures — prevent heap escape, keep setup cost out of benchmark loop.
var (
	invA = &Inventory{Name: "A"}
	invB = &Inventory{Name: "B"}
	invC = &Inventory{Name: "C"}
	invD = &Inventory{Name: "D"}

	inventoryStruct1 = &InventoryStruct{
		A: invA, B: invB, C: invC,
		Dynamic: map[string]*Inventory{"D": invD},
	}
	inventoryStruct2 = &InventoryStruct2{
		inventories: map[string]*Inventory{"A": invA, "B": invB, "C": invC, "D": invD},
	}
	inventoryStruct3 = &InventoryStruct3{
		inventories: []*Inventory{invA, invB, invC},
	}

	// Integer key fixtures
	inventoryStructInt = &InventoryStructInt{
		K1: invA, K2: invB, K3: invC,
		Dynamic: map[int]*Inventory{4: invD},
	}
	inventoryStruct2Int = &InventoryStruct2Int{
		inventories: map[int]*Inventory{1: invA, 2: invB, 3: invC, 4: invD},
	}
)

// ===== InventoryStruct Benchmarks (switch + dynamic fallthrough) =====

// BenchmarkInventoryStructGetA measures InventoryStruct.GetInventory("A") — switch branch 1.
func BenchmarkInventoryStructGetA(b *testing.B) {
	b.ReportAllocs()
	var dst *Inventory
	for b.Loop() {
		dst = inventoryStruct1.GetInventory("A")
	}
	_ = dst
}

// BenchmarkInventoryStructGetB measures InventoryStruct.GetInventory("B") — switch branch 2.
func BenchmarkInventoryStructGetB(b *testing.B) {
	b.ReportAllocs()
	var dst *Inventory
	for b.Loop() {
		dst = inventoryStruct1.GetInventory("B")
	}
	_ = dst
}

// BenchmarkInventoryStructGetC measures InventoryStruct.GetInventory("C") — switch branch 3.
func BenchmarkInventoryStructGetC(b *testing.B) {
	b.ReportAllocs()
	var dst *Inventory
	for b.Loop() {
		dst = inventoryStruct1.GetInventory("C")
	}
	_ = dst
}

// BenchmarkInventoryStructGetDynamic measures InventoryStruct.GetInventory("D") — map fallthrough.
func BenchmarkInventoryStructGetDynamic(b *testing.B) {
	b.ReportAllocs()
	var dst *Inventory
	for b.Loop() {
		dst = inventoryStruct1.GetInventory("D")
	}
	_ = dst
}

// ===== InventoryStruct2 Benchmarks (pure map) =====

// BenchmarkInventoryStruct2GetA measures InventoryStruct2.GetInventory("A") via map.
func BenchmarkInventoryStruct2GetA(b *testing.B) {
	b.ReportAllocs()
	var dst *Inventory
	for b.Loop() {
		dst = inventoryStruct2.GetInventory("A")
	}
	_ = dst
}

// BenchmarkInventoryStruct2GetB measures InventoryStruct2.GetInventory("B") via map.
func BenchmarkInventoryStruct2GetB(b *testing.B) {
	b.ReportAllocs()
	var dst *Inventory
	for b.Loop() {
		dst = inventoryStruct2.GetInventory("B")
	}
	_ = dst
}

// BenchmarkInventoryStruct2GetC measures InventoryStruct2.GetInventory("C") via map.
func BenchmarkInventoryStruct2GetC(b *testing.B) {
	b.ReportAllocs()
	var dst *Inventory
	for b.Loop() {
		dst = inventoryStruct2.GetInventory("C")
	}
	_ = dst
}

// BenchmarkInventoryStruct2GetDynamic measures InventoryStruct2.GetInventory("D") via map.
func BenchmarkInventoryStruct2GetDynamic(b *testing.B) {
	b.ReportAllocs()
	var dst *Inventory
	for b.Loop() {
		dst = inventoryStruct2.GetInventory("D")
	}
	_ = dst
}

// ===== InventoryStruct3 Benchmarks (slice — name scan) =====

// BenchmarkInventoryStruct3GetByNameA measures InventoryStruct3.GetInventoryByName("A") — found at index 0.
func BenchmarkInventoryStruct3GetByNameA(b *testing.B) {
	b.ReportAllocs()
	var dst *Inventory
	for b.Loop() {
		dst = inventoryStruct3.GetInventoryByName("A")
	}
	_ = dst
}

// BenchmarkInventoryStruct3GetByNameB measures InventoryStruct3.GetInventoryByName("B") — found at index 1.
func BenchmarkInventoryStruct3GetByNameB(b *testing.B) {
	b.ReportAllocs()
	var dst *Inventory
	for b.Loop() {
		dst = inventoryStruct3.GetInventoryByName("B")
	}
	_ = dst
}

// BenchmarkInventoryStruct3GetByNameC measures InventoryStruct3.GetInventoryByName("C") — found at index 2 (worst-case scan).
func BenchmarkInventoryStruct3GetByNameC(b *testing.B) {
	b.ReportAllocs()
	var dst *Inventory
	for b.Loop() {
		dst = inventoryStruct3.GetInventoryByName("C")
	}
	_ = dst
}

// BenchmarkInventoryStruct3GetByNameNotFound measures InventoryStruct3.GetInventoryByName("D") — full traversal, returns nil.
func BenchmarkInventoryStruct3GetByNameNotFound(b *testing.B) {
	b.ReportAllocs()
	var dst *Inventory
	for b.Loop() {
		dst = inventoryStruct3.GetInventoryByName("D")
	}
	_ = dst
}

// ===== InventoryStruct3 Benchmarks (slice — direct index) =====

// BenchmarkInventoryStruct3GetByIndex0 measures InventoryStruct3.GetInventory(0) — direct slice index.
func BenchmarkInventoryStruct3GetByIndex0(b *testing.B) {
	b.ReportAllocs()
	var dst *Inventory
	for b.Loop() {
		dst = inventoryStruct3.GetInventory(0)
	}
	_ = dst
}

// BenchmarkInventoryStruct3GetByIndex1 measures InventoryStruct3.GetInventory(1) — direct slice index.
func BenchmarkInventoryStruct3GetByIndex1(b *testing.B) {
	b.ReportAllocs()
	var dst *Inventory
	for b.Loop() {
		dst = inventoryStruct3.GetInventory(1)
	}
	_ = dst
}

// BenchmarkInventoryStruct3GetByIndex2 measures InventoryStruct3.GetInventory(2) — direct slice index.
func BenchmarkInventoryStruct3GetByIndex2(b *testing.B) {
	b.ReportAllocs()
	var dst *Inventory
	for b.Loop() {
		dst = inventoryStruct3.GetInventory(2)
	}
	_ = dst
}

// ===== InventoryStructInt Benchmarks (int key switch + dynamic fallthrough) =====

// BenchmarkInventoryStructIntGetK1 measures InventoryStructInt.GetInventory(1) — switch branch 1.
func BenchmarkInventoryStructIntGetK1(b *testing.B) {
	b.ReportAllocs()
	var dst *Inventory
	for b.Loop() {
		dst = inventoryStructInt.GetInventory(1)
	}
	_ = dst
}

// BenchmarkInventoryStructIntGetK2 measures InventoryStructInt.GetInventory(2) — switch branch 2.
func BenchmarkInventoryStructIntGetK2(b *testing.B) {
	b.ReportAllocs()
	var dst *Inventory
	for b.Loop() {
		dst = inventoryStructInt.GetInventory(2)
	}
	_ = dst
}

// BenchmarkInventoryStructIntGetK3 measures InventoryStructInt.GetInventory(3) — switch branch 3.
func BenchmarkInventoryStructIntGetK3(b *testing.B) {
	b.ReportAllocs()
	var dst *Inventory
	for b.Loop() {
		dst = inventoryStructInt.GetInventory(3)
	}
	_ = dst
}

// BenchmarkInventoryStructIntGetDynamic measures InventoryStructInt.GetInventory(4) — map fallthrough.
func BenchmarkInventoryStructIntGetDynamic(b *testing.B) {
	b.ReportAllocs()
	var dst *Inventory
	for b.Loop() {
		dst = inventoryStructInt.GetInventory(4)
	}
	_ = dst
}

// ===== InventoryStruct2Int Benchmarks (pure map with int keys) =====

// BenchmarkInventoryStruct2IntGetK1 measures InventoryStruct2Int.GetInventory(1) via map.
func BenchmarkInventoryStruct2IntGetK1(b *testing.B) {
	b.ReportAllocs()
	var dst *Inventory
	for b.Loop() {
		dst = inventoryStruct2Int.GetInventory(1)
	}
	_ = dst
}

// BenchmarkInventoryStruct2IntGetK2 measures InventoryStruct2Int.GetInventory(2) via map.
func BenchmarkInventoryStruct2IntGetK2(b *testing.B) {
	b.ReportAllocs()
	var dst *Inventory
	for b.Loop() {
		dst = inventoryStruct2Int.GetInventory(2)
	}
	_ = dst
}

// BenchmarkInventoryStruct2IntGetK3 measures InventoryStruct2Int.GetInventory(3) via map.
func BenchmarkInventoryStruct2IntGetK3(b *testing.B) {
	b.ReportAllocs()
	var dst *Inventory
	for b.Loop() {
		dst = inventoryStruct2Int.GetInventory(3)
	}
	_ = dst
}

// BenchmarkInventoryStruct2IntGetDynamic measures InventoryStruct2Int.GetInventory(4) via map.
func BenchmarkInventoryStruct2IntGetDynamic(b *testing.B) {
	b.ReportAllocs()
	var dst *Inventory
	for b.Loop() {
		dst = inventoryStruct2Int.GetInventory(4)
	}
	_ = dst
}
