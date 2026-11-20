//go:build dwcasdebug

package dwcas_test

import (
	"testing"

	"code.hybscloud.com/dwcas"
)

func TestCompareExchangeDebugGuard(t *testing.T) {
	misaligned := misalignedPair()
	for _, method := range casMethods {
		t.Run(method.name, func(t *testing.T) {
			mustPanic(t, func() { method.fn(nil, dwcas.Uint128{}, dwcas.Uint128{}) })
			mustPanic(t, func() { method.fn(misaligned, dwcas.Uint128{}, dwcas.Uint128{}) })
		})
	}
}
