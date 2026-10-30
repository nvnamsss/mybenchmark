package mybenchmark

import "testing"

// BenchmarkScalarInt32x16GreaterMasked is the scalar baseline for BenchmarkSIMDInt32x16GreaterMasked.
// For each lane: dst[i] = src1[i] if src1[i] > src2[i], else 0.
func BenchmarkScalarInt32x16GreaterMasked(b *testing.B) {
	src1 := [16]int32{10, 2, 30, 4, 50, 6, 70, 8, 90, 10, 110, 12, 130, 14, 150, 16}
	src2 := [16]int32{1, 20, 3, 40, 5, 60, 7, 80, 9, 100, 11, 120, 13, 140, 15, 160}
	dst := [16]int32{}

	b.ResetTimer()
	for b.Loop() {
		for i := range 16 {
			if src1[i] > src2[i] {
				dst[i] = src1[i]
			} else {
				dst[i] = 0
			}
		}
	}
	_ = dst
}

// BenchmarkScalarUint8x64Compress is the scalar baseline for BenchmarkSIMDUint8x64Compress.
// Packs lanes where the corresponding mask bit is 1 contiguously to the front of dst.
// Uses the same alternating mask 0xAAAAAAAAAAAAAAAA as the SIMD benchmark.
func BenchmarkScalarUint8x64Compress(b *testing.B) {
	src := [64]uint8{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16,
		17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32,
		33, 34, 35, 36, 37, 38, 39, 40, 41, 42, 43, 44, 45, 46, 47, 48,
		49, 50, 51, 52, 53, 54, 55, 56, 57, 58, 59, 60, 61, 62, 63, 64}
	dst := [64]uint8{}
	const maskBits uint64 = 0xAAAAAAAAAAAAAAAA

	b.ResetTimer()
	for b.Loop() {
		j := 0
		for i := range 64 {
			if (maskBits>>i)&1 == 1 {
				dst[j] = src[i]
				j++
			}
		}
	}
	_ = dst
}

// BenchmarkScalarUint8x64Expand is the scalar baseline for BenchmarkSIMDUint8x64Expand.
// Scatters packed source lanes into positions indicated by the mask, zeroing the rest.
// Uses the same alternating mask 0xAAAAAAAAAAAAAAAA as the SIMD benchmark.
func BenchmarkScalarUint8x64Expand(b *testing.B) {
	src := [64]uint8{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16,
		17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32,
		33, 34, 35, 36, 37, 38, 39, 40, 41, 42, 43, 44, 45, 46, 47, 48,
		49, 50, 51, 52, 53, 54, 55, 56, 57, 58, 59, 60, 61, 62, 63, 64}
	dst := [64]uint8{}
	const maskBits uint64 = 0xAAAAAAAAAAAAAAAA

	b.ResetTimer()
	for b.Loop() {
		j := 0
		for i := range 64 {
			if (maskBits>>i)&1 == 1 {
				dst[i] = src[j]
				j++
			} else {
				dst[i] = 0
			}
		}
	}
	_ = dst
}
