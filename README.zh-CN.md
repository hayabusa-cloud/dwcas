# dwcas

[![Go Reference](https://pkg.go.dev/badge/code.hybscloud.com/dwcas.svg)](https://pkg.go.dev/code.hybscloud.com/dwcas)
[![Go Report Card](https://goreportcard.com/badge/code.hybscloud.com/dwcas)](https://goreportcard.com/report/code.hybscloud.com/dwcas)
[![Codecov](https://codecov.io/gh/hayabusa-cloud/dwcas/graph/badge.svg)](https://codecov.io/gh/hayabusa-cloud/dwcas)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

语言： [English](README.md) | **简体中文** | [日本語](README.ja.md) | [Español](README.es.md) | [Français](README.fr.md)

为 Go 提供 128 位（双字）原子加载、存储和比较交换（CAS）操作。

## 一览

- 在 `amd64` / `arm64` 上提供单次原子 128 位 CAS。
- 面向无锁算法（版本化状态、ABA 缓解、复合更新）。
- arm64 默认构建要求 FEAT_LSE2。较早的 CPU 请使用 `-tags=dwcas_llsc`。

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
fmt.Println(prev, swapped, v.LoadRelaxed())
```

## API

### 核心类型

- `type Uint128 struct { Lo uint64; Hi uint64 }`

### 加载与存储

- `LoadRelaxed() Uint128` 和 `LoadAcquire() Uint128` 返回同一时刻的完整双字值。
- `StoreRelaxed(value Uint128)` 和 `StoreRelease(value Uint128)` 以一次原子操作替换两个字。

### CAS 方法

四个 `Try` 方法接收 `(old, new Uint128)` 并返回 `bool`。`false` 表示观察到的值与 `old` 不同。arm64 LL/SC 后端在值匹配但独占预留失效时重试，因此单次调用的步数没有固定上限。

四个 compare-exchange 方法还会返回观测到的完整值对。

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

CAS 方法分别指定成功和失败时的内存序。

| 方法 | 成功 | 失败 |
|------|------|------|
| `(*Uint128).Relaxed`, `(*Uint128).TryRelaxed` | relaxed | relaxed |
| `(*Uint128).Acquire`, `(*Uint128).TryAcquire` | acquire | relaxed |
| `(*Uint128).Release`, `(*Uint128).TryRelease` | release | relaxed |
| `(*Uint128).AcqRel`, `(*Uint128).TryAcqRel`  | acq_rel | relaxed |

说明：

- 某些后端可能比请求的更强。
  - 在 `amd64` 上，带 `LOCK` 的操作在成功与失败上至少是 acquire-release。
  - 在 arm64 上，具有 acquire 语义的 CAS 方法在比较失败时也具有 acquire 语义。

### 手动屏障（manual barriers）

少数情况下，调用方可能需要在 CAS128 原语之外，显式建立额外的有序边。`dwcas` 提供三个手动屏障：

- `dwcas.BarrierAcquire`：`arm64` 生成 `DMB ISHLD`；`amd64` 仅为编译器屏障。
- `dwcas.BarrierRelease`：`arm64` 生成 `DMB ISHST`；`amd64` 仅为编译器屏障。
- `dwcas.BarrierFull`：`arm64` 生成 `DMB ISH`；`amd64` 仅为编译器屏障。

## 对齐要求

接收者必须指向 16 字节对齐的可写内存。在 arm64 默认后端上，加载和存储要求使用 Inner 和 Outer 均为 Write-Back 的 Normal 内存，例如 Go 堆内存。加载或失败的 CAS 也可能写回观察到的值。

- 默认构建不做运行时检查。
- 可选调试保护：使用 `-tags=dwcasdebug` 构建，在 nil 或未对齐时触发 panic。

使用 `dwcas.New` 可在 Go 管理的内存中分配按 16 字节对齐的 `*Uint128`。

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

- 在 `amd64` 上，CAS 和默认的加载操作使用 `LOCK CMPXCHG16B`。默认的存储操作在比较失败时重试。
  - 对于读密集型负载，使用 `-tags=dwcas_avx` 后，每次独立加载或存储只需一条对齐的 `VMOVDQA` 指令访问目标。操作系统须启用 AVX，且 CPU 须保证对写回内存中按 16 字节对齐的 16 字节向量访问具有原子性。
- `arm64`：
  - 宽松加载和存储使用 `LDP` 和 `STP`。获取加载在 `LDP` 后执行 `DMB ISHLD`，释放存储在 `STP` 前执行 `DMB ISH`。CAS 使用 LSE 的 `CASP` 系列指令。
  - 可选 LL/SC：`-tags=dwcas_llsc`（通过 `LDXP`/`LDAXP` 配合 `STXP` 或 `STLXP` 的可移植基线）。

## 安全注意事项

`dwcas` 使用 `unsafe` 与架构相关汇编。

- 让 `*Uint128` 作为 Go 指针保持可达（reachable）。不要把它转换为 `uintptr` 再转换回来，也不要把它存放在未被追踪的内存里。
- 拷贝出来的 `Uint128` 值不携带对齐保证。对齐是你传入这些方法的地址属性。

## License

MIT — 见 [LICENSE](./LICENSE)。

©2025 Hayabusa Cloud Co., Ltd.
