//go:build dwcasdebug

package dwcas_test

import (
	"testing"
	"unsafe"

	"code.hybscloud.com/dwcas"
)

func TestCheckAligned_PanicsOnMisalignment(t *testing.T) {
	// Construct an 8(mod16) pointer. checkAligned must panic before any dereference.
	buf := make([]byte, 64)
	base := uintptr(unsafe.Pointer(unsafe.SliceData(buf)))
	off := (uintptr(8) - (base & uintptr(15)) + uintptr(16)) & uintptr(15)
	if off == 0 {
		off = 8
	}
	p := (*dwcas.Uint128)(unsafe.Pointer(unsafe.SliceData(buf[off:])))

	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic for misaligned pointer")
		}
	}()

	_, _ = p.AcqRel(dwcas.Uint128{}, dwcas.Uint128{Lo: 1, Hi: 1})
}

func TestCheckAligned_PanicsOnNilPointer(t *testing.T) {
	var p *dwcas.Uint128
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic for nil pointer")
		}
	}()
	_, _ = p.Relaxed(dwcas.Uint128{}, dwcas.Uint128{Lo: 1, Hi: 1})
}
