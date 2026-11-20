// ©Hayabusa Cloud Co., Ltd. 2025. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package dwcas

import "code.hybscloud.com/dwcas/internal/arch"

// Uint128 is a pair of 64-bit words that can be loaded, stored, or compared
// and swapped atomically.
//
// Layout is stable and contiguous in memory:
//
//	word 0: Lo
//	word 1: Hi
//
// Method receivers must point to writable, 16-byte-aligned memory. On the
// default arm64 backend, loads and stores require Normal memory with Inner
// and Outer Write-Back cacheability, as in the Go heap. The amd64 dwcas_avx
// backend also requires write-back memory. A load or failed
// comparison may write back the value it observes. Use the methods for
// concurrent access to the pair.
//
// Helpers:
//   - [New] allocates a 16-byte-aligned pair.
//   - For placing a *Uint128 into a byte buffer, use [CanPlaceAlignedUint128] and
//     [PlaceAlignedUint128].
type Uint128 struct {
	Lo uint64
	Hi uint64
}

// LoadAcquire atomically loads the pair with acquire ordering.
func (p *Uint128) LoadAcquire() Uint128 {
	checkAligned(p)
	return arch.LoadAcquire((*arch.Pair)(p))
}

// StoreRelease atomically stores value with release ordering.
func (p *Uint128) StoreRelease(value Uint128) {
	checkAligned(p)
	arch.StoreRelease((*arch.Pair)(p), value)
}

// TryRelaxed atomically replaces old with new and reports whether they matched.
// Success and failure have relaxed ordering.
func (p *Uint128) TryRelaxed(old, new Uint128) bool {
	checkAligned(p)
	return arch.TryRelaxed((*arch.Pair)(p), old, new)
}

// TryAcquire atomically replaces old with new and reports whether they matched.
// Success has acquire ordering and failure has relaxed ordering.
func (p *Uint128) TryAcquire(old, new Uint128) bool {
	checkAligned(p)
	return arch.TryAcquire((*arch.Pair)(p), old, new)
}

// TryRelease atomically replaces old with new and reports whether they matched.
// Success has release ordering and failure has relaxed ordering.
func (p *Uint128) TryRelease(old, new Uint128) bool {
	checkAligned(p)
	return arch.TryRelease((*arch.Pair)(p), old, new)
}

// TryAcqRel atomically replaces old with new and reports whether they matched.
// Success has acquire-release ordering and failure has relaxed ordering.
func (p *Uint128) TryAcqRel(old, new Uint128) bool {
	checkAligned(p)
	return arch.TryAcqRel((*arch.Pair)(p), old, new)
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
	prev, swapped = arch.Cas128Relaxed((*arch.Pair)(p), old, new)
	return
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
	prev, swapped = arch.Cas128Acquire((*arch.Pair)(p), old, new)
	return
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
	prev, swapped = arch.Cas128Release((*arch.Pair)(p), old, new)
	return
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
	prev, swapped = arch.Cas128AcqRel((*arch.Pair)(p), old, new)
	return
}
