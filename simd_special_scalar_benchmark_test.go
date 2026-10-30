package mybenchmark

import (
	"crypto/aes"
	"math/bits"
	"testing"
)

// gfMulByte multiplies two elements in GF(2^8) with the AES reduction polynomial
// x^8 + x^4 + x^3 + x + 1 (0x1B). Used by BenchmarkScalarUint8x16GaloisFieldMul.
func gfMulByte(a, b uint8) uint8 {
	var p uint8
	for range 8 {
		if b&1 != 0 {
			p ^= a
		}
		hi := a & 0x80
		a <<= 1
		if hi != 0 {
			a ^= 0x1B
		}
		b >>= 1
	}
	return p
}

// BenchmarkScalarUint8x16AESEncryptOneRound is the scalar baseline for
// BenchmarkSIMDUint8x16AESEncryptOneRound.
// NOTE: crypto/aes performs a full 10-round AES-128 encrypt, whereas the SIMD
// benchmark executes a single round — the comparison shows the hardware primitive
// cost vs the complete Go stdlib path.
func BenchmarkScalarUint8x16AESEncryptOneRound(b *testing.B) {
	key := [16]byte{0x2b, 0x7e, 0x15, 0x16, 0x28, 0xae, 0xd2, 0xa6,
		0xab, 0xf7, 0x15, 0x88, 0x09, 0xcf, 0x4f, 0x3c}
	plaintext := [16]byte{0x32, 0x43, 0xf6, 0xa8, 0x88, 0x5a, 0x30, 0x8d,
		0x31, 0x31, 0x98, 0xa2, 0xe0, 0x37, 0x07, 0x34}
	dst := [16]byte{}

	c, _ := aes.NewCipher(key[:])
	b.ResetTimer()
	for b.Loop() {
		c.Encrypt(dst[:], plaintext[:])
	}
	_ = dst
}

// BenchmarkScalarUint8x16GaloisFieldMul is the scalar baseline for
// BenchmarkSIMDUint8x16GaloisFieldMul.
// Performs element-wise GF(2^8) multiplication over 16 uint8 lanes using a
// software XOR-shift loop with the AES reduction polynomial.
func BenchmarkScalarUint8x16GaloisFieldMul(b *testing.B) {
	src1 := [16]uint8{0x53, 0xCA, 0x01, 0xFF, 0x1B, 0x47, 0xE3, 0x9A,
		0x57, 0xAB, 0x2D, 0x7F, 0xC3, 0x84, 0x16, 0x5E}
	src2 := [16]uint8{0x02, 0x03, 0x02, 0x03, 0x02, 0x03, 0x02, 0x03,
		0x02, 0x03, 0x02, 0x03, 0x02, 0x03, 0x02, 0x03}
	dst := [16]uint8{}

	b.ResetTimer()
	for b.Loop() {
		for i := range 16 {
			dst[i] = gfMulByte(src1[i], src2[i])
		}
	}
	_ = dst
}

// BenchmarkScalarInt32x16LeadingZeros is the scalar baseline for
// BenchmarkSIMDInt32x16LeadingZeros.
// Uses math/bits.LeadingZeros32 over 16 int32 lanes.
func BenchmarkScalarInt32x16LeadingZeros(b *testing.B) {
	src := [16]int32{0x0F0F0F0F, 0x12345678, 1, -1,
		0x01010101, -16711936, 0x55555555, 0x00FF00FF,
		0x0000FFFF, -65536, 0x11111111, 0x22222222,
		0x44444444, 7, 255, 1 << 30}
	dst := [16]int32{}

	b.ResetTimer()
	for b.Loop() {
		for i := range 16 {
			dst[i] = int32(bits.LeadingZeros32(uint32(src[i])))
		}
	}
	_ = dst
}

// BenchmarkScalarUint32x16RotateAllLeft is the scalar baseline for
// BenchmarkSIMDUint32x16RotateAllLeft.
// Uses math/bits.RotateLeft32 with a uniform 8-bit rotation over 16 uint32 lanes.
func BenchmarkScalarUint32x16RotateAllLeft(b *testing.B) {
	src := [16]uint32{0xDEADBEEF, 0xCAFEBABE, 0x01234567, 0x89ABCDEF,
		0x11111111, 0x22222222, 0x44444444, 0x88888888,
		0x0F0F0F0F, 0xF0F0F0F0, 0x00FF00FF, 0xFF00FF00,
		0x55555555, 0xAAAAAAAA, 0x0000FFFF, 0xFFFF0000}
	dst := [16]uint32{}

	b.ResetTimer()
	for b.Loop() {
		for i := range 16 {
			dst[i] = bits.RotateLeft32(src[i], 8)
		}
	}
	_ = dst
}
