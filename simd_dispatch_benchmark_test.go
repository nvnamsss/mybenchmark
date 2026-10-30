//go:build goexperiment.simd

package mybenchmark

import (
	"fmt"
	"simd/archsimd"
	"testing"
)

// simdFloat32x16Add is the SIMD path for a 16-lane float32 addition.
// Callers must have verified archsimd.X86.AVX512() before calling.
func simdFloat32x16Add(src1, src2 *[16]float32, dst *[16]float32) {
	archsimd.LoadFloat32x16(src1).Add(archsimd.LoadFloat32x16(src2)).Store(dst)
}

// scalarFloat32x16Add is the scalar fallback path for a 16-lane float32 addition,
// used on machines where AVX512 is not available.
func scalarFloat32x16Add(src1, src2 *[16]float32, dst *[16]float32) {
	for i := range 16 {
		dst[i] = src1[i] + src2[i]
	}
}

// dispatchFloat32x16Add demonstrates the runtime CPU dispatch pattern:
// selects the SIMD or scalar path based on AVX512 availability at call time.
func dispatchFloat32x16Add(src1, src2 *[16]float32, dst *[16]float32) {
	if archsimd.X86.AVX512() {
		simdFloat32x16Add(src1, src2, dst)
	} else {
		scalarFloat32x16Add(src1, src2, dst)
	}
}

// TestSIMDCPUFeatures prints the CPU feature flags relevant to this benchmark suite.
// Run with: GOEXPERIMENT=simd go test -run=TestSIMDCPUFeatures -v ./...
func TestSIMDCPUFeatures(t *testing.T) {
	features := []struct {
		name    string
		present bool
	}{
		{"AVX2", archsimd.X86.AVX2()},
		{"AVX512", archsimd.X86.AVX512()},
		{"FMA", archsimd.X86.FMA()},
		{"AVXAES", archsimd.X86.AVXAES()},
		{"AVX512GFNI (GaloisFieldMul)", archsimd.X86.AVX512GFNI()},
		{"AVX512VBMI2 (Compress/Expand)", archsimd.X86.AVX512VBMI2()},
		{"AVX512VNNI", archsimd.X86.AVX512VNNI()},
		{"AVX512VPOPCNTDQ", archsimd.X86.AVX512VPOPCNTDQ()},
		{"AVX512BITALG", archsimd.X86.AVX512BITALG()},
	}
	for _, f := range features {
		status := "YES"
		if !f.present {
			status = "no"
		}
		fmt.Printf("  %-36s %s\n", f.name, status)
	}
	if !archsimd.X86.AVX512() {
		t.Skip("AVX512 not available — SIMD benchmarks will be skipped on this machine")
	}
}

// BenchmarkSIMDDispatchFloat32x16Add benchmarks the SIMD path of the CPU dispatcher.
// Skips if AVX512 is not available on the current machine.
// Scalar side-by-side: BenchmarkScalarFallbackFloat32x16Add.
func BenchmarkSIMDDispatchFloat32x16Add(b *testing.B) {
	if !archsimd.X86.AVX512() {
		b.Skip("AVX512 not available")
	}
	src1 := [16]float32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	src2 := [16]float32{16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}
	dst := [16]float32{}

	b.ResetTimer()
	for b.Loop() {
		simdFloat32x16Add(&src1, &src2, &dst)
	}
}

// BenchmarkScalarFallbackFloat32x16Add benchmarks the scalar fallback path that
// would execute on machines without AVX512. Runs unconditionally.
// SIMD side-by-side: BenchmarkSIMDDispatchFloat32x16Add.
func BenchmarkScalarFallbackFloat32x16Add(b *testing.B) {
	src1 := [16]float32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	src2 := [16]float32{16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}
	dst := [16]float32{}

	b.ResetTimer()
	for b.Loop() {
		scalarFloat32x16Add(&src1, &src2, &dst)
	}
}

// BenchmarkDispatchFloat32x16Add benchmarks the full dispatcher including the
// runtime AVX512 branch, showing the overhead of the CPU feature check itself.
func BenchmarkDispatchFloat32x16Add(b *testing.B) {
	src1 := [16]float32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	src2 := [16]float32{16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}
	dst := [16]float32{}

	b.ResetTimer()
	for b.Loop() {
		dispatchFloat32x16Add(&src1, &src2, &dst)
	}
}
