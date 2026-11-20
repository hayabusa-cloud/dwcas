// ©Hayabusa Cloud Co., Ltd. 2026. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

//go:build amd64 || arm64

package dwcas_test

import (
	"testing"

	"code.hybscloud.com/dwcas"
)

func TestPairLoadStore(t *testing.T) {
	initial := dwcas.Uint128{Lo: 1 << 63, Hi: 7}
	next := dwcas.Uint128{Lo: 11, Hi: 1 << 62}
	p := dwcas.New(initial.Lo, initial.Hi)
	if got := p.LoadRelaxed(); got != initial {
		t.Fatalf("LoadRelaxed returned %v, want %v", got, initial)
	}
	p.StoreRelease(next)
	if got := p.LoadAcquire(); got != next {
		t.Fatalf("LoadAcquire returned %v, want %v", got, next)
	}
	if !p.TryRelaxed(next, initial) {
		t.Fatal("CAS failed to match the pair written by StoreRelease")
	}
	var zero dwcas.Uint128
	p.StoreRelaxed(zero)
	if got := p.LoadRelaxed(); got != zero {
		t.Fatalf("load after StoreRelaxed returned %v, want %v", got, zero)
	}
}
