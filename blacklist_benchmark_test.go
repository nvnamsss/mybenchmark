package mybenchmark

import (
	"testing"
)

// generateBlacklist returns a slice of ints [0, size)
func generateBlacklist(size int) []int {
	s := make([]int, size)
	for i := range s {
		s[i] = i
	}
	return s
}

// generateCandidates returns a slice where half are in the blacklist and half are not.
// Elements in [0, blacklistSize) are "present"; elements starting at blacklistSize*10 are "absent".
func generateCandidates(size, blacklistSize int) []int {
	s := make([]int, size)
	for i := range s {
		if i%2 == 0 {
			s[i] = i % blacklistSize // present in blacklist
		} else {
			s[i] = blacklistSize*10 + i // absent from blacklist
		}
	}
	return s
}

// containsLinear checks whether v is in the blacklist via linear scan.
func containsLinear(blacklist []int, v int) bool {
	for _, b := range blacklist {
		if b == v {
			return true
		}
	}
	return false
}

// buildHashmap converts a blacklist slice into a map[int]struct{}.
func buildHashmap(blacklist []int) map[int]struct{} {
	m := make(map[int]struct{}, len(blacklist))
	for _, v := range blacklist {
		m[v] = struct{}{}
	}
	return m
}

// --- Blacklist=5, Candidates=20 ---

func BenchmarkLinearScan_BL5_C20(b *testing.B) {
	bl := generateBlacklist(5)
	candidates := generateCandidates(20, 5)
	b.ResetTimer()
	for b.Loop() {
		for _, c := range candidates {
			_ = containsLinear(bl, c)
		}
	}
}

func BenchmarkHashmap_BL5_C20(b *testing.B) {
	bl := generateBlacklist(5)
	candidates := generateCandidates(20, 5)
	b.ResetTimer()
	for b.Loop() {
		m := buildHashmap(bl)
		for _, c := range candidates {
			_, _ = m[c]
		}
	}
}

func BenchmarkHashmapPrebuilt_BL5_C20(b *testing.B) {
	bl := generateBlacklist(5)
	candidates := generateCandidates(20, 5)
	m := buildHashmap(bl)
	b.ResetTimer()
	for b.Loop() {
		for _, c := range candidates {
			_, _ = m[c]
		}
	}
}

// --- Blacklist=5, Candidates=40 ---

func BenchmarkLinearScan_BL5_C40(b *testing.B) {
	bl := generateBlacklist(5)
	candidates := generateCandidates(40, 5)
	b.ResetTimer()
	for b.Loop() {
		for _, c := range candidates {
			_ = containsLinear(bl, c)
		}
	}
}

func BenchmarkHashmap_BL5_C40(b *testing.B) {
	bl := generateBlacklist(5)
	candidates := generateCandidates(40, 5)
	b.ResetTimer()
	for b.Loop() {
		m := buildHashmap(bl)
		for _, c := range candidates {
			_, _ = m[c]
		}
	}
}

func BenchmarkHashmapPrebuilt_BL5_C40(b *testing.B) {
	bl := generateBlacklist(5)
	candidates := generateCandidates(40, 5)
	m := buildHashmap(bl)
	b.ResetTimer()
	for b.Loop() {
		for _, c := range candidates {
			_, _ = m[c]
		}
	}
}

// --- Blacklist=5, Candidates=400 ---

func BenchmarkLinearScan_BL5_C400(b *testing.B) {
	bl := generateBlacklist(5)
	candidates := generateCandidates(400, 5)
	b.ResetTimer()
	for b.Loop() {
		for _, c := range candidates {
			_ = containsLinear(bl, c)
		}
	}
}

func BenchmarkHashmap_BL5_C400(b *testing.B) {
	bl := generateBlacklist(5)
	candidates := generateCandidates(400, 5)
	b.ResetTimer()
	for b.Loop() {
		m := buildHashmap(bl)
		for _, c := range candidates {
			_, _ = m[c]
		}
	}
}

func BenchmarkHashmapPrebuilt_BL5_C400(b *testing.B) {
	bl := generateBlacklist(5)
	candidates := generateCandidates(400, 5)
	m := buildHashmap(bl)
	b.ResetTimer()
	for b.Loop() {
		for _, c := range candidates {
			_, _ = m[c]
		}
	}
}

// --- Blacklist=10, Candidates=20 ---

func BenchmarkLinearScan_BL10_C20(b *testing.B) {
	bl := generateBlacklist(10)
	candidates := generateCandidates(20, 10)
	b.ResetTimer()
	for b.Loop() {
		for _, c := range candidates {
			_ = containsLinear(bl, c)
		}
	}
}

func BenchmarkHashmap_BL10_C20(b *testing.B) {
	bl := generateBlacklist(10)
	candidates := generateCandidates(20, 10)
	b.ResetTimer()
	for b.Loop() {
		m := buildHashmap(bl)
		for _, c := range candidates {
			_, _ = m[c]
		}
	}
}

func BenchmarkHashmapPrebuilt_BL10_C20(b *testing.B) {
	bl := generateBlacklist(10)
	candidates := generateCandidates(20, 10)
	m := buildHashmap(bl)
	b.ResetTimer()
	for b.Loop() {
		for _, c := range candidates {
			_, _ = m[c]
		}
	}
}

// --- Blacklist=10, Candidates=40 ---

func BenchmarkLinearScan_BL10_C40(b *testing.B) {
	bl := generateBlacklist(10)
	candidates := generateCandidates(40, 10)
	b.ResetTimer()
	for b.Loop() {
		for _, c := range candidates {
			_ = containsLinear(bl, c)
		}
	}
}

func BenchmarkHashmap_BL10_C40(b *testing.B) {
	bl := generateBlacklist(10)
	candidates := generateCandidates(40, 10)
	b.ResetTimer()
	for b.Loop() {
		m := buildHashmap(bl)
		for _, c := range candidates {
			_, _ = m[c]
		}
	}
}

func BenchmarkHashmapPrebuilt_BL10_C40(b *testing.B) {
	bl := generateBlacklist(10)
	candidates := generateCandidates(40, 10)
	m := buildHashmap(bl)
	b.ResetTimer()
	for b.Loop() {
		for _, c := range candidates {
			_, _ = m[c]
		}
	}
}

// --- Blacklist=10, Candidates=400 ---

func BenchmarkLinearScan_BL10_C400(b *testing.B) {
	bl := generateBlacklist(10)
	candidates := generateCandidates(400, 10)
	b.ResetTimer()
	for b.Loop() {
		for _, c := range candidates {
			_ = containsLinear(bl, c)
		}
	}
}

func BenchmarkHashmap_BL10_C400(b *testing.B) {
	bl := generateBlacklist(10)
	candidates := generateCandidates(400, 10)
	b.ResetTimer()
	for b.Loop() {
		m := buildHashmap(bl)
		for _, c := range candidates {
			_, _ = m[c]
		}
	}
}

func BenchmarkHashmapPrebuilt_BL10_C400(b *testing.B) {
	bl := generateBlacklist(10)
	candidates := generateCandidates(400, 10)
	m := buildHashmap(bl)
	b.ResetTimer()
	for b.Loop() {
		for _, c := range candidates {
			_, _ = m[c]
		}
	}
}

// --- Blacklist=100, Candidates=20 ---

func BenchmarkLinearScan_BL100_C20(b *testing.B) {
	bl := generateBlacklist(100)
	candidates := generateCandidates(20, 100)
	b.ResetTimer()
	for b.Loop() {
		for _, c := range candidates {
			_ = containsLinear(bl, c)
		}
	}
}

func BenchmarkHashmap_BL100_C20(b *testing.B) {
	bl := generateBlacklist(100)
	candidates := generateCandidates(20, 100)
	b.ResetTimer()
	for b.Loop() {
		m := buildHashmap(bl)
		for _, c := range candidates {
			_, _ = m[c]
		}
	}
}

func BenchmarkHashmapPrebuilt_BL100_C20(b *testing.B) {
	bl := generateBlacklist(100)
	candidates := generateCandidates(20, 100)
	m := buildHashmap(bl)
	b.ResetTimer()
	for b.Loop() {
		for _, c := range candidates {
			_, _ = m[c]
		}
	}
}

// --- Blacklist=100, Candidates=40 ---

func BenchmarkLinearScan_BL100_C40(b *testing.B) {
	bl := generateBlacklist(100)
	candidates := generateCandidates(40, 100)
	b.ResetTimer()
	for b.Loop() {
		for _, c := range candidates {
			_ = containsLinear(bl, c)
		}
	}
}

func BenchmarkHashmap_BL100_C40(b *testing.B) {
	bl := generateBlacklist(100)
	candidates := generateCandidates(40, 100)
	b.ResetTimer()
	for b.Loop() {
		m := buildHashmap(bl)
		for _, c := range candidates {
			_, _ = m[c]
		}
	}
}

func BenchmarkHashmapPrebuilt_BL100_C40(b *testing.B) {
	bl := generateBlacklist(100)
	candidates := generateCandidates(40, 100)
	m := buildHashmap(bl)
	b.ResetTimer()
	for b.Loop() {
		for _, c := range candidates {
			_, _ = m[c]
		}
	}
}

// --- Blacklist=100, Candidates=400 ---

func BenchmarkLinearScan_BL100_C400(b *testing.B) {
	bl := generateBlacklist(100)
	candidates := generateCandidates(400, 100)
	b.ResetTimer()
	for b.Loop() {
		for _, c := range candidates {
			_ = containsLinear(bl, c)
		}
	}
}

func BenchmarkHashmap_BL100_C400(b *testing.B) {
	bl := generateBlacklist(100)
	candidates := generateCandidates(400, 100)
	b.ResetTimer()
	for b.Loop() {
		m := buildHashmap(bl)
		for _, c := range candidates {
			_, _ = m[c]
		}
	}
}

func BenchmarkHashmapPrebuilt_BL100_C400(b *testing.B) {
	bl := generateBlacklist(100)
	candidates := generateCandidates(400, 100)
	m := buildHashmap(bl)
	b.ResetTimer()
	for b.Loop() {
		for _, c := range candidates {
			_, _ = m[c]
		}
	}
}
