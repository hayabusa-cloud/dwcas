//go:build !amd64 && !arm64

package dwcas_test

import (
	"testing"

	"code.hybscloud.com/dwcas"
)

func TestCompareExchange_UnsupportedArchPanics(t *testing.T) {
	p := dwcas.New(1, 2)
	for _, method := range casMethods {
		t.Run(method.name, func(t *testing.T) {
			mustPanic(t, func() { method.fn(p, dwcas.Uint128{}, dwcas.Uint128{}) })
			mustPanic(t, func() { method.try(p, dwcas.Uint128{}, dwcas.Uint128{}) })
		})
	}
}

func TestPairAccessUnsupportedArchPanics(t *testing.T) {
	p := dwcas.New(1, 2)
	mustPanic(t, func() { p.LoadRelaxed() })
	mustPanic(t, func() { p.LoadAcquire() })
	mustPanic(t, func() { p.StoreRelaxed(dwcas.Uint128{}) })
	mustPanic(t, func() { p.StoreRelease(dwcas.Uint128{}) })
}
