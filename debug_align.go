//go:build dwcasdebug

package dwcas

import (
	"strconv"
	"unsafe"
)

func checkAligned(p *Uint128) {
	addr := uintptr(unsafe.Pointer(p))
	if addr == 0 || addr%16 != 0 {
		panic(invalidAddressMessage(addr))
	}
}

func invalidAddressMessage(addr uintptr) string {
	if addr == 0 {
		return "dwcas: *Uint128 is nil"
	}
	hex := strconv.FormatUint(uint64(addr), 16)
	return "dwcas: *Uint128 at 0x" + hex + " is not 16-byte aligned"
}
