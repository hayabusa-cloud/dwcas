# dwcas

[![Go Reference](https://pkg.go.dev/badge/code.hybscloud.com/dwcas.svg)](https://pkg.go.dev/code.hybscloud.com/dwcas)
[![Go Report Card](https://goreportcard.com/badge/code.hybscloud.com/dwcas)](https://goreportcard.com/report/code.hybscloud.com/dwcas)
[![Codecov](https://codecov.io/gh/hayabusa-cloud/dwcas/graph/badge.svg)](https://codecov.io/gh/hayabusa-cloud/dwcas)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

言語: [English](README.md) | [简体中文](README.zh-CN.md) | **日本語** | [Español](README.es.md) | [Français](README.fr.md)

Go 向けの 128 ビット（ダブルワード）のアトミックなロード、ストア、比較交換（CAS）を提供します。

## 概要

- `amd64` / `arm64` で単一の原子的 128-bit CAS を提供します。
- ロックフリーアルゴリズム（バージョン付き状態、ABA 緩和、複合更新）向けです。
- arm64 の既定ビルドは FEAT_LSE2 を前提とします。従来の CPU では `-tags=dwcas_llsc` を指定してください。

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
fmt.Println(prev, swapped, v.LoadRelaxed())
```

## API

### コア型

- `type Uint128 struct { Lo uint64; Hi uint64 }`

### ロードとストア

- `LoadRelaxed() Uint128` と `LoadAcquire() Uint128` は一貫したペアを返します。
- `StoreRelaxed(value Uint128)` と `StoreRelease(value Uint128)` は両方のワードをアトミックに置き換えます。

### CAS メソッド

4 つの `Try` メソッドは `(old, new Uint128)` を受け取り、`bool` を返します。`false` は観測した値が `old` と異なったことを示します。arm64 の LL/SC では、値が一致しても予約を失うと再試行するため、1 回の呼び出しのステップ数に固定の上限はありません。

4 つの compare-exchange メソッドは観測した値の組も返します。

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

CAS メソッドは成功時と失敗時のメモリオーダーを個別に定めます。

| メソッド | 成功 | 失敗 |
|----------|------|------|
| `(*Uint128).Relaxed`, `(*Uint128).TryRelaxed` | relaxed | relaxed |
| `(*Uint128).Acquire`, `(*Uint128).TryAcquire` | acquire | relaxed |
| `(*Uint128).Release`, `(*Uint128).TryRelease` | release | relaxed |
| `(*Uint128).AcqRel`, `(*Uint128).TryAcqRel`  | acq_rel | relaxed |

補足:

- 一部のバックエンドは要求より強いオーダーを提供します。
  - `amd64` では `LOCK` 付き命令は成功/失敗ともに少なくとも acquire-release です。
  - arm64 では、acquire 順序の CAS メソッドは比較が失敗した場合も acquire 順序で読み取ります。

### 手動バリア

まれに、CAS128 の外側で明示的な順序付けが必要になることがあります。`dwcas` は 3 つの手動バリアを提供します。

- `dwcas.BarrierAcquire`: `arm64` は `DMB ISHLD`、`amd64` はコンパイラバリアのみ。
- `dwcas.BarrierRelease` は、ライトバック方式の共有メモリへの通常のアクセスで、先行する書き込みと後続の書き込みを順序付けます。`arm64` では `DMB ISHST`、`amd64` ではコンパイラバリアを使用します。先行する読み込みも順序付けるには、release ストアまたは成功した release CAS で公開します。
- `dwcas.BarrierFull`: `arm64` は `DMB ISH`、`amd64` はコンパイラバリアのみ。

## アラインメント要件

レシーバーは 16 バイト境界にアラインされた書き込み可能なメモリを指す必要があります。arm64 の既定バックエンドでロード・ストアを行うには、Go のヒープと同じく Inner と Outer がともに Write-Back の Normal メモリが必要です。ロードや失敗した CAS でも、読み取った値を書き戻す場合があります。

- 既定ビルドではランタイムチェックを行いません。
- デバッグ用ガード: `-tags=dwcasdebug` でビルドすると、nil または未アラインで panic します。

`dwcas.New` は Go が管理するメモリに 16 バイト境界に整列した `*Uint128` を確保します。

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

- `amd64` では、CAS と既定のロードに `LOCK CMPXCHG16B` を使います。既定のストアは比較に失敗すると再試行します。
  - 読み取りが多い場合、`-tags=dwcas_avx` を指定すると、単独のロードとストアはそれぞれ整列した `VMOVDQA` 1 命令で対象を読み書きします。OS が AVX を有効にし、CPU がライトバックメモリに対して 16 バイト境界に整列した 16 バイトのベクトルアクセスの原子性を保証する必要があります。
- `arm64`:
  - relaxed ロード・ストアには `LDP` と `STP` を使います。acquire ロードでは `LDP` の後に `DMB ISHLD` を、release ストアでは `STP` の前に `DMB ISH` を実行します。CAS には LSE の `CASP` 命令群を使います。
  - 任意の LL/SC: `-tags=dwcas_llsc`（`LDXP`/`LDAXP` と `STXP` または `STLXP` によるポータブルなベースライン）。

## 安全上の注意

`dwcas` は `unsafe` とアーキテクチャ固有のアセンブリを使用します。

- `*Uint128` は Go ポインタとして到達可能な状態に保ってください。`uintptr` に変換して戻したり、追跡されないメモリに保存しないでください。
- `Uint128` 値をコピーしてもアラインメント保証は引き継がれません。アラインメントは、メソッドに渡すアドレスの性質です。

## License

MIT — [LICENSE](./LICENSE) を参照。

©2025 Hayabusa Cloud Co., Ltd.
