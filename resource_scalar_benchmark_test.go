package mybenchmark

import "testing"

// Scalar SoA layout for 100,000 units — 9 separate [100000]float32 arrays.
// Total size: 9 × 400,000 bytes = 3.6 MB (fits in L3 cache).
var (
	resHp      [100000]float32
	resHpRegen [100000]float32
	resMaxHp   [100000]float32

	resMp      [100000]float32
	resMpRegen [100000]float32
	resMaxMp   [100000]float32

	resSp      [100000]float32
	resSpRegen [100000]float32
	resMaxSp   [100000]float32
)

func init() {
	for i := range resHp {
		resHp[i] = 50 + float32(i%50)
		resHpRegen[i] = 1.0
		resMaxHp[i] = 100.0

		resMp[i] = 30 + float32(i%30)
		resMpRegen[i] = 0.5
		resMaxMp[i] = 80.0

		resSp[i] = 20 + float32(i%20)
		resSpRegen[i] = 0.8
		resMaxSp[i] = 60.0
	}
}

// BenchmarkScalarResourceTick runs one regen tick for 100,000 units using a
// plain for loop: hp[i] = min(hp[i]+regen[i], max[i]) for each resource.
// SIMD counterpart: BenchmarkSIMDResourceTick.
func BenchmarkScalarResourceTick(b *testing.B) {
	b.ResetTimer()
	for b.Loop() {
		for i := range resHp {
			resHp[i] = min(resHp[i]+resHpRegen[i], resMaxHp[i])
			resMp[i] = min(resMp[i]+resMpRegen[i], resMaxMp[i])
			resSp[i] = min(resSp[i]+resSpRegen[i], resMaxSp[i])
		}
	}
}
