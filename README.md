# dwcas

[![Go Reference](https://pkg.go.dev/badge/code.hybscloud.com/dwcas.svg)](https://pkg.go.dev/code.hybscloud.com/dwcas)
[![Codecov](https://codecov.io/gh/hayabusa-cloud/dwcas/graph/badge.svg)](https://codecov.io/gh/hayabusa-cloud/dwcas)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

**Languages:** English | [简体中文](README.zh-CN.md) | [Español](README.es.md) | [日本語](README.ja.md) | [Français](README.fr.md)

128-bit (double-word) atomic loads, stores, and compare-and-swap (CAS) for Go.

## At a glance

- Single atomic 128-bit compare-and-swap (CAS) on amd64/arm64.
- Intended for lock-free algorithms (versioned state, ABA mitigation, composite updates).
- Default arm64 builds target FEAT_LSE2 for atomic pair loads and stores. Use `-tags=dwcas_llsc` on older CPUs.

## Install

```bash
go get code.hybscloud.com/dwcas
```

## Examples

### Compare and swap a pair

`dwcas.New` allocates a 16-byte-aligned pair. `AcqRel` returns the pair it observed and whether it replaced it.

```go
v := dwcas.New(1, 2)

old := dwcas.Uint128{Lo: 1, Hi: 2}
newv := dwcas.Uint128{Lo: 3, Hi: 4}

prev, swapped := v.AcqRel(old, newv)
fmt.Println(prev, swapped, v.LoadRelaxed()) // {1 2} true {3 4}
```

### Publish a buffer range

A producer publishes the byte offset (`Lo`) and length (`Hi`) of a nonempty payload together. A zero length marks an empty descriptor.

```go
buffer := make([]byte, 64)
published := dwcas.New(0, 0)

// Producer
const offset = 8
n := copy(buffer[offset:], "hello")
published.StoreRelease(dwcas.Uint128{Lo: offset, Hi: uint64(n)})

// Reader
span := published.LoadAcquire()
if span.Hi != 0 {
	fmt.Printf("%s\n", buffer[span.Lo:span.Lo+span.Hi])
}
```

When the reader observes the published pair, it can read the initialized bytes. The producer keeps that range unchanged until the reader finishes.

The application defines the layout of `published`. In a Linux `io_uring` provided-buffer ring, a 128-bit store to the first descriptor would also write the 16-bit ring tail at byte offset 14. `uring.BufferRing.Add` writes the address, length, and buffer ID. `Advance` updates the tail to publish the buffers. [Linux buffer-ring layout](https://github.com/torvalds/linux/blob/v6.18/include/uapi/linux/io_uring.h#L816-L836)

## APIs

### Core type

- `type Uint128 struct { Lo uint64; Hi uint64 }`

### Loads and stores

- `LoadRelaxed() Uint128` and `LoadAcquire() Uint128` return one coherent pair.
- `StoreRelaxed(value Uint128)` and `StoreRelease(value Uint128)` replace both words atomically.

### Compare-and-swap methods

The four `Try` methods take `(old, new Uint128)` and return `bool`. A false result means the value differed from `old`. On arm64 LL/SC, a matching comparison retries if the reservation is lost, so a call has no fixed step bound.

The four compare-exchange methods also return the observed pair.

- `prev`: the value observed in memory at the time of the CAS attempt
- `swapped`: `true` if the swap happened

Methods:

- `(*Uint128) Relaxed(old, new Uint128) (prev Uint128, swapped bool)`
- `(*Uint128) Acquire(old, new Uint128) (prev Uint128, swapped bool)`
- `(*Uint128) Release(old, new Uint128) (prev Uint128, swapped bool)`
- `(*Uint128) AcqRel(old, new Uint128) (prev Uint128, swapped bool)`

### Allocation and placement

- `func New(lo, hi uint64) *Uint128`
- `func CanPlaceAlignedUint128(p []byte, off int) bool`
- `func PlaceAlignedUint128(p []byte, off int) (n int, u128 *Uint128)`

## Memory ordering

The CAS methods specify separate success and failure orderings.

| Method          | Success | Failure |
|-----------------|---|---|
| `(*Uint128).Relaxed`, `(*Uint128).TryRelaxed` | relaxed | relaxed |
| `(*Uint128).Acquire`, `(*Uint128).TryAcquire` | acquire | relaxed |
| `(*Uint128).Release`, `(*Uint128).TryRelease` | release | relaxed |
| `(*Uint128).AcqRel`, `(*Uint128).TryAcqRel`  | acq_rel | relaxed |

Notes:

- Some backends are stronger than requested.
  - On `amd64`, `LOCK`ed operations are at least acquire-release on both success and failure.
  - On arm64, CAS methods with acquire ordering also acquire on a mismatch.

### Manual barriers

On arm64, these barriers order ordinary accesses to write-back shared memory within the inner-shareable domain.

- `dwcas.BarrierAcquire` orders earlier loads before later loads and stores (`DMB ISHLD`).
- `dwcas.BarrierRelease` orders earlier stores before later stores (`DMB ISHST`). To order earlier reads as well, publish with a release store or a successful release CAS.
- `dwcas.BarrierFull` orders earlier loads and stores before later loads and stores (`DMB ISH`).

On amd64 all three act as compiler barriers. The CPU may reorder a load with an earlier store to a different address.

## Alignment requirement

The receiver must point to writable, 16-byte-aligned memory. Default arm64 loads and stores require Normal memory with Inner and Outer Write-Back cacheability, as in the Go heap. A load or failed CAS may write back the observed value. Use the atomic methods for concurrent access.

Default builds leave receiver alignment to the caller. With `-tags=dwcasdebug`, the atomic methods panic on a nil or misaligned receiver.

Use `dwcas.New` to allocate a 16-byte-aligned `*Uint128` in Go-managed memory.

`CanPlaceAlignedUint128` reports whether `off` is nonnegative and at least 31 bytes remain in the slice from that offset. `PlaceAlignedUint128` panics if the check fails. On success it returns the byte count (16 to 31, including padding) and the aligned pointer.

```go
buf := make([]byte, 256)
off := 7

n, v := dwcas.PlaceAlignedUint128(buf, off)
*v = dwcas.Uint128{Lo: 1, Hi: 2}

off += n
```

## Supported architectures and backends

- On `amd64`, CAS and default pair loads use `LOCK CMPXCHG16B`. Default stores retry on a mismatch.
  - For read-heavy workloads, `-tags=dwcas_avx` uses one aligned `VMOVDQA` per standalone load or store. The OS must enable AVX and the CPU must guarantee atomic aligned 16-byte vector accesses to write-back memory.
- `arm64`:
  - Default builds use `LDP` and `STP` for relaxed pair access. Acquire loads add `DMB ISHLD` after the load. Release stores add `DMB ISH` before the store. CAS uses the LSE `CASP` family.
  - opt-in LL/SC: `-tags=dwcas_llsc` (portable baseline via `LDXP`/`LDAXP` with `STXP` or `STLXP`).

## Safety notes

`dwcas` uses `unsafe` and architecture-specific assembly.

- Keep a `*Uint128` reachable as a Go pointer. Do not convert it to `uintptr` and back, and do not store it in untracked memory.
- The garbage collector can reclaim a Go object whose address is stored only in `Lo` or `Hi`. A queue can keep pointers in a reachable slice and store a slot index and generation in the pair.
- A copied `Uint128` value does not carry alignment guarantees. Alignment is a property of the address you pass to these methods.

## License

MIT — see [LICENSE](./LICENSE).

©2025 Hayabusa Cloud Co., Ltd.
