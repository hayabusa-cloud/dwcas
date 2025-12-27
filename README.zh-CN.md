# dwcas

[![Go Reference](https://pkg.go.dev/badge/github.com/hayabusa-cloud/dwcas.svg)](https://pkg.go.dev/github.com/hayabusa-cloud/dwcas)
[![Go Report Card](https://goreportcard.com/badge/github.com/hayabusa-cloud/dwcas)](https://goreportcard.com/report/github.com/hayabusa-cloud/dwcas)
[![Codecov](https://codecov.io/gh/hayabusa-cloud/dwcas/graph/badge.svg)](https://codecov.io/gh/hayabusa-cloud/dwcas)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

语言： [English](README.md) | **简体中文** | [日本語](README.ja.md) | [Español](README.es.md) | [Français](README.fr.md)

为 Go 提供可移植的 128 位（双字）比较并交换（CAS）原语。

## 一览

- 在 `amd64` / `arm64` 上提供单次原子 128 位 CAS。
- 面向无锁算法（版本化状态、ABA 缓解、复合更新）。
- `arm64` 默认使用 LSE pair-CAS 快路径；也提供 LL/SC（通过 build tag）用于兼容性。

## 安装

```bash
go get code.hybscloud.com/dwcas
```

## 快速开始

```text
// New returns a 16-byte aligned *Uint128 suitable for 128-bit compare-and-swap.
v := dwcas.New(1, 2)

old := dwcas.Uint128{Lo: 1, Hi: 2}
newv := dwcas.Uint128{Lo: 3, Hi: 4}

prev, swapped := v.AcqRel(old, newv)
fmt.Println(prev, swapped, v.Lo, v.Hi)
```

## API

### 核心类型

- `type Uint128 struct { Lo uint64; Hi uint64 }`

### Compare-exchange 方法

`*dwcas.Uint128` 上的每个方法，都是以 compare-exchange 形式提供的 128 位 CAS。

所有方法返回：

- `prev`：本次 CAS 尝试时，在内存中观察到的值
- `swapped`：若发生交换则为 `true`

方法：

- `(*Uint128) Relaxed(old, new Uint128) (prev Uint128, swapped bool)`
- `(*Uint128) Acquire(old, new Uint128) (prev Uint128, swapped bool)`
- `(*Uint128) Release(old, new Uint128) (prev Uint128, swapped bool)`
- `(*Uint128) AcqRel(old, new Uint128) (prev Uint128, swapped bool)`

### 分配与放置（placement）

- `func New(lo, hi uint64) *Uint128`
- `func CanPlaceAlignedUint128(p []byte, off int) bool`
- `func PlaceAlignedUint128(p []byte, off int) (n int, u128 *Uint128)`

## 内存序（memory ordering）

每个方法都指定内存序；并且 **成功序** 与 **失败序** 可能不同：

| 方法 | 成功 | 失败 |
|------|------|------|
| `dwcas.Relaxed` | relaxed | relaxed |
| `dwcas.Acquire` | acquire | relaxed |
| `dwcas.Release` | release | relaxed |
| `dwcas.AcqRel`  | acq_rel | relaxed |

说明：

- 某些后端可能比请求的更强。
  - 在 `amd64` 上，带 `LOCK` 的操作在成功与失败上至少是 acquire-release。
  - 在 `arm64` 默认（LSE）构建中，`Release`/`AcqRel` 会在 CAS 前放置 release barrier，这也会对失败尝试产生排序效果。

### 手动屏障（manual barriers）

少数情况下，调用方可能需要在 CAS128 原语之外，显式建立额外的有序边。`dwcas` 提供三个手动屏障：

- `dwcas.BarrierAcquire`：`arm64` 生成 `DMB ISHLD`；`amd64` 仅为编译器屏障。
- `dwcas.BarrierRelease`：`arm64` 生成 `DMB ISHST`；`amd64` 仅为编译器屏障。
- `dwcas.BarrierFull`：`arm64` 生成 `DMB ISH`；`amd64` 仅为编译器屏障。

## 对齐要求

对这些方法使用的 `*dwcas.Uint128` 地址必须 16 字节对齐，并且指针必须非 nil。

- 默认构建不做运行时检查。
- 可选调试保护：使用 `-tags=dwcasdebug` 构建，在 nil 或未对齐时触发 panic。

如果你需要一个堆分配且 16 字节对齐的 `*Uint128`，请使用 `dwcas.New`。

如果你需要把 `*Uint128` 放在调用方提供的 byte buffer 中（例如手写 arena），请使用 placement helpers。从 `off` 开始的最坏剩余字节需求为 31。

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

## 支持的架构与后端

- `amd64`：通过 `CMPXCHG16B` 实现。
- `arm64`：
  - 默认：LSE pair-CAS（CASP family）（最佳性能）。
  - 可选 LL/SC：`-tags=dwcas_llsc`（通过 `LDXP` 配合 `STXP` 或 `STLXP` 的可移植基线）。

## 安全注意事项

`dwcas` 使用 `unsafe` 与架构相关汇编。

- 让 `*Uint128` 作为 Go 指针保持可达（reachable）。不要把它转换为 `uintptr` 再转换回来，也不要把它存放在未被追踪的内存里。
- 拷贝出来的 `Uint128` 值不携带对齐保证。对齐是你传入这些方法的地址属性。

## License

MIT — 见 [LICENSE](./LICENSE)。

©2025 Hayabusa Cloud Co., Ltd.
