//go:build dwcasdebug

package dwcas_test

import (
	"testing"
	"unsafe"

	"code.hybscloud.com/dwcas"
)

func misalignedPair() *dwcas.Uint128 {
	_, aligned := dwcas.PlaceAlignedUint128(make([]byte, 64), 0)
	return (*dwcas.Uint128)(unsafe.Add(unsafe.Pointer(aligned), 8))
}

func TestTryDebugGuard(t *testing.T) {
	misaligned := misalignedPair()
	for _, method := range casMethods {
		t.Run(method.name, func(t *testing.T) {
			mustPanic(t, func() { method.try(nil, dwcas.Uint128{}, dwcas.Uint128{}) })
			mustPanic(t, func() { method.try(misaligned, dwcas.Uint128{}, dwcas.Uint128{}) })
		})
	}
}

func TestPairAccessDebugGuard(t *testing.T) {
	misaligned := misalignedPair()
	for _, p := range []*dwcas.Uint128{nil, misaligned} {
		mustPanic(t, func() { p.LoadRelaxed() })
		mustPanic(t, func() { p.LoadAcquire() })
		mustPanic(t, func() { p.StoreRelaxed(dwcas.Uint128{}) })
		mustPanic(t, func() { p.StoreRelease(dwcas.Uint128{}) })
	}
}
