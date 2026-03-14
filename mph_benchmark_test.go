package mybenchmark

import (
	"runtime"
	"strconv"
	"testing"

	"github.com/dgryski/go-mph"
)

var (
	mphSinkBool   bool
	mphSinkUint32 uint32
	mphSinkInt32  int32
)

var mphDatasetSizes = []int{1_000, 100_000, 1_000_000}

func benchmarkKeys(n int) []string {
	keys := make([]string, n)
	for i := 0; i < n; i++ {
		keys[i] = "key_" + strconv.Itoa(i)
	}
	return keys
}

func benchmarkMissKeys(keys []string) []string {
	misses := make([]string, len(keys))
	for i, k := range keys {
		misses[i] = "miss_" + k
	}
	return misses
}

func buildMapIndex(keys []string) map[string]uint32 {
	idx := make(map[string]uint32, len(keys))
	for i, k := range keys {
		idx[k] = uint32(i)
	}
	return idx
}

func buildMPHIndex(keys []string) *mph.Table {
	return mph.New(keys)
}

func mphContains(table *mph.Table, keys []string, key string) (int32, bool) {
	i := table.Query(key)
	if i < 0 || int(i) >= len(keys) {
		return 0, false
	}
	if keys[i] != key {
		return 0, false
	}
	return i, true
}

func measureRetainedHeapBytes(build func() any) uint64 {
	var before runtime.MemStats
	var after runtime.MemStats

	runtime.GC()
	runtime.ReadMemStats(&before)

	obj := build()

	runtime.GC()
	runtime.ReadMemStats(&after)
	runtime.KeepAlive(obj)

	if after.HeapAlloc <= before.HeapAlloc {
		return 0
	}
	return after.HeapAlloc - before.HeapAlloc
}

func BenchmarkMapVsMPH(b *testing.B) {
	b.Run("Lookup", func(b *testing.B) {
		for _, size := range mphDatasetSizes {
			size := size
			b.Run("Size_"+strconv.Itoa(size), func(b *testing.B) {
				keys := benchmarkKeys(size)
				misses := benchmarkMissKeys(keys)

				mapIndex := buildMapIndex(keys)
				mphIndex := buildMPHIndex(keys)

				b.Run("Map_Hit", func(b *testing.B) {
					probes := keys
					n := len(probes)
					b.ReportAllocs()
					b.ResetTimer()
					for i := 0; i < b.N; i++ {
						key := probes[i%n]
						v, ok := mapIndex[key]
						mphSinkUint32 = v
						mphSinkBool = ok
					}
				})

				b.Run("Map_Miss", func(b *testing.B) {
					probes := misses
					n := len(probes)
					b.ReportAllocs()
					b.ResetTimer()
					for i := 0; i < b.N; i++ {
						key := probes[i%n]
						v, ok := mapIndex[key]
						mphSinkUint32 = v
						mphSinkBool = ok
					}
				})

				b.Run("MPH_Hit", func(b *testing.B) {
					probes := keys
					n := len(probes)
					b.ReportAllocs()
					b.ResetTimer()
					for i := 0; i < b.N; i++ {
						key := probes[i%n]
						v, ok := mphContains(mphIndex, keys, key)
						mphSinkInt32 = v
						mphSinkBool = ok
					}
				})

				b.Run("MPH_Miss", func(b *testing.B) {
					probes := misses
					n := len(probes)
					b.ReportAllocs()
					b.ResetTimer()
					for i := 0; i < b.N; i++ {
						key := probes[i%n]
						v, ok := mphContains(mphIndex, keys, key)
						mphSinkInt32 = v
						mphSinkBool = ok
					}
				})
			})
		}
	})

	b.Run("Build", func(b *testing.B) {
		for _, size := range mphDatasetSizes {
			size := size
			b.Run("Size_"+strconv.Itoa(size), func(b *testing.B) {
				keys := benchmarkKeys(size)

				b.Run("Map_Build", func(b *testing.B) {
					b.ReportAllocs()
					b.ResetTimer()
					for i := 0; i < b.N; i++ {
						idx := buildMapIndex(keys)
						mphSinkUint32 = uint32(len(idx))
					}
				})

				b.Run("MPH_Build", func(b *testing.B) {
					b.ReportAllocs()
					b.ResetTimer()
					for i := 0; i < b.N; i++ {
						t := buildMPHIndex(keys)
						mphSinkInt32 = int32(len(t.Values))
					}
				})
			})
		}
	})

	b.Run("ParallelLookup", func(b *testing.B) {
		for _, size := range mphDatasetSizes {
			size := size
			b.Run("Size_"+strconv.Itoa(size), func(b *testing.B) {
				keys := benchmarkKeys(size)
				misses := benchmarkMissKeys(keys)
				mapIndex := buildMapIndex(keys)
				mphIndex := buildMPHIndex(keys)

				b.Run("Map_Hit", func(b *testing.B) {
					probes := keys
					n := len(probes)
					b.ReportAllocs()
					b.ResetTimer()
					b.RunParallel(func(pb *testing.PB) {
						i := 0
						for pb.Next() {
							k := probes[i%n]
							i++
							v, ok := mapIndex[k]
							mphSinkUint32 = v
							mphSinkBool = ok
						}
					})
				})

				b.Run("Map_Miss", func(b *testing.B) {
					probes := misses
					n := len(probes)
					b.ReportAllocs()
					b.ResetTimer()
					b.RunParallel(func(pb *testing.PB) {
						i := 0
						for pb.Next() {
							k := probes[i%n]
							i++
							v, ok := mapIndex[k]
							mphSinkUint32 = v
							mphSinkBool = ok
						}
					})
				})

				b.Run("MPH_Hit", func(b *testing.B) {
					probes := keys
					n := len(probes)
					b.ReportAllocs()
					b.ResetTimer()
					b.RunParallel(func(pb *testing.PB) {
						i := 0
						for pb.Next() {
							k := probes[i%n]
							i++
							v, ok := mphContains(mphIndex, keys, k)
							mphSinkInt32 = v
							mphSinkBool = ok
						}
					})
				})

				b.Run("MPH_Miss", func(b *testing.B) {
					probes := misses
					n := len(probes)
					b.ReportAllocs()
					b.ResetTimer()
					b.RunParallel(func(pb *testing.PB) {
						i := 0
						for pb.Next() {
							k := probes[i%n]
							i++
							v, ok := mphContains(mphIndex, keys, k)
							mphSinkInt32 = v
							mphSinkBool = ok
						}
					})
				})
			})
		}
	})

	b.Run("RetainedMemory", func(b *testing.B) {
		for _, size := range mphDatasetSizes {
			size := size
			b.Run("Size_"+strconv.Itoa(size), func(b *testing.B) {
				keys := benchmarkKeys(size)

				mapBytes := measureRetainedHeapBytes(func() any {
					return buildMapIndex(keys)
				})

				mphBytes := measureRetainedHeapBytes(func() any {
					return buildMPHIndex(keys)
				})

				// Keep this metric separate: MPH miss checks in this suite retain keys.
				mphWithKeysBytes := measureRetainedHeapBytes(func() any {
					return struct {
						table *mph.Table
						keys  []string
					}{
						table: buildMPHIndex(keys),
						keys:  keys,
					}
				})

				b.ReportMetric(float64(mapBytes), "map_heap_B")
				b.ReportMetric(float64(mphBytes), "mph_heap_B")
				b.ReportMetric(float64(mphWithKeysBytes), "mph_plus_keys_heap_B")

				if mapBytes > 0 {
					b.ReportMetric(float64(mphBytes)/float64(mapBytes), "mph_to_map_heap_ratio")
					b.ReportMetric(float64(mphWithKeysBytes)/float64(mapBytes), "mph_plus_keys_to_map_heap_ratio")
				}
			})
		}
	})
}
