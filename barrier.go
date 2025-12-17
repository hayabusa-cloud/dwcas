package dwcas

import "code.hybscloud.com/dwcas/internal/arch"

// BarrierAcquire emits an acquire barrier.
//
// This API is intentionally rare and expert-only. Prefer the ordering variants of
// [Uint128] CAS methods when possible.
//
// Semantics by architecture:
//
//   - arm64: emits DMB ISHLD.
//   - amd64: compiler barrier only (prevents compile-time reordering across the
//     call). It is not an MFENCE and is not required for cache coherence.
func BarrierAcquire() {
	arch.BarrierAcquire()
}

// BarrierRelease emits a release barrier.
//
// This API is intentionally rare and expert-only. Prefer the ordering variants of
// [Uint128] CAS methods when possible.
//
// Semantics by architecture:
//
//   - arm64: emits DMB ISHST.
//   - amd64: compiler barrier only (prevents compile-time reordering across the
//     call). It is not an MFENCE and is not required for cache coherence.
func BarrierRelease() {
	arch.BarrierRelease()
}

// BarrierFull emits a full barrier.
//
// This API is intentionally rare and expert-only. Prefer the ordering variants of
// [Uint128] CAS methods when possible.
//
// Semantics by architecture:
//
//   - arm64: emits DMB ISH.
//   - amd64: compiler barrier only (prevents compile-time reordering across the
//     call). It is not an MFENCE and is not required for cache coherence.
func BarrierFull() {
	arch.BarrierFull()
}
