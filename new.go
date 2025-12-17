package dwcas

import "unsafe"

// New returns a heap-allocated *Uint128 whose address is guaranteed to be 16-byte
// aligned.
//
// This is primarily a convenience for algorithms that require 16-byte alignment
// but do not control allocator/layout details (e.g. when a Uint128 cannot be
// embedded into a manually-aligned struct).
//
// Safety notes:
//   - The returned pointer refers to Go-managed memory. Keep the *Uint128 reachable
//     (do not convert it to uintptr and back).
//   - Alignment is guaranteed, but only for the returned pointer itself. If you copy
//     the value into another location, you must re-establish 16-byte alignment.
//
//go:nocheckptr
func New(lo, hi uint64) *Uint128 {
	// A []uint64 data pointer is always 8-byte aligned, so the only possible
	// 16-byte misalignment is +8. Allocating 3 words guarantees space for an
	// aligned 2-word window.
	mem := make([]uint64, 3)
	base := uintptr(unsafe.Pointer(unsafe.SliceData(mem)))
	pad := (uintptr(16) - (base & uintptr(15))) & uintptr(15) // 0 or 8
	off := pad >> 3                                           // 0 or 1 (words)
	u128 := mem[off : off+2]

	p := (*Uint128)(unsafe.Pointer(unsafe.SliceData(u128)))
	p.Lo = lo
	p.Hi = hi
	return p
}

// CanPlaceAlignedUint128 reports whether p has enough remaining capacity from off
// to place a 16-byte aligned *Uint128 at or after p[off].
//
// Worst-case required remaining bytes from off is 31:
//   - up to 15 bytes of padding to reach a 16-byte boundary
//   - 16 bytes for the Uint128 itself
func CanPlaceAlignedUint128(p []byte, off int) bool {
	if off < 0 || off > len(p) {
		return false
	}
	return len(p)-off >= 31
}

// PlaceAlignedUint128 returns a *Uint128 placed within p, starting at or after
// p[off], such that the returned address is 16-byte aligned.
//
// It returns n, the number of bytes the caller must "consume" starting from off
// to cover the aligned 16-byte region (including any alignment padding).
//
// The caller is responsible for ensuring [CanPlaceAlignedUint128] is true.
// If not, PlaceAlignedUint128 panics.
//
// Safety notes:
//   - The returned pointer refers to p's backing array. Keep p reachable.
//   - Do not convert the returned pointer to uintptr and back.
//
//go:nocheckptr
func PlaceAlignedUint128(p []byte, off int) (n int, u128 *Uint128) {
	if !CanPlaceAlignedUint128(p, off) {
		panic("dwcas: PlaceAlignedUint128: insufficient space")
	}

	base := uintptr(unsafe.Pointer(unsafe.SliceData(p)))
	start := base + uintptr(off)

	aligned := (start + 15) &^ uintptr(15) // round up to 16-byte boundary
	pad := int(aligned - start)            // 0..15
	n = pad + 16                           // 16..31

	u128 = (*Uint128)(unsafe.Pointer(aligned))
	return n, u128
}
