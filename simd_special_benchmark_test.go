//go:build goexperiment.simd

package mybenchmark

import (
	"simd/archsimd"
	"testing"
)

// BenchmarkSIMDUint8x16AESEncryptOneRound benchmarks a single AES encryption round
// using the VAESENC instruction via AESEncryptOneRound.
// The round key is supplied as Uint32x4 (one 128-bit chunk of the key schedule).
// Scalar baseline: BenchmarkScalarUint8x16AESEncryptOneRound (full 10-round crypto/aes).
func BenchmarkSIMDUint8x16AESEncryptOneRound(b *testing.B) {
	// AES-128 example: FIPS-197 test vector state and first round key
	state := [16]uint8{0x32, 0x43, 0xf6, 0xa8, 0x88, 0x5a, 0x30, 0x8d,
		0x31, 0x31, 0x98, 0xa2, 0xe0, 0x37, 0x07, 0x34}
	roundKey := [4]uint32{0x2b7e1516, 0x28aed2a6, 0xabf71588, 0x09cf4f3c}
	dst := [16]uint8{}

	s := archsimd.LoadUint8x16(&state)
	k := archsimd.LoadUint32x4(&roundKey)
	b.ResetTimer()
	for b.Loop() {
		result := s.AESEncryptOneRound(k)
		result.Store(&dst)
	}
}

// BenchmarkSIMDUint8x16GaloisFieldMul benchmarks element-wise GF(2^8) multiplication
// over 16 uint8 lanes using the VGF2P8MULB instruction (AVX512GFNI).
// Scalar baseline: BenchmarkScalarUint8x16GaloisFieldMul.
func BenchmarkSIMDUint8x16GaloisFieldMul(b *testing.B) {
	src1 := [16]uint8{0x53, 0xCA, 0x01, 0xFF, 0x1B, 0x47, 0xE3, 0x9A,
		0x57, 0xAB, 0x2D, 0x7F, 0xC3, 0x84, 0x16, 0x5E}
	src2 := [16]uint8{0x02, 0x03, 0x02, 0x03, 0x02, 0x03, 0x02, 0x03,
		0x02, 0x03, 0x02, 0x03, 0x02, 0x03, 0x02, 0x03}
	dst := [16]uint8{}

	a := archsimd.LoadUint8x16(&src1)
	bv := archsimd.LoadUint8x16(&src2)
	b.ResetTimer()
	for b.Loop() {
		result := a.GaloisFieldMul(bv)
		result.Store(&dst)
	}
}

// BenchmarkSIMDInt32x16LeadingZeros benchmarks VPLZCNTD: leading zero count
// over 16 int32 lanes simultaneously.
// Scalar baseline: BenchmarkScalarInt32x16LeadingZeros.
func BenchmarkSIMDInt32x16LeadingZeros(b *testing.B) {
	src := [16]int32{0x0F0F0F0F, 0x12345678, 1, -1,
		0x01010101, -16711936, 0x55555555, 0x00FF00FF,
		0x0000FFFF, -65536, 0x11111111, 0x22222222,
		0x44444444, 7, 255, 1 << 30}
	dst := [16]int32{}

	a := archsimd.LoadInt32x16(&src)
	b.ResetTimer()
	for b.Loop() {
		result := a.LeadingZeros()
		result.Store(&dst)
	}
}

// BenchmarkSIMDUint32x16RotateAllLeft benchmarks VPROLVD with a uniform 8-bit
// left rotation over 16 uint32 lanes simultaneously.
// Scalar baseline: BenchmarkScalarUint32x16RotateAllLeft.
func BenchmarkSIMDUint32x16RotateAllLeft(b *testing.B) {
	src := [16]uint32{0xDEADBEEF, 0xCAFEBABE, 0x01234567, 0x89ABCDEF,
		0x11111111, 0x22222222, 0x44444444, 0x88888888,
		0x0F0F0F0F, 0xF0F0F0F0, 0x00FF00FF, 0xFF00FF00,
		0x55555555, 0xAAAAAAAA, 0x0000FFFF, 0xFFFF0000}
	dst := [16]uint32{}

	a := archsimd.LoadUint32x16(&src)
	b.ResetTimer()
	for b.Loop() {
		result := a.RotateAllLeft(8)
		result.Store(&dst)
	}
}
