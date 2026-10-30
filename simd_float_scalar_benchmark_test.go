package mybenchmark

import (
	"math"
	"testing"
)

// BenchmarkScalarFloat32x16Add is the scalar baseline for BenchmarkSIMDFloat32x16Add.
// Performs element-wise float32 addition over 16 lanes using a plain loop.
func BenchmarkScalarFloat32x16Add(b *testing.B) {
	src1 := [16]float32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	src2 := [16]float32{16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}
	dst := [16]float32{}

	b.ResetTimer()
	for b.Loop() {
		for i := range 16 {
			dst[i] = src1[i] + src2[i]
		}
	}
	_ = dst
}

// BenchmarkScalarFloat32x16MulAdd is the scalar baseline for BenchmarkSIMDFloat32x16MulAdd.
// Computes (x*y)+z over 16 float32 lanes using a plain loop.
func BenchmarkScalarFloat32x16MulAdd(b *testing.B) {
	srcX := [16]float32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	srcY := [16]float32{2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2}
	srcZ := [16]float32{0.5, 1, 1.5, 2, 2.5, 3, 3.5, 4, 4.5, 5, 5.5, 6, 6.5, 7, 7.5, 8}
	dst := [16]float32{}

	b.ResetTimer()
	for b.Loop() {
		for i := range 16 {
			dst[i] = srcX[i]*srcY[i] + srcZ[i]
		}
	}
	_ = dst
}

// BenchmarkScalarFloat64x8Sqrt is the scalar baseline for BenchmarkSIMDFloat64x8Sqrt.
// Computes math.Sqrt over 8 float64 lanes using a plain loop.
func BenchmarkScalarFloat64x8Sqrt(b *testing.B) {
	src := [8]float64{1, 4, 9, 16, 25, 36, 49, 64}
	dst := [8]float64{}

	b.ResetTimer()
	for b.Loop() {
		for i := range 8 {
			dst[i] = math.Sqrt(src[i])
		}
	}
	_ = dst
}

// BenchmarkScalarFloat32x16Reciprocal is the scalar baseline for BenchmarkSIMDFloat32x16Reciprocal.
// Computes 1/x over 16 float32 lanes using a plain loop.
func BenchmarkScalarFloat32x16Reciprocal(b *testing.B) {
	src := [16]float32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	dst := [16]float32{}

	b.ResetTimer()
	for b.Loop() {
		for i := range 16 {
			dst[i] = 1.0 / src[i]
		}
	}
	_ = dst
}
