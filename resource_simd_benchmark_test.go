//go:build goexperiment.simd

package mybenchmark

import (
	"simd/archsimd"
	"testing"
)

// ResourceBatch holds regen state for 16 units in SoA Float32x16 layout.
// One Tick() call = 6 SIMD instructions for all 16 units simultaneously.
type ResourceBatch struct {
	Hp, HpRegen, MaxHp archsimd.Float32x16
	Mp, MpRegen, MaxMp archsimd.Float32x16
	Sp, SpRegen, MaxSp archsimd.Float32x16
}

// Tick regenerates Hp, Mp, Sp for all 16 slots: Add + Min per resource.
func (r *ResourceBatch) Tick() {
	r.Hp = r.Hp.Add(r.HpRegen).Min(r.MaxHp)
	r.Mp = r.Mp.Add(r.MpRegen).Min(r.MaxMp)
	r.Sp = r.Sp.Add(r.SpRegen).Min(r.MaxSp)
}

// resBatches holds 6,250 batches × 16 slots = 100,000 units.
// Total size: 6250 × 9 × 64 bytes = 3.6 MB (fits in L3 cache).
var resBatches [6250]ResourceBatch

func init() {
	for b := range resBatches {
		var hp, hpRegen, maxHp [16]float32
		var mp, mpRegen, maxMp [16]float32
		var sp, spRegen, maxSp [16]float32
		for slot := range 16 {
			i := b*16 + slot
			hp[slot] = 50 + float32(i%50)
			hpRegen[slot] = 1.0
			maxHp[slot] = 100.0

			mp[slot] = 30 + float32(i%30)
			mpRegen[slot] = 0.5
			maxMp[slot] = 80.0

			sp[slot] = 20 + float32(i%20)
			spRegen[slot] = 0.8
			maxSp[slot] = 60.0
		}
		resBatches[b].Hp = archsimd.LoadFloat32x16(&hp)
		resBatches[b].HpRegen = archsimd.LoadFloat32x16(&hpRegen)
		resBatches[b].MaxHp = archsimd.LoadFloat32x16(&maxHp)
		resBatches[b].Mp = archsimd.LoadFloat32x16(&mp)
		resBatches[b].MpRegen = archsimd.LoadFloat32x16(&mpRegen)
		resBatches[b].MaxMp = archsimd.LoadFloat32x16(&maxMp)
		resBatches[b].Sp = archsimd.LoadFloat32x16(&sp)
		resBatches[b].SpRegen = archsimd.LoadFloat32x16(&spRegen)
		resBatches[b].MaxSp = archsimd.LoadFloat32x16(&maxSp)
	}
}

// BenchmarkSIMDResourceTick runs one regen tick for 100,000 units as 6,250
// Float32x16 batches: 6 SIMD instructions per batch vs 300,000 scalar ops.
// Scalar baseline: BenchmarkScalarResourceTick.
func BenchmarkSIMDResourceTick(b *testing.B) {
	b.ResetTimer()
	for b.Loop() {
		for i := range resBatches {
			resBatches[i].Tick()
		}
	}
}
