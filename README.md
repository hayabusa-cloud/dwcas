# dwcas

[![Go Reference](https://pkg.go.dev/badge/github.com/hayabusa-cloud/dwcas.svg)](https://pkg.go.dev/github.com/hayabusa-cloud/dwcas)
[![Go Report Card](https://goreportcard.com/badge/github.com/hayabusa-cloud/dwcas)](https://goreportcard.com/report/github.com/hayabusa-cloud/dwcas)
[![Codecov](https://codecov.io/gh/hayabusa-cloud/dwcas/graph/badge.svg)](https://codecov.io/gh/hayabusa-cloud/dwcas)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

**Languages:** English | [简体中文](README.zh-CN.md) | [日本語](README.ja.md) | [Español](README.es.md) | [Français](README.fr.md)

Portable 128-bit (double-word) compare-and-swap (CAS) primitive for Go.

## At a glance

- Single atomic 128-bit compare-and-swap (CAS) on amd64/arm64.
- Intended for lock-free algorithms (versioned state, ABA mitigation, composite updates).
- Arm64 uses an LSE pair-CAS fast path by default; LL/SC is available via build tag for compatibility.

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
fmt.Println(prev, swapped, v.Lo, v.Hi)
```

## APIs

### Core type

- `type Uint128 struct { Lo uint64; Hi uint64 }`

### Compare-exchange methods

Each method on `*dwcas.Uint128` is a 128-bit compare-and-swap in compare-exchange
form.

All methods return:

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

Ordering is specified per method, with **success ordering** potentially different
from **failure ordering**:

| Method          | Success | Failure |
|-----------------|---|---|
| `dwcas.Relaxed` | relaxed | relaxed |
| `dwcas.Acquire` | acquire | relaxed |
| `dwcas.Release` | release | relaxed |
| `dwcas.AcqRel`  | acq_rel | relaxed |

Notes:

- Some backends are stronger than requested.
  - On `amd64`, `LOCK`ed operations are at least acquire-release on both success
    and failure.
  - On `arm64` default (LSE) builds, `Release`/`AcqRel` place a release barrier
    before the CAS, which also orders a failed attempt.

### Manual barriers

In rare cases, a caller may need an explicit ordering edge outside the CAS128
primitives. `dwcas` provides three manual barriers:

- `dwcas.BarrierAcquire`: arm64 emits `DMB ISHLD`; amd64 is a compiler barrier
  only.
- `dwcas.BarrierRelease`: arm64 emits `DMB ISHST`; amd64 is a compiler barrier
  only.
- `dwcas.BarrierFull`: arm64 emits `DMB ISH`; amd64 is a compiler barrier only.

## Alignment requirement

The address of a `*dwcas.Uint128` used with these methods must be 16-byte aligned.
The pointer must also be non-nil.

- Default builds perform no runtime checks.
- Opt-in debug guard: build with `-tags=dwcasdebug` to panic on nil and misaligned pointers.

Use `dwcas.New` if you need a heap-allocated 16-byte aligned `*Uint128`.

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

- `amd64`: implemented via `CMPXCHG16B`.
- `arm64`:
  - default: LSE pair-CAS (CASP family) (best performance).
  - opt-in LL/SC: `-tags=dwcas_llsc` (portable baseline via `LDAXP`/`STLXP`).

## Safety notes

`dwcas` uses `unsafe` and architecture-specific assembly.

- Keep a `*Uint128` reachable as a Go pointer. Do not convert it to `uintptr` and
  back, and do not store it in untracked memory.
- A copied `Uint128` value does not carry alignment guarantees. Alignment is a
  property of the address you pass to these methods.

## License

MIT — see [LICENSE](./LICENSE).

©2025 Hayabusa Cloud Co., Ltd.
