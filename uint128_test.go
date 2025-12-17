package dwcas_test

import (
	"runtime"
	"sync"
	"testing"
	"time"
	"unsafe"

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
	methods := []struct {
		name string
		fn   func(p *dwcas.Uint128, old, new dwcas.Uint128) (dwcas.Uint128, bool)
	}{
		{"Relaxed", func(p *dwcas.Uint128, old, new dwcas.Uint128) (dwcas.Uint128, bool) { return p.Relaxed(old, new) }},
		{"Acquire", func(p *dwcas.Uint128, old, new dwcas.Uint128) (dwcas.Uint128, bool) { return p.Acquire(old, new) }},
		{"Release", func(p *dwcas.Uint128, old, new dwcas.Uint128) (dwcas.Uint128, bool) { return p.Release(old, new) }},
		{"AcqRel", func(p *dwcas.Uint128, old, new dwcas.Uint128) (dwcas.Uint128, bool) { return p.AcqRel(old, new) }},
	}

	for _, tt := range methods {
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

func TestNew_AlignmentAndInit(t *testing.T) {
	p := dwcas.New(10, 20)
	if p == nil {
		t.Fatalf("New returned nil")
	}
	if uintptr(unsafe.Pointer(p))%16 != 0 {
		t.Fatalf("New returned misaligned pointer: %p", p)
	}
	if p.Lo != 10 || p.Hi != 20 {
		t.Fatalf("New init mismatch: got (%d,%d)", p.Lo, p.Hi)
	}
}

func TestCAS_Contention_Toggle(t *testing.T) {
	// This test avoids Go-level loads/stores of *p while contended.
	// Each goroutine makes progress only via CAS operations.
	p := dwcas.New(0, 0)
	a := dwcas.Uint128{Lo: 0, Hi: 0}
	b := dwcas.Uint128{Lo: 1, Hi: 1}

	const (
		goroutines = 16
		iters      = 200000
	)

	// Bound runtime for CI and avoid hangs on unexpected backend issues.
	deadline := time.After(5 * time.Second)
	done := make(chan struct{})
	go func() {
		var wg sync.WaitGroup
		wg.Add(goroutines)
		for g := 0; g < goroutines; g++ {
			go func() {
				defer wg.Done()
				for i := 0; i < iters; i++ {
					for {
						if _, ok := p.Relaxed(a, b); ok {
							break
						}
						if _, ok := p.Relaxed(b, a); ok {
							break
						}
						// Bounded backoff for pathological contention.
						runtime.Gosched()
					}
				}
			}()
		}
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// final value is either a or b; both are valid end states.
		if (p.Lo == a.Lo && p.Hi == a.Hi) || (p.Lo == b.Lo && p.Hi == b.Hi) {
			return
		}
		t.Fatalf("final state is invalid: got (%d,%d)", p.Lo, p.Hi)
	case <-deadline:
		t.Fatalf("contention test timed out")
	}
}
