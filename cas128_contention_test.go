// ©Hayabusa Cloud Co., Ltd. 2026. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

//go:build amd64 || arm64

package dwcas_test

import (
	"sync"
	"testing"

	"code.hybscloud.com/dwcas"
)

func TestTry(t *testing.T) {
	old := dwcas.Uint128{Lo: 1 << 60, Hi: 1 << 62}
	next := dwcas.Uint128{Lo: 7, Hi: 11}
	for _, method := range casMethods {
		t.Run(method.name, func(t *testing.T) {
			p := dwcas.New(old.Lo, old.Hi)
			for _, mismatch := range []dwcas.Uint128{{Hi: old.Hi}, {Lo: old.Lo}} {
				if swapped := method.try(p, mismatch, next); swapped || *p != old {
					t.Fatalf("mismatch %v: got (%t, %v), want (false, %v)", mismatch, swapped, *p, old)
				}
			}
			if !method.try(p, old, old) {
				t.Fatal("equal replacement must succeed")
			}
			if !method.try(p, old, next) || *p != next {
				t.Fatalf("expected successful replacement with %v, got %v", next, *p)
			}
		})
	}
}

func TestCASFailureReturnsWholePair(t *testing.T) {
	a := dwcas.Uint128{Lo: 1, Hi: 0}
	b := dwcas.Uint128{Lo: 0, Hi: 1}
	p := dwcas.New(a.Lo, a.Hi)

	started := make(chan struct{})
	ctx := t.Context()
	var writer sync.WaitGroup
	writer.Go(func() {
		close(started)
		for ctx.Err() == nil {
			for _, method := range casMethods {
				if !method.try(p, a, b) || !method.try(p, b, a) {
					t.Errorf("Try%s returned false for the sole writer", method.name)
					return
				}
			}
			p.StoreRelaxed(b)
			p.StoreRelease(a)
		}
	})
	t.Cleanup(writer.Wait)
	<-started

	// Neither stored pair is zero. Every comparison must fail with a or b.
	for range 1 << 14 {
		for _, got := range [...]dwcas.Uint128{p.LoadRelaxed(), p.LoadAcquire()} {
			if got != a && got != b {
				t.Fatalf("load returned %v, want %v or %v", got, a, b)
			}
		}
		for _, method := range casMethods {
			if method.try(p, dwcas.Uint128{}, dwcas.Uint128{}) {
				t.Fatalf("Try%s accepted a pair the writer never stored", method.name)
			}
			prev, swapped := method.fn(p, dwcas.Uint128{}, dwcas.Uint128{})
			if swapped || (prev != a && prev != b) {
				t.Fatalf("%s returned (%v, %t), want (%v or %v, false)", method.name, prev, swapped, a, b)
			}
		}
	}
}
