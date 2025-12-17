//go:build !amd64 && !arm64

package dwcas_test

import (
	"testing"

	"code.hybscloud.com/dwcas"
)

func TestCompareExchange_UnsupportedArchPanics(t *testing.T) {
	p := dwcas.New(1, 2)
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic on unsupported architecture")
		}
	}()
	_, _ = p.Relaxed(dwcas.Uint128{Lo: 1, Hi: 2}, dwcas.Uint128{Lo: 3, Hi: 4})
}
