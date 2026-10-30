//go:build goexperiment.simd

package mybenchmark

import (
	"simd/archsimd"
	"testing"
)

// BenchmarkSIMDUint8x64Add benchmarks 512-bit element-wise addition over 64 uint8 lanes.
// Scalar baseline: BenchmarkScalarUint8x64Add.
func BenchmarkSIMDUint8x64Add(b *testing.B) {
	src1 := [64]uint8{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16,
		17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32,
		33, 34, 35, 36, 37, 38, 39, 40, 41, 42, 43, 44, 45, 46, 47, 48,
		49, 50, 51, 52, 53, 54, 55, 56, 57, 58, 59, 60, 61, 62, 63, 64}
	src2 := [64]uint8{64, 63, 62, 61, 60, 59, 58, 57, 56, 55, 54, 53, 52, 51, 50, 49,
		48, 47, 46, 45, 44, 43, 42, 41, 40, 39, 38, 37, 36, 35, 34, 33,
		32, 31, 30, 29, 28, 27, 26, 25, 24, 23, 22, 21, 20, 19, 18, 17,
		16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}
	dst := [64]uint8{}

	a := archsimd.LoadUint8x64(&src1)
	bv := archsimd.LoadUint8x64(&src2)
	b.ResetTimer()
	for b.Loop() {
		result := a.Add(bv)
		result.Store(&dst)
	}
}

// BenchmarkSIMDInt32x16Mul benchmarks 512-bit element-wise multiplication over 16 int32 lanes.
// Scalar baseline: BenchmarkScalarInt32x16Mul.
func BenchmarkSIMDInt32x16Mul(b *testing.B) {
	src1 := [16]int32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	src2 := [16]int32{16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}
	dst := [16]int32{}

	a := archsimd.LoadInt32x16(&src1)
	bv := archsimd.LoadInt32x16(&src2)
	b.ResetTimer()
	for b.Loop() {
		result := a.Mul(bv)
		result.Store(&dst)
	}
}

// BenchmarkSIMDInt16x32DotProductPairs benchmarks 512-bit dot product over 32 int16 lanes,
// producing 16 int32 results (each output is the sum of two adjacent int16 products).
// Scalar baseline: BenchmarkScalarInt16x32DotProductPairs.
func BenchmarkSIMDInt16x32DotProductPairs(b *testing.B) {
	src1 := [32]int16{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16,
		17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32}
	src2 := [32]int16{32, 31, 30, 29, 28, 27, 26, 25, 24, 23, 22, 21, 20, 19, 18, 17,
		16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}
	dst := [16]int32{}

	a := archsimd.LoadInt16x32(&src1)
	bv := archsimd.LoadInt16x32(&src2)
	b.ResetTimer()
	for b.Loop() {
		result := a.DotProductPairs(bv)
		result.Store(&dst)
	}
}

// BenchmarkSIMDUint8x64SumAbsDiff benchmarks 512-bit sum of absolute differences over 64 uint8
// lanes, producing 32 uint16 results (one SAD per group of 2 lanes).
// Scalar baseline: BenchmarkScalarUint8x64SumAbsDiff.
func BenchmarkSIMDUint8x64SumAbsDiff(b *testing.B) {
	src1 := [64]uint8{10, 20, 30, 40, 50, 60, 70, 80, 90, 100, 110, 120, 130, 140, 150, 160,
		170, 180, 190, 200, 210, 220, 230, 240, 250, 5, 15, 25, 35, 45, 55, 65,
		75, 85, 95, 105, 115, 125, 135, 145, 155, 165, 175, 185, 195, 205, 215, 225,
		235, 245, 1, 11, 21, 31, 41, 51, 61, 71, 81, 91, 101, 111, 121, 131}
	src2 := [64]uint8{5, 15, 25, 35, 45, 55, 65, 75, 85, 95, 105, 115, 125, 135, 145, 155,
		165, 175, 185, 195, 205, 215, 225, 235, 245, 10, 20, 30, 40, 50, 60, 70,
		80, 90, 100, 110, 120, 130, 140, 150, 160, 170, 180, 190, 200, 210, 220, 230,
		240, 250, 2, 12, 22, 32, 42, 52, 62, 72, 82, 92, 102, 112, 122, 132}
	dst := [32]uint16{}

	a := archsimd.LoadUint8x64(&src1)
	bv := archsimd.LoadUint8x64(&src2)
	b.ResetTimer()
	for b.Loop() {
		result := a.SumAbsDiff(bv)
		result.Store(&dst)
	}
}

// BenchmarkSIMDInt32x16OnesCount benchmarks 512-bit popcount over 16 int32 lanes.
// Scalar baseline: BenchmarkScalarInt32x16OnesCount.
func BenchmarkSIMDInt32x16OnesCount(b *testing.B) {
	src := [16]int32{0x0F0F0F0F, 0x12345678, -559038737, -889275714,
		0x01010101, -16711936, -1431655766, 0x55555555,
		0x00FF00FF, -252645136, 0x0000FFFF, -65536,
		0x11111111, 0x22222222, 0x44444444, -2004318072}
	dst := [16]int32{}

	a := archsimd.LoadInt32x16(&src)
	b.ResetTimer()
	for b.Loop() {
		result := a.OnesCount()
		result.Store(&dst)
	}
}
