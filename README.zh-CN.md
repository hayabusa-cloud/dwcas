# dwcas

[![Go Reference](https://pkg.go.dev/badge/code.hybscloud.com/dwcas.svg)](https://pkg.go.dev/code.hybscloud.com/dwcas)
[![Benchmarks](https://github.com/hayabusa-cloud/dwcas/actions/workflows/benchmarks.yml/badge.svg)](https://github.com/hayabusa-cloud/dwcas/actions/workflows/benchmarks.yml)
[![Codecov](https://codecov.io/gh/hayabusa-cloud/dwcas/graph/badge.svg)](https://codecov.io/gh/hayabusa-cloud/dwcas)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

语言： [English](README.md) | **简体中文** | [Español](README.es.md) | [日本語](README.ja.md) | [Français](README.fr.md)

为 Go 提供 128 位（双字）原子加载、存储和比较交换（CAS）操作。

## 一览

- 在 `amd64` / `arm64` 上提供单次原子 128 位 CAS。
- 面向无锁算法（版本化状态、ABA 缓解、复合更新）。
- arm64 默认构建要求 FEAT_LSE2。较早的 CPU 请使用 `-tags=dwcas_llsc`。

## 安装

```bash
go get code.hybscloud.com/dwcas
```

## 示例

### 用 CAS 更新数值对

`dwcas.New` 分配一个按 16 字节对齐的 `Uint128`。`AcqRel` 返回比较时读到的数值对，以及表示交换是否成功的布尔值。

```go
v := dwcas.New(1, 2)

old := dwcas.Uint128{Lo: 1, Hi: 2}
newv := dwcas.Uint128{Lo: 3, Hi: 4}

prev, swapped := v.AcqRel(old, newv)
fmt.Println(prev, swapped, v.LoadRelaxed()) // {1 2} true {3 4}
```

### 发布缓冲区中的一段数据

生产者一次发布一段非空数据的字节偏移量（`Lo`）和字节数（`Hi`）。长度为零表示描述符为空。

```go
buffer := make([]byte, 64)
published := dwcas.New(0, 0)

// 生产者
const offset = 8
n := copy(buffer[offset:], "hello")
published.StoreRelease(dwcas.Uint128{Lo: offset, Hi: uint64(n)})

// 消费者
span := published.LoadAcquire()
if span.Hi != 0 {
	fmt.Printf("%s\n", buffer[span.Lo:span.Lo+span.Hi])
}
```

消费者读到已发布的数值对后，就可以读取生产者写入的字节。消费者读完之前，生产者须保持这段数据的内容不变。

如果两个 goroutine 仅通过这个数值对同步，`-race` 可能会报告缓冲区上的数据竞争，因为读写数值对的汇编代码未做竞态检测插桩。

`published` 的布局由应用程序定义。在 Linux `io_uring` 的提供缓冲区环中，对第一个描述符做一次 128 位写入，会同时写入字节偏移量 14 处的 16 位环尾索引。`uring.BufferRing.Add` 写入地址、长度和缓冲区 ID，`Advance` 更新环尾索引，将缓冲区发布给内核。 [Linux 缓冲区环布局](https://github.com/torvalds/linux/blob/v6.18/include/uapi/linux/io_uring.h#L816-L836)

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

在 arm64 上，这些屏障在内部共享域（Inner Shareable）内对写回型共享内存的常规访问进行排序。

- `dwcas.BarrierAcquire` 保证先前的读取排在后续的读取和写入之前（`DMB ISHLD`）。
- `dwcas.BarrierRelease` 保证先前的写入排在后续的写入之前（`DMB ISHST`）。如需让先前的读取也排在发布之前，请使用释放存储或成功的释放 CAS。
- `dwcas.BarrierFull` 保证先前的读写都排在后续的读写之前（`DMB ISH`）。

在 amd64 上，这三个函数都是编译器屏障。处理器可以把一次读取重排到对另一地址的先前写入之前。

## 对齐要求

接收者必须指向 16 字节对齐的可写内存。在 arm64 默认后端上，加载和存储要求使用 Inner 和 Outer 均为 Write-Back 的 Normal 内存，例如 Go 堆内存。加载或失败的 CAS 也可能写回观察到的值。并发访问这个数值对时，请使用它的原子方法。

默认构建由调用方保证接收者的对齐。使用 `-tags=dwcasdebug` 时，如果接收者为 `nil` 或未按 16 字节对齐，原子方法会触发 panic。

使用 `dwcas.New` 可在 Go 管理的内存中分配按 16 字节对齐的 `*Uint128`。

`CanPlaceAlignedUint128` 检查 `off` 是否非负，以及切片从该偏移量起是否至少还剩 31 字节。若条件不满足，`PlaceAlignedUint128` 会触发 panic。成功时，它返回实际占用的字节数（含对齐填充，共 16 至 31 字节）和对齐后的指针。

```go
buf := make([]byte, 256)
off := 7

n, v := dwcas.PlaceAlignedUint128(buf, off)
*v = dwcas.Uint128{Lo: 1, Hi: 2}

off += n
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
- 如果 Go 对象的地址只保存在 `Lo` 或 `Hi` 中，垃圾回收器仍可回收该对象。队列可以将指针保存在可达的切片中，用这两个字段记录槽位索引和对应的版本号。
- 拷贝出来的 `Uint128` 值不携带对齐保证。对齐是你传入这些方法的地址属性。

## License

MIT — 见 [LICENSE](./LICENSE)。

©2025 Hayabusa Cloud Co., Ltd.
