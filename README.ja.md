# dwcas

[![Go Reference](https://pkg.go.dev/badge/github.com/hayabusa-cloud/dwcas.svg)](https://pkg.go.dev/github.com/hayabusa-cloud/dwcas)
[![Go Report Card](https://goreportcard.com/badge/github.com/hayabusa-cloud/dwcas)](https://goreportcard.com/report/github.com/hayabusa-cloud/dwcas)
[![Codecov](https://codecov.io/gh/hayabusa-cloud/dwcas/graph/badge.svg)](https://codecov.io/gh/hayabusa-cloud/dwcas)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

言語: [English](README.md) | [简体中文](README.zh-CN.md) | **日本語** | [Español](README.es.md) | [Français](README.fr.md)

Go 向けのポータブルな 128-bit（double-word）compare-and-swap（CAS）プリミティブです。

## 概要

- `amd64` / `arm64` で単一の原子的 128-bit CAS を提供します。
- ロックフリーアルゴリズム（バージョン付き状態、ABA 緩和、複合更新）向けです。
- `arm64` は既定で LSE pair-CAS の高速パスを使用します。互換性のために LL/SC も build tag で選択できます。

## インストール

```bash
go get code.hybscloud.com/dwcas
```

## クイックスタート

```text
// New returns a 16-byte aligned *Uint128 suitable for 128-bit compare-and-swap.
v := dwcas.New(1, 2)

old := dwcas.Uint128{Lo: 1, Hi: 2}
newv := dwcas.Uint128{Lo: 3, Hi: 4}

prev, swapped := v.AcqRel(old, newv)
fmt.Println(prev, swapped, v.Lo, v.Hi)
```

## API

### コア型

- `type Uint128 struct { Lo uint64; Hi uint64 }`

### Compare-exchange メソッド

`*dwcas.Uint128` の各メソッドは compare-exchange 形式の 128-bit CAS です。

全メソッドの返り値:

- `prev`: この CAS 試行時点でメモリ上に観測された値
- `swapped`: 交換が発生した場合 `true`

メソッド:

- `(*Uint128) Relaxed(old, new Uint128) (prev Uint128, swapped bool)`
- `(*Uint128) Acquire(old, new Uint128) (prev Uint128, swapped bool)`
- `(*Uint128) Release(old, new Uint128) (prev Uint128, swapped bool)`
- `(*Uint128) AcqRel(old, new Uint128) (prev Uint128, swapped bool)`

### 割り当てと配置（placement）

- `func New(lo, hi uint64) *Uint128`
- `func CanPlaceAlignedUint128(p []byte, off int) bool`
- `func PlaceAlignedUint128(p []byte, off int) (n int, u128 *Uint128)`

## メモリオーダー

オーダーはメソッドごとに指定され、**成功時**と**失敗時**で異なる場合があります。

| メソッド | 成功 | 失敗 |
|----------|------|------|
| `dwcas.Relaxed` | relaxed | relaxed |
| `dwcas.Acquire` | acquire | relaxed |
| `dwcas.Release` | release | relaxed |
| `dwcas.AcqRel`  | acq_rel | relaxed |

補足:

- 一部のバックエンドは要求より強いオーダーを提供します。
  - `amd64` では `LOCK` 付き命令は成功/失敗ともに少なくとも acquire-release です。
  - `arm64` の既定（LSE）ビルドでは、`Release`/`AcqRel` は CAS の前に release barrier を置き、失敗試行も順序付けます。

### 手動バリア

まれに、CAS128 の外側で明示的な順序付けが必要になることがあります。`dwcas` は 3 つの手動バリアを提供します。

- `dwcas.BarrierAcquire`: `arm64` は `DMB ISHLD`、`amd64` はコンパイラバリアのみ。
- `dwcas.BarrierRelease`: `arm64` は `DMB ISHST`、`amd64` はコンパイラバリアのみ。
- `dwcas.BarrierFull`: `arm64` は `DMB ISH`、`amd64` はコンパイラバリアのみ。

## アラインメント要件

これらのメソッドで使用する `*dwcas.Uint128` のアドレスは 16 バイト境界にアラインされている必要があります。ポインタは non-nil である必要があります。

- 既定ビルドではランタイムチェックを行いません。
- デバッグ用ガード: `-tags=dwcasdebug` でビルドすると、nil または未アラインで panic します。

ヒープ上に 16 バイトアラインされた `*Uint128` が必要な場合は `dwcas.New` を使用してください。

呼び出し側の byte buffer 内に `*Uint128` を配置したい場合（手動 arena など）は、placement helpers を使用してください。`off` からの最悪必要残量は 31 バイトです。

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

## 対応アーキテクチャとバックエンド

- `amd64`: `CMPXCHG16B` により実装。
- `arm64`:
  - 既定: LSE pair-CAS（CASP family）（最高性能）。
  - 任意の LL/SC: `-tags=dwcas_llsc`（`LDAXP`/`STLXP` によるポータブルなベースライン）。

## 安全上の注意

`dwcas` は `unsafe` とアーキテクチャ固有のアセンブリを使用します。

- `*Uint128` は Go ポインタとして到達可能な状態に保ってください。`uintptr` に変換して戻したり、追跡されないメモリに保存しないでください。
- `Uint128` 値をコピーしてもアラインメント保証は引き継がれません。アラインメントは、メソッドに渡すアドレスの性質です。

## License

MIT — [LICENSE](./LICENSE) を参照。

©2025 Hayabusa Cloud Co., Ltd.
