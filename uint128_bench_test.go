package dwcas_test

import (
	"runtime"
	"sync/atomic"
	"testing"

	"code.hybscloud.com/dwcas"
)

var (
	sinkPlaceN    int
	sinkPlaceU128 *dwcas.Uint128
)

func BenchmarkCAS_Single(b *testing.B) {
	p := dwcas.New(0, 0)
	a := dwcas.Uint128{Lo: 0, Hi: 0}
	c := dwcas.Uint128{Lo: 1, Hi: 1}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		// Toggle between two states to avoid Go-level loads.
		for {
			if _, ok := p.Relaxed(a, c); ok {
				a, c = c, a
				break
			}
			if _, ok := p.Relaxed(c, a); ok {
				a, c = c, a
				break
			}
		}
	}
}

func BenchmarkPlaceAlignedUint128(b *testing.B) {
	buf := make([]byte, 256)
	const off = 7
	if !dwcas.CanPlaceAlignedUint128(buf, off) {
		b.Fatalf("unexpected: cannot place at off=%d", off)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		n, p := dwcas.PlaceAlignedUint128(buf, off)
		p.Lo = uint64(i)
		p.Hi = uint64(i)
		sinkPlaceN = n
		sinkPlaceU128 = p
	}
}

func benchmarkCASContended(b *testing.B, procs int) {
	old := runtime.GOMAXPROCS(procs)
	defer runtime.GOMAXPROCS(old)

	p := dwcas.New(0, 0)
	a := dwcas.Uint128{Lo: 0, Hi: 0}
	c := dwcas.Uint128{Lo: 1, Hi: 1}

	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for {
				if _, ok := p.Relaxed(a, c); ok {
					break
				}
				if _, ok := p.Relaxed(c, a); ok {
					break
				}
			}
		}
	})
}

func BenchmarkCAS_Contended_P2(b *testing.B) { benchmarkCASContended(b, 2) }
func BenchmarkCAS_Contended_P4(b *testing.B) { benchmarkCASContended(b, 4) }
func BenchmarkCAS_Contended_P8(b *testing.B) { benchmarkCASContended(b, 8) }

func BenchmarkCAS_VersionedBump_Single(b *testing.B) {
	// Models a common lock-free pattern: (value, version) updated together.
	p := dwcas.New(0, 0)

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		old := *p
		newv := dwcas.Uint128{Lo: old.Lo + 1, Hi: old.Hi + 1}
		for {
			prev, ok := p.AcqRel(old, newv)
			if ok {
				break
			}
			old = prev
			newv.Lo = old.Lo + 1
			newv.Hi = old.Hi + 1
		}
	}
}

func BenchmarkAtomicUint64Pair_Baseline(b *testing.B) {
	// Baseline only: two independent 64-bit CAS operations are NOT equivalent to a
	// single 128-bit CAS (not atomic as a pair), but provide a rough reference for
	// the overhead of contention and CAS loops.
	var lo, hi uint64
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		oldLo := atomic.LoadUint64(&lo)
		oldHi := atomic.LoadUint64(&hi)
		_ = atomic.CompareAndSwapUint64(&lo, oldLo, oldLo+1)
		_ = atomic.CompareAndSwapUint64(&hi, oldHi, oldHi+1)
	}
}
