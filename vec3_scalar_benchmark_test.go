package mybenchmark

import (
	"math"
	"testing"
)

// Vec3 is already defined in aabb_scalar_benchmark_test.go.

var (
	scalarA = Vec3{1, 2, 3}
	scalarB = Vec3{4, 5, 6}
)

// Vec3Add adds two Vec3s component-wise: 3 float32 add instructions.
// SIMD counterpart: Vec3F4.Add — 1 VADDPS instruction.
func Vec3Add(a, b Vec3) Vec3 {
	return Vec3{a.X + b.X, a.Y + b.Y, a.Z + b.Z}
}

// Vec3Scale multiplies each component of v by s: 3 float32 mul instructions.
// SIMD counterpart: Vec3F4.Scale — 1 VMULPS instruction.
func Vec3Scale(v Vec3, s float32) Vec3 {
	return Vec3{v.X * s, v.Y * s, v.Z * s}
}

// Vec3Lerp interpolates from a to b at t: 3 sub + 3 fused-mul-add instructions.
// SIMD counterpart: Vec3F4.Lerp — 1 VSUBPS + 1 VFMADD213PS instruction.
func Vec3Lerp(a, b Vec3, t float32) Vec3 {
	return Vec3{
		a.X + t*(b.X-a.X),
		a.Y + t*(b.Y-a.Y),
		a.Z + t*(b.Z-a.Z),
	}
}

// BenchmarkScalarVec3Add calls Vec3Add once per iteration: 3 float32 adds.
// SIMD counterpart: BenchmarkSIMDVec3Add.
func BenchmarkScalarVec3Add(b *testing.B) {
	var dst Vec3
	for b.Loop() {
		dst = Vec3Add(scalarA, scalarB)
	}
	_ = dst
}

// BenchmarkScalarVec3Scale calls Vec3Scale once per iteration: 3 float32 muls.
// SIMD counterpart: BenchmarkSIMDVec3Scale.
func BenchmarkScalarVec3Scale(b *testing.B) {
	var dst Vec3
	for b.Loop() {
		dst = Vec3Scale(scalarA, 2.5)
	}
	_ = dst
}

// BenchmarkScalarVec3Lerp calls Vec3Lerp once per iteration: 3 subs + 3 fused-mul-adds.
// SIMD counterpart: BenchmarkSIMDVec3Lerp.
func BenchmarkScalarVec3Lerp(b *testing.B) {
	var dst Vec3
	for b.Loop() {
		dst = Vec3Lerp(scalarA, scalarB, 0.3)
	}
	_ = dst
}

// Vec3Dot computes the dot product a·b: 3 float32 muls + 2 adds.
// SIMD counterpart: Vec3F4.Dot — 1 VMULPS + horizontal sum of 3 lanes.
func Vec3Dot(a, b Vec3) float32 {
	return a.X*b.X + a.Y*b.Y + a.Z*b.Z
}

// BenchmarkScalarVec3Dot calls Vec3Dot once per iteration: 3 muls + 2 adds.
// SIMD counterpart: BenchmarkSIMDVec3Dot.
func BenchmarkScalarVec3Dot(b *testing.B) {
	var dst float32
	for b.Loop() {
		dst = Vec3Dot(scalarA, scalarB)
	}
	_ = dst
}

// Vec3Sub subtracts b from a component-wise: 3 float32 subs.
// SIMD counterpart: Vec3F4.Sub — 1 VSUBPS instruction.
func Vec3Sub(a, b Vec3) Vec3 {
	return Vec3{a.X - b.X, a.Y - b.Y, a.Z - b.Z}
}

func BenchmarkScalarVec3Sub(b *testing.B) {
	var dst Vec3
	for b.Loop() {
		dst = Vec3Sub(scalarA, scalarB)
	}
	_ = dst
}

// Vec3Div divides each component of v by s: 3 float32 divs.
// SIMD counterpart: Vec3F4.Div — 1 VMULPS (with 1/s broadcast).
func Vec3Div(v Vec3, s float32) Vec3 {
	return Vec3{v.X / s, v.Y / s, v.Z / s}
}

func BenchmarkScalarVec3Div(b *testing.B) {
	var dst Vec3
	for b.Loop() {
		dst = Vec3Div(scalarA, 2.0)
	}
	_ = dst
}

// Vec3Equal reports whether a and b are approximately equal (squared distance < 1e-10).
func Vec3Equal(a, b Vec3) bool {
	d := Vec3Sub(a, b)
	return Vec3Dot(d, d) < 1e-10
}

func BenchmarkScalarVec3Equal(b *testing.B) {
	var dst bool
	for b.Loop() {
		dst = Vec3Equal(scalarA, scalarB)
	}
	_ = dst
}

// Vec3NotEqual reports whether a and b are not approximately equal.
func Vec3NotEqual(a, b Vec3) bool {
	return !Vec3Equal(a, b)
}

func BenchmarkScalarVec3NotEqual(b *testing.B) {
	var dst bool
	for b.Loop() {
		dst = Vec3NotEqual(scalarA, scalarB)
	}
	_ = dst
}

// Vec3Magnitude returns the length of v: sqrt(x²+y²+z²).
// SIMD counterpart: Vec3F4.Magnitude.
func Vec3Magnitude(v Vec3) float32 {
	return float32(math.Sqrt(float64(Vec3Dot(v, v))))
}

func BenchmarkScalarVec3Magnitude(b *testing.B) {
	var dst float32
	for b.Loop() {
		dst = Vec3Magnitude(scalarA)
	}
	_ = dst
}

// Vec3Cross returns the cross product a×b: 6 muls + 3 subs.
// SIMD counterpart: Vec3F4.Cross (Store→scalar→Load; no Permute available).
func Vec3Cross(a, b Vec3) Vec3 {
	return Vec3{
		a.Y*b.Z - a.Z*b.Y,
		a.Z*b.X - a.X*b.Z,
		a.X*b.Y - a.Y*b.X,
	}
}

func BenchmarkScalarVec3Cross(b *testing.B) {
	var dst Vec3
	for b.Loop() {
		dst = Vec3Cross(scalarA, scalarB)
	}
	_ = dst
}

// Vec3Distance returns the Euclidean distance between points a and b.
// SIMD counterpart: Vec3F4.Distance.
func Vec3Distance(a, b Vec3) float32 {
	d := Vec3Sub(a, b)
	return Vec3Magnitude(d)
}

func BenchmarkScalarVec3Distance(b *testing.B) {
	var dst float32
	for b.Loop() {
		dst = Vec3Distance(scalarA, scalarB)
	}
	_ = dst
}

// Vec3Angle returns the angle in radians between vectors a and b.
// SIMD counterpart: Vec3F4.Angle.
func Vec3Angle(a, b Vec3) float32 {
	dot := Vec3Dot(a, b)
	mag := Vec3Magnitude(a) * Vec3Magnitude(b)
	return float32(math.Acos(float64(dot / mag)))
}

func BenchmarkScalarVec3Angle(b *testing.B) {
	var dst float32
	for b.Loop() {
		dst = Vec3Angle(scalarA, scalarB)
	}
	_ = dst
}

// Vec3ClampMagnitude returns v scaled so its magnitude does not exceed maxLen.
// SIMD counterpart: Vec3F4.ClampMagnitude.
func Vec3ClampMagnitude(v Vec3, maxLen float32) Vec3 {
	if mag := Vec3Magnitude(v); mag > maxLen {
		return Vec3Scale(v, maxLen/mag)
	}
	return v
}

func BenchmarkScalarVec3ClampMagnitude(b *testing.B) {
	var dst Vec3
	for b.Loop() {
		dst = Vec3ClampMagnitude(scalarA, 5.0)
	}
	_ = dst
}
