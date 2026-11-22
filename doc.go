// Package dwcas provides 128-bit (double-word) atomic loads, stores, and
// compare-and-swap (CAS) on [Uint128].
//
// Intended use cases
//
//   - Lock-free algorithms that pair a slot index with a generation counter
//     to mitigate ABA when a slot is reused.
//   - Composite state machines where two 64-bit words must move together.
//   - Low-level runtime-like data structures (queues, stacks) where a single 128-bit
//     CAS reduces coordination overhead.
//
// # Atomicity and memory ordering
//
// Each method observes or updates both words as one atomic operation.
// [Uint128.LoadRelaxed] and [Uint128.LoadAcquire] return a coherent snapshot.
// [Uint128.StoreRelaxed] and [Uint128.StoreRelease] replace the entire pair.
// When an acquire load observes a release store, accesses before the store
// are ordered before accesses after the load.
//
// # Return values
//
// [Uint128.TryRelaxed], [Uint128.TryAcquire], [Uint128.TryRelease], and
// [Uint128.TryAcqRel] return true on a swap and false only on a mismatch. The four
// compare-exchange methods ([Uint128.Relaxed], [Uint128.Acquire],
// [Uint128.Release], and [Uint128.AcqRel]) return:
//   - prev: the value observed in memory at the time of the CAS attempt
//   - swapped: true if the swap happened
//
// # Ordering contracts
//
// Success ordering vs failure ordering:
//
//   - Relaxed and TryRelaxed: success = relaxed, failure = relaxed
//   - Acquire and TryAcquire: success = acquire, failure = relaxed
//   - Release and TryRelease: success = release, failure = relaxed
//   - AcqRel and TryAcqRel: success = acq_rel, failure = relaxed
//
// Some architectures and backends may provide stronger ordering than requested.
// In particular:
//   - amd64's LOCKed instructions are at least acquire-release for both success
//     and failure.
//   - On arm64, CAS methods with acquire ordering also acquire on a mismatch.
//
// # Manual barriers
//
// A ring producer can call [BarrierRelease] after writing descriptors and
// before storing the new tail. When returning a buffer for reuse, the producer
// can order both its payload reads and descriptor writes with one release
// store to the tail.
//
// On arm64, [BarrierAcquire], [BarrierRelease], and [BarrierFull] apply to
// the inner-shareable domain. On amd64 all three are compiler barriers.
// They panic on other architectures.
//
// # Alignment
//
// The receiver must point to writable, 16-byte-aligned memory. Default arm64
// loads and stores require Normal memory with Inner and Outer Write-Back
// cacheability, as in the Go heap. Some backends write the observed pair back
// even for a load or a failed CAS.
//
// [New] allocates an aligned pair in Go-managed memory. In a caller-provided
// byte slice, [PlaceAlignedUint128] requires at least 31 bytes starting at a
// valid offset and consumes 16 to 31 bytes, including padding. It panics if
// [CanPlaceAlignedUint128] would return false.
//
// Normal builds leave receiver alignment to the caller. With the dwcasdebug
// build tag, the atomic methods panic on a nil or misaligned receiver.
//
// # Architecture support
//
//   - amd64: CAS uses LOCK CMPXCHG16B. By default, loads use the same instruction
//     and stores retry it on contention.
//   - arm64: default builds target FEAT_LSE2. Relaxed loads and stores use
//     LDP or STP on the receiver. Acquire adds DMB ISHLD after LDP and release
//     adds DMB ISH before STP. CAS uses CASP, CASPA, CASPL, or CASPAL.
//   - Other architectures: atomic methods and manual barriers panic.
//
// On amd64, -tags=dwcas_avx selects one aligned 128-bit VMOVDQA on the receiver
// per standalone load or store. This requires OS-enabled AVX and a CPU that
// guarantees atomic aligned 16-byte vector accesses to write-back memory.
//
// On arm64 without FEAT_LSE2, build with -tags=dwcas_llsc. This backend uses
// exclusive pairs. A load validates its snapshot with a successful write-back.
// A lost reservation can cause unbounded retries. A Try method returns false
// as soon as either word differs.
package dwcas
