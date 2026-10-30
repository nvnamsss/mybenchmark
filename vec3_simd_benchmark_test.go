//go:build goexperiment.simd

package mybenchmark

import (
	"math"
	"simd/archsimd"
	"testing"
)

// Vec3F4 is a SIMD-backed 3D vector using a single Float32x4 register.
// Layout: lane0=X, lane1=Y, lane2=Z, lane3=0 (unused).
// One 128-bit VADDPS/VMULPS/VFMADD213PS instruction handles all 3 components at once.
type Vec3F4 struct {
	v archsimd.Float32x4
}

// Add returns a + b using 1 VADDPS instruction, replacing 3 scalar float32 adds.
func (a Vec3F4) Add(b Vec3F4) Vec3F4 {
	return Vec3F4{v: a.v.Add(b.v)}
}

// Scale returns v * s using 1 VMULPS instruction, replacing 3 scalar float32 muls.
func (a Vec3F4) Scale(s float32) Vec3F4 {
	return Vec3F4{v: a.v.Mul(archsimd.BroadcastFloat32x4(s))}
}

// Lerp returns a + t*(b-a) using VSUBPS + VFMADD213PS, replacing 6 scalar float32 ops.
func (a Vec3F4) Lerp(b Vec3F4, t float32) Vec3F4 {
	tv := archsimd.BroadcastFloat32x4(t)
	return Vec3F4{v: b.v.Sub(a.v).MulAdd(tv, a.v)}
}

// newVec3F4 constructs a Vec3F4 from x, y, z scalars.
func newVec3F4(x, y, z float32) Vec3F4 {
	arr := [4]float32{x, y, z, 0}
	return Vec3F4{v: archsimd.LoadFloat32x4(&arr)}
}

var (
	simdA = newVec3F4(1, 2, 3)
	simdB = newVec3F4(4, 5, 6)
)

// BenchmarkSIMDVec3Add calls Vec3F4.Add once per iteration: 1 VADDPS replaces 3 scalar adds.
// Scalar baseline: BenchmarkScalarVec3Add.
func BenchmarkSIMDVec3Add(b *testing.B) {
	var dst Vec3F4
	for b.Loop() {
		dst = simdA.Add(simdB)
	}
	_ = dst
}

// BenchmarkSIMDVec3Scale calls Vec3F4.Scale once per iteration: 1 VMULPS replaces 3 scalar muls.
// Scalar baseline: BenchmarkScalarVec3Scale.
func BenchmarkSIMDVec3Scale(b *testing.B) {
	var dst Vec3F4
	for b.Loop() {
		dst = simdA.Scale(2.5)
	}
	_ = dst
}

// BenchmarkSIMDVec3Lerp calls Vec3F4.Lerp once per iteration: VSUBPS + VFMADD213PS replaces 6 scalar ops.
// Scalar baseline: BenchmarkScalarVec3Lerp.
func BenchmarkSIMDVec3Lerp(b *testing.B) {
	var dst Vec3F4
	for b.Loop() {
		dst = simdA.Lerp(simdB, 0.3)
	}
	_ = dst
}

// Dot returns the dot product a·b = X*X + Y*Y + Z*Z as a float32.
// Multiplies all 3 components simultaneously with 1 VMULPS, then sums lanes 0–2.
func (a Vec3F4) Dot(b Vec3F4) float32 {
	var arr [4]float32
	a.v.Mul(b.v).Store(&arr)
	return arr[0] + arr[1] + arr[2]
}

// BenchmarkSIMDVec3Dot calls Vec3F4.Dot once per iteration: 1 VMULPS + 2 adds.
// Scalar baseline: BenchmarkScalarVec3Dot.
func BenchmarkSIMDVec3Dot(b *testing.B) {
	var dst float32
	for b.Loop() {
		dst = simdA.Dot(simdB)
	}
	_ = dst
}

// Sub returns a - b using 1 VSUBPS instruction, replacing 3 scalar float32 subs.
func (a Vec3F4) Sub(b Vec3F4) Vec3F4 {
	return Vec3F4{v: a.v.Sub(b.v)}
}

func BenchmarkSIMDVec3Sub(b *testing.B) {
	var dst Vec3F4
	for b.Loop() {
		dst = simdA.Sub(simdB)
	}
	_ = dst
}

// Div returns v / s using 1 VMULPS instruction (broadcasts 1/s), replacing 3 scalar divs.
func (a Vec3F4) Div(s float32) Vec3F4 {
	return Vec3F4{v: a.v.Mul(archsimd.BroadcastFloat32x4(1.0 / s))}
}

func BenchmarkSIMDVec3Div(b *testing.B) {
	var dst Vec3F4
	for b.Loop() {
		dst = simdA.Div(2.0)
	}
	_ = dst
}

// Equal reports whether a and b are approximately equal (squared distance < 1e-10).
func (a Vec3F4) Equal(b Vec3F4) bool {
	d := a.Sub(b)
	return d.Dot(d) < 1e-10
}

func BenchmarkSIMDVec3Equal(b *testing.B) {
	var dst bool
	for b.Loop() {
		dst = simdA.Equal(simdB)
	}
	_ = dst
}

// NotEqual reports whether a and b are not approximately equal.
func (a Vec3F4) NotEqual(b Vec3F4) bool {
	return !a.Equal(b)
}

func BenchmarkSIMDVec3NotEqual(b *testing.B) {
	var dst bool
	for b.Loop() {
		dst = simdA.NotEqual(simdB)
	}
	_ = dst
}

// Magnitude returns the length of v: sqrt(x²+y²+z²).
// Uses Dot (1 VMULPS + Store) then scalar sqrt.
func (v Vec3F4) Magnitude() float32 {
	return float32(math.Sqrt(float64(v.Dot(v))))
}

func BenchmarkSIMDVec3Magnitude(b *testing.B) {
	var dst float32
	for b.Loop() {
		dst = simdA.Magnitude()
	}
	_ = dst
}

// Cross returns the cross product a×b.
// Float32x4.Permute is unavailable, so both vectors are stored to arrays,
// the cross product is computed with scalar arithmetic, then reloaded.
func (a Vec3F4) Cross(b Vec3F4) Vec3F4 {
	var aa, bb [4]float32
	a.v.Store(&aa)
	b.v.Store(&bb)
	return newVec3F4(
		aa[1]*bb[2]-aa[2]*bb[1],
		aa[2]*bb[0]-aa[0]*bb[2],
		aa[0]*bb[1]-aa[1]*bb[0],
	)
}

func BenchmarkSIMDVec3Cross(b *testing.B) {
	var dst Vec3F4
	for b.Loop() {
		dst = simdA.Cross(simdB)
	}
	_ = dst
}

// Distance returns the Euclidean distance between a and b.
// Sub uses 1 VSUBPS; the rest is Dot (VMULPS + Store) + scalar sqrt.
func (a Vec3F4) Distance(b Vec3F4) float32 {
	d := a.Sub(b)
	return float32(math.Sqrt(float64(d.Dot(d))))
}

func BenchmarkSIMDVec3Distance(b *testing.B) {
	var dst float32
	for b.Loop() {
		dst = simdA.Distance(simdB)
	}
	_ = dst
}

// Angle returns the angle in radians between vectors a and b.
func (a Vec3F4) Angle(b Vec3F4) float32 {
	dot := a.Dot(b)
	mag := a.Magnitude() * b.Magnitude()
	return float32(math.Acos(float64(dot / mag)))
}

func BenchmarkSIMDVec3Angle(b *testing.B) {
	var dst float32
	for b.Loop() {
		dst = simdA.Angle(simdB)
	}
	_ = dst
}

// ClampMagnitude returns v scaled so its magnitude does not exceed maxLen.
// If clamping occurs, Scale uses 1 VMULPS.
func (v Vec3F4) ClampMagnitude(maxLen float32) Vec3F4 {
	if mag := v.Magnitude(); mag > maxLen {
		return v.Scale(maxLen / mag)
	}
	return v
}

func BenchmarkSIMDVec3ClampMagnitude(b *testing.B) {
	var dst Vec3F4
	for b.Loop() {
		dst = simdA.ClampMagnitude(5.0)
	}
	_ = dst
}
