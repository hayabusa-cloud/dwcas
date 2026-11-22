//go:build amd64 || arm64

package dwcas_test

import (
	"errors"
	"runtime"
	"sync/atomic"
	"testing"

	"code.hybscloud.com/dwcas"
)

func BenchmarkCAS_Single(b *testing.B) {
	p := dwcas.New(0, 0)
	a := dwcas.Uint128{Lo: 0, Hi: 0}
	c := dwcas.Uint128{Lo: 1, Hi: 1}

	b.ReportAllocs()
	for range b.N {
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

func BenchmarkCASResult(b *testing.B) {
	value := dwcas.Uint128{Lo: 1, Hi: 2}
	for _, tc := range []struct {
		name string
		old  dwcas.Uint128
	}{
		{"Match", value},
		{"Mismatch", dwcas.Uint128{Lo: 1, Hi: 3}},
	} {
		b.Run(tc.name+"/Bool", func(b *testing.B) {
			p := dwcas.New(value.Lo, value.Hi)
			for b.Loop() {
				p.TryAcqRel(tc.old, value)
			}
		})
		b.Run(tc.name+"/Pair", func(b *testing.B) {
			p := dwcas.New(value.Lo, value.Hi)
			for b.Loop() {
				p.AcqRel(tc.old, value)
			}
		})
	}
}

func BenchmarkPairAccess(b *testing.B) {
	p := dwcas.New(1, 2)
	value := dwcas.Uint128{Lo: 3, Hi: 4}
	b.Run("LoadRelaxed", func(b *testing.B) {
		for b.Loop() {
			p.LoadRelaxed()
		}
	})
	b.Run("LoadViaCAS", func(b *testing.B) {
		for b.Loop() {
			p.Relaxed(dwcas.Uint128{}, dwcas.Uint128{})
		}
	})
	b.Run("LoadAcquire", func(b *testing.B) {
		for b.Loop() {
			p.LoadAcquire()
		}
	})
	b.Run("StoreRelaxed", func(b *testing.B) {
		for b.Loop() {
			p.StoreRelaxed(value)
		}
	})
	b.Run("StoreRelease", func(b *testing.B) {
		for b.Loop() {
			p.StoreRelease(value)
		}
	})
}

func BenchmarkPairLoadShared(b *testing.B) {
	p := dwcas.New(1, 2)
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			p.LoadRelaxed()
		}
	})
}

func BenchmarkPairHandoff(b *testing.B) {
	// Start with an odd request so the consumer completes one warm-up exchange.
	slot := dwcas.New(1, ^uint64(1))
	rounds := b.N
	done := make(chan struct{})
	go func() {
		for i := 0; i <= rounds; i++ {
			request := waitHandoffTurn(slot, 1)
			slot.StoreRelease(dwcas.Uint128{Lo: request.Lo + 1, Hi: ^request.Hi})
		}
		close(done)
	}()
	waitHandoffTurn(slot, 0)

	seq := uint64(1)
	mismatches := 0
	b.ReportAllocs()
	b.ResetTimer()
	start, startErr := processCPUNanos()
	for range rounds {
		seq += 2
		slot.StoreRelease(dwcas.Uint128{Lo: seq, Hi: ^seq})
		ack := waitHandoffTurn(slot, 0)
		if ack.Lo != seq+1 || ack.Hi != seq {
			mismatches++
		}
	}
	end, endErr := processCPUNanos()
	b.StopTimer()
	<-done

	if mismatches != 0 {
		b.Fatalf("%d of %d acknowledgments had the wrong sequence or check word", mismatches, rounds)
	}
	if err := errors.Join(startErr, endErr); err != nil {
		b.Fatal(err)
	}
	if cpu := end - start; cpu > 0 {
		b.ReportMetric(float64(cpu)/float64(rounds), "cpu-ns/op")
	}
}

func waitHandoffTurn(slot *dwcas.Uint128, parity uint64) dwcas.Uint128 {
	for {
		for range 8 {
			pair := slot.LoadAcquire()
			if pair.Lo&1 == parity {
				return pair
			}
		}
		runtime.Gosched()
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
	for range b.N {
		old := p.LoadRelaxed()
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

func BenchmarkCAS_VersionedBump_Parallel(b *testing.B) {
	run := func(name string, worker func(*dwcas.Uint128, *testing.PB) uint64) {
		b.Run(name, func(b *testing.B) {
			p := dwcas.New(0, 0)
			var retries atomic.Uint64
			b.ReportAllocs()
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				retries.Add(worker(p, pb))
			})
			b.StopTimer()

			want := dwcas.Uint128{Lo: uint64(b.N), Hi: uint64(b.N)}
			if got := p.LoadRelaxed(); got != want {
				b.Fatalf("after %d updates got %v, want %v", b.N, got, want)
			}
			b.ReportMetric(float64(retries.Load())/float64(b.N), "retries/op")
		})
	}

	run("Pair", func(p *dwcas.Uint128, pb *testing.PB) uint64 {
		var old dwcas.Uint128
		var retries uint64
		for pb.Next() {
			for {
				next := dwcas.Uint128{Lo: old.Lo + 1, Hi: old.Hi + 1}
				prev, swapped := p.AcqRel(old, next)
				if swapped {
					old = next
					break
				}
				old = prev
				retries++
			}
		}
		return retries
	})

	run("TryReload", func(p *dwcas.Uint128, pb *testing.PB) uint64 {
		var old dwcas.Uint128
		var retries uint64
		for pb.Next() {
			for {
				next := dwcas.Uint128{Lo: old.Lo + 1, Hi: old.Hi + 1}
				if p.TryAcqRel(old, next) {
					old = next
					break
				}
				old = p.LoadRelaxed()
				retries++
			}
		}
		return retries
	})
}

func BenchmarkAtomicUint64Pair_Baseline(b *testing.B) {
	// Baseline only: two independent 64-bit CAS operations are NOT equivalent to a
	// single 128-bit CAS (not atomic as a pair), but provide a rough reference for
	// the overhead of contention and CAS loops.
	var lo, hi uint64
	b.ReportAllocs()
	for range b.N {
		oldLo := atomic.LoadUint64(&lo)
		oldHi := atomic.LoadUint64(&hi)
		_ = atomic.CompareAndSwapUint64(&lo, oldLo, oldLo+1)
		_ = atomic.CompareAndSwapUint64(&hi, oldHi, oldHi+1)
	}
}
