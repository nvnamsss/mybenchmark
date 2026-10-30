//go:build goexperiment.simd

package mybenchmark

import (
	"simd/archsimd"
	"testing"
)

// BenchmarkSIMDFloat32x16Add benchmarks 512-bit element-wise float32 addition over 16 lanes.
// Scalar baseline: BenchmarkScalarFloat32x16Add.
func BenchmarkSIMDFloat32x16Add(b *testing.B) {
	src1 := [16]float32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	src2 := [16]float32{16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}
	dst := [16]float32{}

	a := archsimd.LoadFloat32x16(&src1)
	bv := archsimd.LoadFloat32x16(&src2)
	b.ResetTimer()
	for b.Loop() {
		result := a.Add(bv)
		result.Store(&dst)
	}
}

// BenchmarkSIMDFloat32x16MulAdd benchmarks 512-bit fused multiply-add (FMA) over 16 float32 lanes.
// Computes (x*y)+z in a single instruction.
// Scalar baseline: BenchmarkScalarFloat32x16MulAdd.
func BenchmarkSIMDFloat32x16MulAdd(b *testing.B) {
	srcX := [16]float32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	srcY := [16]float32{2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2}
	srcZ := [16]float32{0.5, 1, 1.5, 2, 2.5, 3, 3.5, 4, 4.5, 5, 5.5, 6, 6.5, 7, 7.5, 8}
	dst := [16]float32{}

	x := archsimd.LoadFloat32x16(&srcX)
	y := archsimd.LoadFloat32x16(&srcY)
	z := archsimd.LoadFloat32x16(&srcZ)
	b.ResetTimer()
	for b.Loop() {
		result := x.MulAdd(y, z)
		result.Store(&dst)
	}
}

// BenchmarkSIMDFloat64x8Sqrt benchmarks 512-bit square root over 8 float64 lanes.
// Scalar baseline: BenchmarkScalarFloat64x8Sqrt.
func BenchmarkSIMDFloat64x8Sqrt(b *testing.B) {
	src := [8]float64{1, 4, 9, 16, 25, 36, 49, 64}
	dst := [8]float64{}

	a := archsimd.LoadFloat64x8(&src)
	b.ResetTimer()
	for b.Loop() {
		result := a.Sqrt()
		result.Store(&dst)
	}
}

// BenchmarkSIMDFloat32x16Reciprocal benchmarks 512-bit reciprocal (1/x) over 16 float32 lanes.
// Scalar baseline: BenchmarkScalarFloat32x16Reciprocal.
func BenchmarkSIMDFloat32x16Reciprocal(b *testing.B) {
	src := [16]float32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	dst := [16]float32{}

	a := archsimd.LoadFloat32x16(&src)
	b.ResetTimer()
	for b.Loop() {
		result := a.Reciprocal()
		result.Store(&dst)
	}
}
