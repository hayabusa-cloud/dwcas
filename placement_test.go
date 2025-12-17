package dwcas_test

import (
	"testing"
	"unsafe"

	"code.hybscloud.com/dwcas"
)

func TestCanPlaceAlignedUint128(t *testing.T) {
	buf := make([]byte, 64)

	tests := []struct {
		name string
		p    []byte
		off  int
		want bool
	}{
		{name: "nil slice", p: nil, off: 0, want: false},
		{name: "off<0", p: buf, off: -1, want: false},
		{name: "off>len", p: buf, off: len(buf) + 1, want: false},
		{name: "off==len", p: buf, off: len(buf), want: false},
		{name: "insufficient (30)", p: make([]byte, 30), off: 0, want: false},
		{name: "exact (31)", p: make([]byte, 31), off: 0, want: true},
		{name: "sufficient", p: buf, off: 0, want: true},
		{name: "boundary ok", p: buf, off: len(buf) - 31, want: true},
		{name: "boundary short", p: buf, off: len(buf) - 30, want: false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if got := dwcas.CanPlaceAlignedUint128(tt.p, tt.off); got != tt.want {
				t.Fatalf("CanPlaceAlignedUint128(len=%d, off=%d) = %v, want %v", len(tt.p), tt.off, got, tt.want)
			}
		})
	}
}

func TestPlaceAlignedUint128_PanicWhenCannotPlace(t *testing.T) {
	mustPanic(t, func() {
		_, _ = dwcas.PlaceAlignedUint128(make([]byte, 30), 0)
	})

	mustPanic(t, func() {
		_, _ = dwcas.PlaceAlignedUint128(make([]byte, 31), 1)
	})

	mustPanic(t, func() {
		_, _ = dwcas.PlaceAlignedUint128(make([]byte, 64), -1)
	})
}

func TestPlaceAlignedUint128_AlignmentAndBounds(t *testing.T) {
	buf := make([]byte, 256)
	base := uintptr(unsafe.Pointer(unsafe.SliceData(buf)))
	end := base + uintptr(len(buf))

	offs := []int{0, 1, 7, 15, 16, 31, 32, 63, 97, 128, 225}
	for _, off := range offs {
		off := off
		t.Run("off="+itoa(off), func(t *testing.T) {
			if !dwcas.CanPlaceAlignedUint128(buf, off) {
				// Skip offsets that are out of range for this fixed buffer.
				return
			}

			n, p := dwcas.PlaceAlignedUint128(buf, off)
			if p == nil {
				t.Fatalf("PlaceAlignedUint128 returned nil pointer")
			}
			if n < 16 || n > 31 {
				t.Fatalf("n out of range: got %d, want [16..31]", n)
			}

			addr := uintptr(unsafe.Pointer(p))
			start := uintptr(unsafe.Pointer(&buf[off]))
			if addr%16 != 0 {
				t.Fatalf("returned pointer is not 16-byte aligned: %p", p)
			}
			if addr < start {
				t.Fatalf("returned address is before &buf[off]: off=%d addr=%#x start=%#x", off, addr, start)
			}
			if addr < base || addr+16 > end {
				t.Fatalf("returned address is out of buf bounds: addr=%#x buf=[%#x..%#x)", addr, base, end)
			}

			pad := int(addr - start)
			if pad < 0 || pad > 15 {
				t.Fatalf("padding out of range: got %d, want [0..15]", pad)
			}
			if n != pad+16 {
				t.Fatalf("n mismatch: got %d, want pad+16=%d", n, pad+16)
			}
		})
	}
}

func mustPanic(t *testing.T, f func()) {
	t.Helper()
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic")
		}
	}()
	f()
}

func itoa(v int) string {
	if v == 0 {
		return "0"
	}
	neg := v < 0
	if neg {
		v = -v
	}
	var buf [32]byte
	i := len(buf)
	for v > 0 {
		i--
		buf[i] = byte('0' + v%10)
		v /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
