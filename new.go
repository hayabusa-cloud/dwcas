package dwcas

import "unsafe"

// New returns a 16-byte-aligned *Uint128 initialized with lo and hi.
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
func New(lo, hi uint64) *Uint128 {
	// The backing object has Uint128's natural alignment. At most
	// 16 - Alignof(Uint128{}) bytes precede the next 16-byte boundary.
	var storage struct {
		_ [16 - unsafe.Alignof(Uint128{})]byte
		Uint128
	}
	_, p := placeAligned(unsafe.Pointer(&storage))
	*p = Uint128{Lo: lo, Hi: hi}
	return p
}

// CanPlaceAlignedUint128 reports whether p has enough room to place a
// 16-byte aligned [Uint128] at or after p[off].
//
// The offset must be valid, with at least 31 bytes left in the slice.
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
func PlaceAlignedUint128(p []byte, off int) (n int, u128 *Uint128) {
	if !CanPlaceAlignedUint128(p, off) {
		panic("dwcas: PlaceAlignedUint128: insufficient space")
	}

	return placeAligned(unsafe.Add(unsafe.Pointer(unsafe.SliceData(p)), off))
}

func placeAligned(start unsafe.Pointer) (int, *Uint128) {
	pad := -uintptr(start) & 15
	return int(pad) + 16, (*Uint128)(unsafe.Add(start, pad))
}
