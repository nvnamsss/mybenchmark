//go:build goexperiment.simd

package mybenchmark

import (
	"simd/archsimd"
	"testing"
)

// BenchmarkSIMDInt8x64ExtendToInt16 benchmarks widening 64 int8 lanes to int16.
// Since no 1024-bit register exists, the operation splits Int8x64 into two Int8x32
// halves via GetLo/GetHi, then widens each half to Int16x32 independently.
// Scalar baseline: BenchmarkScalarInt8x64ExtendToInt16.
func BenchmarkSIMDInt8x64ExtendToInt16(b *testing.B) {
	src := [64]int8{-64, -63, -62, -61, -60, -59, -58, -57,
		-56, -55, -54, -53, -52, -51, -50, -49,
		-48, -47, -46, -45, -44, -43, -42, -41,
		-40, -39, -38, -37, -36, -35, -34, -33,
		1, 2, 3, 4, 5, 6, 7, 8,
		9, 10, 11, 12, 13, 14, 15, 16,
		17, 18, 19, 20, 21, 22, 23, 24,
		25, 26, 27, 28, 29, 30, 31, 32}
	dstLo := [32]int16{}
	dstHi := [32]int16{}

	v := archsimd.LoadInt8x64(&src)
	b.ResetTimer()
	for b.Loop() {
		v.GetLo().ExtendToInt16().Store(&dstLo)
		v.GetHi().ExtendToInt16().Store(&dstHi)
	}
}

// BenchmarkSIMDFloat32x16ConvertToInt32 benchmarks VCVTTPS2DQ: truncate-toward-zero
// conversion of 16 float32 lanes to int32 simultaneously.
// Scalar baseline: BenchmarkScalarFloat32x16ConvertToInt32.
func BenchmarkSIMDFloat32x16ConvertToInt32(b *testing.B) {
	src := [16]float32{1.7, -2.9, 3.1, -4.5, 5.0, -6.6, 7.3, -8.8,
		9.9, -10.1, 11.2, -12.7, 13.4, -14.0, 15.5, -16.3}
	dst := [16]int32{}

	v := archsimd.LoadFloat32x16(&src)
	b.ResetTimer()
	for b.Loop() {
		v.ConvertToInt32().Store(&dst)
	}
}

// BenchmarkSIMDInt32x16SaturateToInt16ConcatGrouped benchmarks VPACKSSDW:
// saturating narrowing of two Int32x16 vectors to one Int16x32, interleaving
// groups from each source. Values outside [-32768, 32767] are clamped.
// Scalar baseline: BenchmarkScalarInt32x16SaturateToInt16ConcatGrouped.
func BenchmarkSIMDInt32x16SaturateToInt16ConcatGrouped(b *testing.B) {
	src1 := [16]int32{100000, -100000, 32767, -32768, 1, -1, 50000, -50000,
		200, -200, 32000, -32000, 99999, -99999, 16384, -16384}
	src2 := [16]int32{200000, -200000, 32767, -32768, 2, -2, 60000, -60000,
		300, -300, 31000, -31000, 88888, -88888, 8192, -8192}
	dst := [32]int16{}

	a := archsimd.LoadInt32x16(&src1)
	bv := archsimd.LoadInt32x16(&src2)
	b.ResetTimer()
	for b.Loop() {
		a.SaturateToInt16ConcatGrouped(bv).Store(&dst)
	}
}
