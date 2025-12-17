// ©Hayabusa Cloud Co., Ltd. 2025. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package dwcas

import (
	"unsafe"

	"code.hybscloud.com/dwcas/internal/arch"
)

// Uint128 is a 16-byte value used with 128-bit compare-and-swap.
//
// Layout is stable and contiguous in memory:
//
//	word 0: Lo
//	word 1: Hi
//
// The address of a *Uint128 used with these methods MUST be 16-byte aligned.
//
// Helpers:
//   - For heap allocation with guaranteed 16-byte alignment, use [New].
//   - For placing a *Uint128 into a byte buffer, use [CanPlaceAlignedUint128] and
//     [PlaceAlignedUint128].
type Uint128 struct {
	Lo uint64
	Hi uint64
}

// Relaxed is a 128-bit compare-and-swap (CAS) with relaxed ordering on both
// success and failure.
//
// It always returns:
//   - prev: the value observed in memory at the time of the CAS attempt.
//   - swapped: true if the swap happened.
//
// Contract:
//   - p must be non-nil.
//   - p must be 16-byte aligned.
//   - On unsupported architectures, Relaxed panics.
func (p *Uint128) Relaxed(old, new Uint128) (prev Uint128, swapped bool) {
	checkAligned(p)
	lo, hi, ok := arch.Cas128Relaxed((*uint64)(unsafe.Pointer(p)), old.Lo, old.Hi, new.Lo, new.Hi)
	return Uint128{Lo: lo, Hi: hi}, ok
}

// Acquire is a 128-bit compare-and-swap (CAS) with acquire ordering on success
// and relaxed ordering on failure.
//
// It always returns:
//   - prev: the value observed in memory at the time of the CAS attempt.
//   - swapped: true if the swap happened.
//
// Contract:
//   - p must be non-nil.
//   - p must be 16-byte aligned.
//   - On unsupported architectures, Acquire panics.
func (p *Uint128) Acquire(old, new Uint128) (prev Uint128, swapped bool) {
	checkAligned(p)
	lo, hi, ok := arch.Cas128Acquire((*uint64)(unsafe.Pointer(p)), old.Lo, old.Hi, new.Lo, new.Hi)
	return Uint128{Lo: lo, Hi: hi}, ok
}

// Release is a 128-bit compare-and-swap (CAS) with release ordering on success
// and relaxed ordering on failure.
//
// It always returns:
//   - prev: the value observed in memory at the time of the CAS attempt.
//   - swapped: true if the swap happened.
//
// Contract:
//   - p must be non-nil.
//   - p must be 16-byte aligned.
//   - On unsupported architectures, Release panics.
func (p *Uint128) Release(old, new Uint128) (prev Uint128, swapped bool) {
	checkAligned(p)
	lo, hi, ok := arch.Cas128Release((*uint64)(unsafe.Pointer(p)), old.Lo, old.Hi, new.Lo, new.Hi)
	return Uint128{Lo: lo, Hi: hi}, ok
}

// AcqRel is a 128-bit compare-and-swap (CAS) with acquire-release ordering on
// success and relaxed ordering on failure.
//
// It always returns:
//   - prev: the value observed in memory at the time of the CAS attempt.
//   - swapped: true if the swap happened.
//
// Contract:
//   - p must be non-nil.
//   - p must be 16-byte aligned.
//   - On unsupported architectures, AcqRel panics.
func (p *Uint128) AcqRel(old, new Uint128) (prev Uint128, swapped bool) {
	checkAligned(p)
	lo, hi, ok := arch.Cas128AcqRel((*uint64)(unsafe.Pointer(p)), old.Lo, old.Hi, new.Lo, new.Hi)
	return Uint128{Lo: lo, Hi: hi}, ok
}
