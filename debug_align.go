//go:build dwcasdebug

package dwcas

import (
	"fmt"
	"unsafe"
)

func checkAligned(p *Uint128) {
	if p == nil {
		panic("dwcas: *Uint128 is nil")
	}
	if uintptr(unsafe.Pointer(p))%16 != 0 {
		panic(fmt.Sprintf("dwcas: *Uint128 at %p is not 16-byte aligned", p))
	}
}
