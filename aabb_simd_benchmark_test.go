//go:build goexperiment.simd

package mybenchmark

import (
	"simd/archsimd"
	"testing"
)

// AABB16 is a Structure-of-Arrays layout for 16 AABBs.
// Packing all MinX values together (and so on) enables SIMD to process all 16
// boxes in a single 512-bit register load, rather than iterating one box at a time.
type AABB16 struct {
	MinX, MinY, MinZ [16]float32
	MaxX, MaxY, MaxZ [16]float32
}

// makeSimdBoxes returns the same 16 boxes as makeScalarBoxes but in SoA layout.
// box[i].Min=(i+1,i+1,i+1)  box[i].Max=(i+6,i+6,i+6)  i=0..15
func makeSimdBoxes() AABB16 {
	var b AABB16
	for i := range b.MinX {
		f := float32(i + 1)
		b.MinX[i] = f
		b.MinY[i] = f
		b.MinZ[i] = f
		b.MaxX[i] = f + 5
		b.MaxY[i] = f + 5
		b.MaxZ[i] = f + 5
	}
	return b
}

// BenchmarkSIMDAABBIntersect16 tests 16 AABB intersections against a query box
// in a single SIMD pass. Computes the separating-axis gap for each axis pair via
// Float32x16.Sub+Max; a single LessEqual(zero) at the end converts to a bitmask.
// Avoids mask-register chaining (VPTERNLOGD spill issue) by keeping all
// intermediate results in ZMM registers as float32 vectors.
// Scalar baseline: BenchmarkScalarAABBIntersect16.
func BenchmarkSIMDAABBIntersect16(b *testing.B) {
	boxes := makeSimdBoxes()

	// Broadcast the 6 query-box scalars once outside the loop.
	qMinX := archsimd.BroadcastFloat32x16(3.0)
	qMaxX := archsimd.BroadcastFloat32x16(10.0)
	qMinY := archsimd.BroadcastFloat32x16(3.0)
	qMaxY := archsimd.BroadcastFloat32x16(10.0)
	qMinZ := archsimd.BroadcastFloat32x16(3.0)
	qMaxZ := archsimd.BroadcastFloat32x16(10.0)

	// Load all 16 box extents once outside the loop.
	boxMinX := archsimd.LoadFloat32x16(&boxes.MinX)
	boxMaxX := archsimd.LoadFloat32x16(&boxes.MaxX)
	boxMinY := archsimd.LoadFloat32x16(&boxes.MinY)
	boxMaxY := archsimd.LoadFloat32x16(&boxes.MaxY)
	boxMinZ := archsimd.LoadFloat32x16(&boxes.MinZ)
	boxMaxZ := archsimd.LoadFloat32x16(&boxes.MaxZ)

	zeroArr := [16]float32{}
	zero := archsimd.LoadFloat32x16(&zeroArr)

	var dst uint16
	b.ResetTimer()
	for b.Loop() {
		// Separating-axis gap per axis pair.
		// gap > 0 means there is separation (no overlap) on that axis.
		gapX := boxMinX.Sub(qMaxX).Max(qMinX.Sub(boxMaxX))
		gapY := boxMinY.Sub(qMaxY).Max(qMinY.Sub(boxMaxY))
		gapZ := boxMinZ.Sub(qMaxZ).Max(qMinZ.Sub(boxMaxZ))
		// Maximum gap across all 3 axes: <= 0 means intersection.
		maxGap := gapX.Max(gapY).Max(gapZ)
		// Single mask operation at the end.
		dst = maxGap.LessEqual(zero).ToBits()
	}
	_ = dst
}

// BenchmarkSIMDAABBContainsPoint16 tests which of 16 AABBs contain point (5,5,5).
// Uses the same separating-axis gap approach as BenchmarkSIMDAABBIntersect16:
// compute max(boxMin-p, p-boxMax) per axis, then max across axes; <= 0 → inside.
// Scalar baseline: BenchmarkScalarAABBContainsPoint16.
func BenchmarkSIMDAABBContainsPoint16(b *testing.B) {
	boxes := makeSimdBoxes()

	pX := archsimd.BroadcastFloat32x16(5.0)
	pY := archsimd.BroadcastFloat32x16(5.0)
	pZ := archsimd.BroadcastFloat32x16(5.0)

	boxMinX := archsimd.LoadFloat32x16(&boxes.MinX)
	boxMaxX := archsimd.LoadFloat32x16(&boxes.MaxX)
	boxMinY := archsimd.LoadFloat32x16(&boxes.MinY)
	boxMaxY := archsimd.LoadFloat32x16(&boxes.MaxY)
	boxMinZ := archsimd.LoadFloat32x16(&boxes.MinZ)
	boxMaxZ := archsimd.LoadFloat32x16(&boxes.MaxZ)

	zeroArr := [16]float32{}
	zero := archsimd.LoadFloat32x16(&zeroArr)

	var dst uint16
	b.ResetTimer()
	for b.Loop() {
		// gap > 0: p is outside the box on that axis.
		gapX := boxMinX.Sub(pX).Max(pX.Sub(boxMaxX))
		gapY := boxMinY.Sub(pY).Max(pY.Sub(boxMaxY))
		gapZ := boxMinZ.Sub(pZ).Max(pZ.Sub(boxMaxZ))
		maxGap := gapX.Max(gapY).Max(gapZ)
		dst = maxGap.LessEqual(zero).ToBits()
	}
	_ = dst
}

// BenchmarkSIMDAABBExpand16 expands 16 AABBs outward by delta=1.0 using pure
// SIMD arithmetic: 6 loads, 6 Sub/Add, 6 stores — 0 allocs.
// Scalar baseline: BenchmarkScalarAABBExpand16.
func BenchmarkSIMDAABBExpand16(b *testing.B) {
	src := makeSimdBoxes()
	var dst AABB16

	delta := archsimd.BroadcastFloat32x16(1.0)
	minX := archsimd.LoadFloat32x16(&src.MinX)
	minY := archsimd.LoadFloat32x16(&src.MinY)
	minZ := archsimd.LoadFloat32x16(&src.MinZ)
	maxX := archsimd.LoadFloat32x16(&src.MaxX)
	maxY := archsimd.LoadFloat32x16(&src.MaxY)
	maxZ := archsimd.LoadFloat32x16(&src.MaxZ)

	b.ResetTimer()
	for b.Loop() {
		minX.Sub(delta).Store(&dst.MinX)
		minY.Sub(delta).Store(&dst.MinY)
		minZ.Sub(delta).Store(&dst.MinZ)
		maxX.Add(delta).Store(&dst.MaxX)
		maxY.Add(delta).Store(&dst.MaxY)
		maxZ.Add(delta).Store(&dst.MaxZ)
	}
	_ = dst
}
