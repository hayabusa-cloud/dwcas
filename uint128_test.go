//go:build amd64 || arm64

package dwcas_test

import (
	"sync"
	"testing"

	"code.hybscloud.com/dwcas"
)

func TestCAS_Success(t *testing.T) {
	p := dwcas.New(1, 2)
	prev, ok := p.AcqRel(dwcas.Uint128{Lo: 1, Hi: 2}, dwcas.Uint128{Lo: 3, Hi: 4})
	if !ok {
		t.Fatalf("AcqRel should succeed")
	}
	if prev != (dwcas.Uint128{Lo: 1, Hi: 2}) {
		t.Fatalf("prev mismatch on success: got (%d,%d)", prev.Lo, prev.Hi)
	}
	if p.Lo != 3 || p.Hi != 4 {
		t.Fatalf("value mismatch after AcqRel: got (%d,%d)", p.Lo, p.Hi)
	}
}

func TestCAS_Fail(t *testing.T) {
	p := dwcas.New(1, 2)
	prev, ok := p.AcqRel(dwcas.Uint128{Lo: 0, Hi: 2}, dwcas.Uint128{Lo: 3, Hi: 4})
	if ok {
		t.Fatalf("AcqRel should fail")
	}
	if prev != (dwcas.Uint128{Lo: 1, Hi: 2}) {
		t.Fatalf("prev mismatch on failure: got (%d,%d)", prev.Lo, prev.Hi)
	}
	if p.Lo != 1 || p.Hi != 2 {
		t.Fatalf("value changed on failed AcqRel: got (%d,%d)", p.Lo, p.Hi)
	}
}

func TestCompareExchange_PrevConvergesWithoutExtraLoads(t *testing.T) {
	for _, tt := range casMethods {
		t.Run(tt.name, func(t *testing.T) {
			p := dwcas.New(5, 6)
			stale := dwcas.Uint128{Lo: 0, Hi: 0}
			desired := dwcas.Uint128{Lo: 7, Hi: 8}

			prev, swapped := tt.fn(p, stale, desired)
			if swapped {
				t.Fatalf("unexpected success with stale expected")
			}
			if prev != (dwcas.Uint128{Lo: 5, Hi: 6}) {
				t.Fatalf("prev mismatch on failure: got (%d,%d)", prev.Lo, prev.Hi)
			}

			prev2, swapped2 := tt.fn(p, prev, desired)
			if !swapped2 {
				t.Fatalf("expected success after updating expected")
			}
			if prev2 != prev {
				t.Fatalf("prev mismatch on success: got (%d,%d)", prev2.Lo, prev2.Hi)
			}
			if p.Lo != desired.Lo || p.Hi != desired.Hi {
				t.Fatalf("value mismatch after swap: got (%d,%d)", p.Lo, p.Hi)
			}
		})
	}
}

func TestCASConcurrentUpdates(t *testing.T) {
	const writers = 4
	const updatesPerWriter = 1 << 14
	p := dwcas.New(0, 0)
	start := make(chan struct{})
	var wg sync.WaitGroup
	for range writers {
		wg.Go(func() {
			<-start
			for completed := 0; completed < updatesPerWriter; {
				old := p.LoadRelaxed()
				next := dwcas.Uint128{Lo: old.Lo + 1, Hi: old.Hi + 2}
				if p.TryRelaxed(old, next) {
					completed++
				}
			}
		})
	}
	close(start)
	wg.Wait()

	const total = writers * updatesPerWriter
	want := dwcas.Uint128{Lo: total, Hi: 2 * total}
	if got := *p; got != want {
		t.Fatalf("after %d successful swaps, got %v, want %v", total, got, want)
	}
}
