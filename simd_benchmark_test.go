//go:build goexperiment.simd

package mybenchmark

import (
	"simd/archsimd"
	"testing"
)

// BenchmarkSIMDSmoke is a minimal smoke benchmark that verifies the simd/archsimd
// package wires up correctly. It loads a [16]uint8 array into a Uint8x16 vector,
// performs an element-wise Add, and stores the result back.
//
// Run with: GOEXPERIMENT=simd go test -bench=BenchmarkSIMDSmoke
func BenchmarkSIMDSmoke(b *testing.B) {
	src := [16]uint8{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	dst := [16]uint8{}

	a := archsimd.LoadUint8x16(&src)
	one := archsimd.BroadcastUint8x16(1)

	b.ResetTimer()
	for b.Loop() {
		result := a.Add(one)
		result.Store(&dst)
	}
}
