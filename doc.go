// Package dwcas provides a portable 128-bit (double-word) compare-and-swap (CAS)
// primitive for Go.
//
// The core operations are compare-and-swap methods on [Uint128] that perform a
// single atomic read-modify-write on a contiguous 16-byte value.
//
// Intended use cases
//
//   - Lock-free algorithms that need to atomically update a value and a version/tag
//     (e.g. ABA mitigation via versioned pointers).
//   - Composite state machines where two 64-bit words must move together.
//   - Low-level runtime-like data structures (queues, stacks) where a single 128-bit
//     CAS reduces coordination overhead.
//
// # Atomicity and memory ordering
//
// Each method is a single atomic 128-bit compare-and-swap on supported
// architectures.
//
// # Return values
//
// All methods return:
//   - prev: the value observed in memory at the time of the CAS attempt
//   - swapped: true if the swap happened
//
// # Ordering contracts
//
// Success ordering vs failure ordering:
//
//   - Relaxed: success = relaxed, failure = relaxed
//   - Acquire: success = acquire, failure = relaxed
//   - Release: success = release, failure = relaxed
//   - AcqRel:  success = acq_rel, failure = relaxed
//
// Some architectures and backends may provide stronger ordering than requested.
// In particular:
//   - amd64's LOCKed instructions are at least acquire-release for both success
//     and failure.
//   - arm64's default LSE backend uses CASPD plus explicit barriers; release-style
//     methods place the release barrier before the CAS, which also orders a failed
//     attempt.
//
// # Manual barriers
//
// This package also exposes manual barriers ([BarrierAcquire], [BarrierRelease],
// [BarrierFull]) for callers who need an explicit ordering edge outside the CAS
// primitives. On arm64 they map to DMB ISH*; on amd64 they are compiler barriers
// (not MFENCE).
//
// # Alignment
//
// The address of a *Uint128 passed to these methods MUST be 16-byte aligned.
// Misalignment is unsupported and may fault on some CPUs/instructions.
//
// Helpers:
//   - [New] returns a heap-allocated 16-byte aligned *Uint128.
//   - [CanPlaceAlignedUint128] / [PlaceAlignedUint128] place a 16-byte aligned
//     *Uint128 within a caller-provided byte buffer. The worst-case required
//     remaining bytes from off is 31.
//
// This package intentionally does not perform runtime alignment checks in normal
// builds. For a debug-only guard, build with `-tags=dwcasdebug` to make these
// methods panic when called with a misaligned pointer.
//
// Architecture support
//
//   - amd64: implemented via CMPXCHG16B.
//   - arm64: implemented via either LSE pair-CAS (CASP family; default) or
//     LL/SC (opt-in).
//   - other architectures: building succeeds, but all CAS methods panic at runtime.
//
// Arm64 backend selection
//
//   - default: LSE pair-CAS (CASP family; CASPAL semantics)
//   - opt-in: `-tags=dwcas_llsc` (LL/SC via LDXP with STXP or STLXP)
package dwcas
