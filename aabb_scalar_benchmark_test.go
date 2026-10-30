package mybenchmark

import "testing"

// Vec3 is a 3D float32 vector.
type Vec3 struct{ X, Y, Z float32 }

// AABB is an axis-aligned bounding box used for collision detection.
type AABB struct{ Min, Max Vec3 }

// Input constants shared with aabb_simd_benchmark_test.go:
//   16 boxes: box[i].Min=(i+1,i+1,i+1)  box[i].Max=(i+6,i+6,i+6)  i=0..15
//   Query box: Min=(3,3,3) Max=(10,10,10)
//   Query point: (5,5,5)
//   Expand delta: 1.0

func makeScalarBoxes() [16]AABB {
	var boxes [16]AABB
	for i := range boxes {
		f := float32(i + 1)
		boxes[i] = AABB{Min: Vec3{f, f, f}, Max: Vec3{f + 5, f + 5, f + 5}}
	}
	return boxes
}

func scalarAABBIntersects(a, b AABB) bool {
	return a.Min.X <= b.Max.X && a.Max.X >= b.Min.X &&
		a.Min.Y <= b.Max.Y && a.Max.Y >= b.Min.Y &&
		a.Min.Z <= b.Max.Z && a.Max.Z >= b.Min.Z
}

func scalarAABBContainsPoint(a AABB, p Vec3) bool {
	return p.X >= a.Min.X && p.X <= a.Max.X &&
		p.Y >= a.Min.Y && p.Y <= a.Max.Y &&
		p.Z >= a.Min.Z && p.Z <= a.Max.Z
}

// BenchmarkScalarAABBIntersect16 loops over 16 AABBs and tests each against a
// query box, accumulating a uint16 bitmask (1 bit per box).
// SIMD counterpart: BenchmarkSIMDAABBIntersect16.
func BenchmarkScalarAABBIntersect16(b *testing.B) {
	boxes := makeScalarBoxes()
	query := AABB{Min: Vec3{3, 3, 3}, Max: Vec3{10, 10, 10}}
	var dst uint16
	for b.Loop() {
		var mask uint16
		for i := range boxes {
			if scalarAABBIntersects(boxes[i], query) {
				mask |= 1 << i
			}
		}
		dst = mask
	}
	_ = dst
}

// BenchmarkScalarAABBContainsPoint16 tests which of 16 AABBs contain point (5,5,5).
// SIMD counterpart: BenchmarkSIMDAABBContainsPoint16.
func BenchmarkScalarAABBContainsPoint16(b *testing.B) {
	boxes := makeScalarBoxes()
	p := Vec3{5, 5, 5}
	var dst uint16
	for b.Loop() {
		var mask uint16
		for i := range boxes {
			if scalarAABBContainsPoint(boxes[i], p) {
				mask |= 1 << i
			}
		}
		dst = mask
	}
	_ = dst
}

// BenchmarkScalarAABBExpand16 expands 16 AABBs outward by delta=1.0 on all sides.
// SIMD counterpart: BenchmarkSIMDAABBExpand16.
func BenchmarkScalarAABBExpand16(b *testing.B) {
	boxes := makeScalarBoxes()
	const delta float32 = 1.0
	var dst [16]AABB
	for b.Loop() {
		for i := range boxes {
			dst[i] = AABB{
				Min: Vec3{
					boxes[i].Min.X - delta,
					boxes[i].Min.Y - delta,
					boxes[i].Min.Z - delta,
				},
				Max: Vec3{
					boxes[i].Max.X + delta,
					boxes[i].Max.Y + delta,
					boxes[i].Max.Z + delta,
				},
			}
		}
	}
	_ = dst
}
