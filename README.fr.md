# dwcas

[![Go Reference](https://pkg.go.dev/badge/github.com/hayabusa-cloud/dwcas.svg)](https://pkg.go.dev/github.com/hayabusa-cloud/dwcas)
[![Go Report Card](https://goreportcard.com/badge/github.com/hayabusa-cloud/dwcas)](https://goreportcard.com/report/github.com/hayabusa-cloud/dwcas)
[![Codecov](https://codecov.io/gh/hayabusa-cloud/dwcas/graph/badge.svg)](https://codecov.io/gh/hayabusa-cloud/dwcas)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

Langues : [English](README.md) | [简体中文](README.zh-CN.md) | [日本語](README.ja.md) | [Español](README.es.md) | **Français**

Primitive portable de compare-and-swap (CAS) 128 bits (double-mot) pour Go.

## En bref

- CAS 128 bits atomique en une seule opération sur `amd64`/`arm64`.
- Conçu pour les algorithmes lock-free (état versionné, mitigation ABA, mises à jour composites).
- Sur `arm64`, le chemin rapide LSE pair-CAS est utilisé par défaut ; LL/SC est disponible via un build tag pour la compatibilité.

## Installation

```bash
go get code.hybscloud.com/dwcas
```

## Démarrage rapide

```text
// New returns a 16-byte aligned *Uint128 suitable for 128-bit compare-and-swap.
v := dwcas.New(1, 2)

old := dwcas.Uint128{Lo: 1, Hi: 2}
newv := dwcas.Uint128{Lo: 3, Hi: 4}

prev, swapped := v.AcqRel(old, newv)
fmt.Println(prev, swapped, v.Lo, v.Hi)
```

## APIs

### Type principal

- `type Uint128 struct { Lo uint64; Hi uint64 }`

### Méthodes compare-exchange

Chaque méthode sur `*dwcas.Uint128` est un CAS 128 bits sous forme compare-exchange.

Toutes les méthodes renvoient :

- `prev` : la valeur observée en mémoire au moment de la tentative de CAS
- `swapped` : `true` si l'échange a eu lieu

Méthodes :

- `(*Uint128) Relaxed(old, new Uint128) (prev Uint128, swapped bool)`
- `(*Uint128) Acquire(old, new Uint128) (prev Uint128, swapped bool)`
- `(*Uint128) Release(old, new Uint128) (prev Uint128, swapped bool)`
- `(*Uint128) AcqRel(old, new Uint128) (prev Uint128, swapped bool)`

### Allocation et placement

- `func New(lo, hi uint64) *Uint128`
- `func CanPlaceAlignedUint128(p []byte, off int) bool`
- `func PlaceAlignedUint128(p []byte, off int) (n int, u128 *Uint128)`

## Ordonnancement mémoire

L'ordonnancement est défini par méthode, et **l'ordonnancement en succès** peut différer de **l'ordonnancement en échec** :

| Méthode | Succès | Échec |
|---------|--------|-------|
| `dwcas.Relaxed` | relaxed | relaxed |
| `dwcas.Acquire` | acquire | relaxed |
| `dwcas.Release` | release | relaxed |
| `dwcas.AcqRel`  | acq_rel | relaxed |

Notes :

- Certains backends peuvent être plus forts que demandé.
  - Sur `amd64`, les opérations `LOCK` sont au minimum acquire-release en succès comme en échec.
  - Sur `arm64` (LSE par défaut), `Release`/`AcqRel` placent une barrière release avant le CAS, ce qui ordonne aussi une tentative échouée.

### Barrières manuelles

Dans de rares cas, l'appelant peut avoir besoin d'une arête d'ordonnancement explicite en dehors des primitives CAS128. `dwcas` fournit trois barrières manuelles :

- `dwcas.BarrierAcquire` : sur `arm64` émet `DMB ISHLD` ; sur `amd64` c'est uniquement une barrière compilateur.
- `dwcas.BarrierRelease` : sur `arm64` émet `DMB ISHST` ; sur `amd64` c'est uniquement une barrière compilateur.
- `dwcas.BarrierFull` : sur `arm64` émet `DMB ISH` ; sur `amd64` c'est uniquement une barrière compilateur.

## Exigence d'alignement

L'adresse du `*dwcas.Uint128` utilisé avec ces méthodes doit être alignée sur 16 octets. Le pointeur doit aussi être non nil.

- Les builds par défaut n'effectuent aucune vérification à l'exécution.
- Garde de debug optionnelle : compiler avec `-tags=dwcasdebug` pour déclencher un panic sur pointeur nil ou mal aligné.

Utilisez `dwcas.New` si vous avez besoin d'un `*Uint128` alloué sur le heap et aligné sur 16 octets.

Si vous devez placer un `*Uint128` dans un buffer de bytes fourni par l'appelant (par exemple, dans une arène gérée manuellement), utilisez les helpers de placement. Le pire cas de bytes restants requis à partir de `off` est 31.

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

## Architectures et backends supportés

- `amd64` : implémenté via `CMPXCHG16B`.
- `arm64` :
  - par défaut : LSE pair-CAS (famille CASP) (meilleures performances).
  - LL/SC optionnel : `-tags=dwcas_llsc` (base portable via `LDAXP`/`STLXP`).

## Notes de sécurité

`dwcas` utilise `unsafe` et de l'assemblage spécifique à l'architecture.

- Gardez `*Uint128` atteignable en tant que pointeur Go. Ne le convertissez pas en `uintptr` puis de nouveau, et ne le stockez pas dans une mémoire non suivie.
- Une valeur `Uint128` copiée ne transporte pas de garanties d'alignement. L'alignement est une propriété de l'adresse passée à ces méthodes.

## Licence

MIT — voir [LICENSE](./LICENSE).

©2025 Hayabusa Cloud Co., Ltd.
