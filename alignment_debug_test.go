//go:build dwcasdebug

package dwcas_test

import (
	"testing"
	"unsafe"

	"code.hybscloud.com/dwcas"
)

func TestAcqRel_PanicsOnMisalignedPointer_DebugGuard(t *testing.T) {
	buf := make([]byte, 64)
	// &buf[1] is intentionally misaligned for 16-byte compare-and-swap.
	p := (*dwcas.Uint128)(unsafe.Pointer(&buf[1]))

	mustPanic(t, func() {
		_, _ = p.AcqRel(dwcas.Uint128{}, dwcas.Uint128{Lo: 1, Hi: 1})
	})
}
