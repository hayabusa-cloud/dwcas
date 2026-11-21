# dwcas

[![Go Reference](https://pkg.go.dev/badge/code.hybscloud.com/dwcas.svg)](https://pkg.go.dev/code.hybscloud.com/dwcas)
[![Go Report Card](https://goreportcard.com/badge/code.hybscloud.com/dwcas)](https://goreportcard.com/report/code.hybscloud.com/dwcas)
[![Codecov](https://codecov.io/gh/hayabusa-cloud/dwcas/graph/badge.svg)](https://codecov.io/gh/hayabusa-cloud/dwcas)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

**Languages:** English | [简体中文](README.zh-CN.md) | [日本語](README.ja.md) | [Español](README.es.md) | [Français](README.fr.md)

128-bit (double-word) atomic loads, stores, and compare-and-swap (CAS) for Go.

## At a glance

- Single atomic 128-bit compare-and-swap (CAS) on amd64/arm64.
- Intended for lock-free algorithms (versioned state, ABA mitigation, composite updates).
- Default arm64 builds target FEAT_LSE2 for atomic pair loads and stores. Use `-tags=dwcas_llsc` on older CPUs.

## Install

```bash
go get code.hybscloud.com/dwcas
```

## Quick start

```text
// New returns a 16-byte aligned *Uint128 suitable for 128-bit compare-and-swap.
v := dwcas.New(1, 2)

old := dwcas.Uint128{Lo: 1, Hi: 2}
newv := dwcas.Uint128{Lo: 3, Hi: 4}

prev, swapped := v.AcqRel(old, newv)
fmt.Println(prev, swapped, v.LoadRelaxed())
```

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
  - On `amd64`, `LOCK`ed operations are at least acquire-release on both success
    and failure.
  - On arm64, CAS methods with acquire ordering also acquire on a mismatch.

### Manual barriers

In rare cases, a caller may need an explicit ordering edge outside the CAS128
primitives. `dwcas` provides three manual barriers:

- `dwcas.BarrierAcquire`: arm64 emits `DMB ISHLD`; amd64 is a compiler barrier
  only.
- `dwcas.BarrierRelease` orders earlier stores before later stores for ordinary
  accesses to write-back shared memory. It uses `DMB ISHST` on arm64 and a
  compiler barrier on amd64. To order earlier loads as well, publish with a
  release store or a successful release CAS.
- `dwcas.BarrierFull`: arm64 emits `DMB ISH`; amd64 is a compiler barrier only.

## Alignment requirement

The receiver must point to writable, 16-byte-aligned memory. Default arm64 loads and stores require Normal memory with Inner and Outer Write-Back cacheability, as in the Go heap. A load or failed CAS may write back the observed value. Use the atomic methods for concurrent access.

- Default builds perform no runtime checks.
- Opt-in debug guard: build with `-tags=dwcasdebug` to panic on nil and misaligned pointers.

Use `dwcas.New` to allocate a 16-byte-aligned `*Uint128` in Go-managed memory.

If you need to place a `*Uint128` inside a caller-provided byte buffer (for
example, in a manually managed arena), use the placement helpers. Worst-case
required remaining bytes from `off` is 31.

```text
buf := make([]byte, 256)
off := 7

if !dwcas.CanPlaceAlignedUint128(buf, off) {
	panic("insufficient space")
}

n, v := dwcas.PlaceAlignedUint128(buf, off)
*v = dwcas.Uint128{Lo: 1, Hi: 2}

_, _ = v.Relaxed(dwcas.Uint128{Lo: 1, Hi: 2}, dwcas.Uint128{Lo: 3, Hi: 4})

off += n // n is in [16..31]
```

## Supported architectures and backends

- On `amd64`, CAS and default pair loads use `LOCK CMPXCHG16B`. Default stores retry on a mismatch.
  - For read-heavy workloads, `-tags=dwcas_avx` uses one aligned `VMOVDQA` per standalone load or store. The OS must enable AVX and the CPU must guarantee atomic aligned 16-byte vector accesses to write-back memory.
- `arm64`:
  - Default builds use `LDP` and `STP` for relaxed pair access. Acquire loads add `DMB ISHLD` after the load. Release stores add `DMB ISH` before the store. CAS uses the LSE `CASP` family.
  - opt-in LL/SC: `-tags=dwcas_llsc` (portable baseline via `LDXP`/`LDAXP` with `STXP` or `STLXP`).

## Safety notes

`dwcas` uses `unsafe` and architecture-specific assembly.

- Keep a `*Uint128` reachable as a Go pointer. Do not convert it to `uintptr` and
  back, and do not store it in untracked memory.
- The garbage collector can reclaim a Go object whose address is stored only in `Lo` or `Hi`. A queue can keep pointers in a reachable slice and store a slot index and generation in the pair.
- A copied `Uint128` value does not carry alignment guarantees. Alignment is a
  property of the address you pass to these methods.

## License

MIT — see [LICENSE](./LICENSE).

©2025 Hayabusa Cloud Co., Ltd.
