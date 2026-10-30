//go:build goexperiment.simd

package mybenchmark

import (
	"simd/archsimd"
	"testing"
)

// BenchmarkSIMDInt32x16GreaterMasked benchmarks a compare-then-zero-mask select:
// computes a.Greater(b) → Mask32x16, then a.Masked(mask) to zero out non-matching lanes.
// Scalar baseline: BenchmarkScalarInt32x16GreaterMasked.
func BenchmarkSIMDInt32x16GreaterMasked(b *testing.B) {
	src1 := [16]int32{10, 2, 30, 4, 50, 6, 70, 8, 90, 10, 110, 12, 130, 14, 150, 16}
	src2 := [16]int32{1, 20, 3, 40, 5, 60, 7, 80, 9, 100, 11, 120, 13, 140, 15, 160}
	dst := [16]int32{}

	a := archsimd.LoadInt32x16(&src1)
	bv := archsimd.LoadInt32x16(&src2)
	b.ResetTimer()
	for b.Loop() {
		mask := a.Greater(bv)
		result := a.Masked(mask)
		result.Store(&dst)
	}
}

// BenchmarkSIMDUint8x64Compress benchmarks VPCOMPRESSB: packs active (mask=1) lanes
// contiguously to the front of the destination vector.
// Uses alternating mask 0xAAAA...AAAA (every other lane active — 32 of 64).
// Scalar baseline: BenchmarkScalarUint8x64Compress.
func BenchmarkSIMDUint8x64Compress(b *testing.B) {
	src := [64]uint8{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16,
		17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32,
		33, 34, 35, 36, 37, 38, 39, 40, 41, 42, 43, 44, 45, 46, 47, 48,
		49, 50, 51, 52, 53, 54, 55, 56, 57, 58, 59, 60, 61, 62, 63, 64}
	dst := [64]uint8{}

	a := archsimd.LoadUint8x64(&src)
	mask := archsimd.Mask8x64FromBits(0xAAAAAAAAAAAAAAAA)
	b.ResetTimer()
	for b.Loop() {
		result := a.Compress(mask)
		result.Store(&dst)
	}
}

// BenchmarkSIMDUint8x64Expand benchmarks VPEXPANDB: scatters packed source lanes
// into positions indicated by the mask, zeroing non-mask positions.
// Uses the same alternating mask as BenchmarkSIMDUint8x64Compress.
// Scalar baseline: BenchmarkScalarUint8x64Expand.
func BenchmarkSIMDUint8x64Expand(b *testing.B) {
	src := [64]uint8{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16,
		17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32,
		33, 34, 35, 36, 37, 38, 39, 40, 41, 42, 43, 44, 45, 46, 47, 48,
		49, 50, 51, 52, 53, 54, 55, 56, 57, 58, 59, 60, 61, 62, 63, 64}
	dst := [64]uint8{}

	a := archsimd.LoadUint8x64(&src)
	mask := archsimd.Mask8x64FromBits(0xAAAAAAAAAAAAAAAA)
	b.ResetTimer()
	for b.Loop() {
		result := a.Expand(mask)
		result.Store(&dst)
	}
}
