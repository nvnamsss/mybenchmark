package mybenchmark

import "testing"

// BenchmarkScalarInt8x64ExtendToInt16 is the scalar baseline for BenchmarkSIMDInt8x64ExtendToInt16.
// Widens all 64 int8 lanes to int16, writing the low 32 to dstLo and high 32 to dstHi,
// mirroring the GetLo/GetHi split of the SIMD path.
func BenchmarkScalarInt8x64ExtendToInt16(b *testing.B) {
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

	b.ResetTimer()
	for b.Loop() {
		for i := range 32 {
			dstLo[i] = int16(src[i])
		}
		for i := range 32 {
			dstHi[i] = int16(src[i+32])
		}
	}
	_ = dstLo
	_ = dstHi
}

// BenchmarkScalarFloat32x16ConvertToInt32 is the scalar baseline for BenchmarkSIMDFloat32x16ConvertToInt32.
// Truncates 16 float32 values toward zero to int32, matching VCVTTPS2DQ semantics.
func BenchmarkScalarFloat32x16ConvertToInt32(b *testing.B) {
	src := [16]float32{1.7, -2.9, 3.1, -4.5, 5.0, -6.6, 7.3, -8.8,
		9.9, -10.1, 11.2, -12.7, 13.4, -14.0, 15.5, -16.3}
	dst := [16]int32{}

	b.ResetTimer()
	for b.Loop() {
		for i := range 16 {
			dst[i] = int32(src[i])
		}
	}
	_ = dst
}

// BenchmarkScalarInt32x16SaturateToInt16ConcatGrouped is the scalar baseline for
// BenchmarkSIMDInt32x16SaturateToInt16ConcatGrouped.
// Saturates 32 int32 values (two groups of 16) to [-32768, 32767] and writes to [32]int16.
func BenchmarkScalarInt32x16SaturateToInt16ConcatGrouped(b *testing.B) {
	src1 := [16]int32{100000, -100000, 32767, -32768, 1, -1, 50000, -50000,
		200, -200, 32000, -32000, 99999, -99999, 16384, -16384}
	src2 := [16]int32{200000, -200000, 32767, -32768, 2, -2, 60000, -60000,
		300, -300, 31000, -31000, 88888, -88888, 8192, -8192}
	dst := [32]int16{}

	b.ResetTimer()
	for b.Loop() {
		for i := range 16 {
			dst[i] = int16(min(max(src1[i], -32768), 32767))
		}
		for i := range 16 {
			dst[i+16] = int16(min(max(src2[i], -32768), 32767))
		}
	}
	_ = dst
}
