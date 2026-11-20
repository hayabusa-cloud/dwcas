# dwcas

[![Go Reference](https://pkg.go.dev/badge/code.hybscloud.com/dwcas.svg)](https://pkg.go.dev/code.hybscloud.com/dwcas)
[![Go Report Card](https://goreportcard.com/badge/code.hybscloud.com/dwcas)](https://goreportcard.com/report/code.hybscloud.com/dwcas)
[![Codecov](https://codecov.io/gh/hayabusa-cloud/dwcas/graph/badge.svg)](https://codecov.io/gh/hayabusa-cloud/dwcas)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

Langues : [English](README.md) | [简体中文](README.zh-CN.md) | [日本語](README.ja.md) | [Español](README.es.md) | **Français**

Chargements, écritures et compare-and-swap (CAS) atomiques de 128 bits pour Go.

## En bref

- CAS 128 bits atomique en une seule opération sur `amd64`/`arm64`.
- Conçu pour les algorithmes lock-free (état versionné, mitigation ABA, mises à jour composites).
- Sur arm64, la compilation par défaut cible FEAT_LSE2. Pour les processeurs plus anciens, utilisez `-tags=dwcas_llsc`.

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
fmt.Println(prev, swapped, v.LoadRelaxed())
```

## APIs

### Type principal

- `type Uint128 struct { Lo uint64; Hi uint64 }`

### Chargements et écritures

- `LoadRelaxed() Uint128` et `LoadAcquire() Uint128` renvoient la paire en une seule lecture atomique.
- `StoreRelaxed(value Uint128)` et `StoreRelease(value Uint128)` remplacent les deux mots en une seule écriture atomique.

### Méthodes de comparaison et d’échange

Les quatre méthodes `Try` prennent `(old, new Uint128)` et renvoient un `bool`. Un résultat `false` indique que la valeur différait de `old`. Sur arm64 avec LL/SC, si les valeurs correspondent et que la réservation est perdue, la méthode réessaie. Le nombre d’étapes par appel est donc sans borne fixe.

Les quatre méthodes compare-exchange renvoient aussi la paire observée.

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

Les méthodes CAS spécifient séparément les ordres mémoire en cas de succès et d’échec.

| Méthode | Succès | Échec |
|---------|--------|-------|
| `(*Uint128).Relaxed`, `(*Uint128).TryRelaxed` | relaxed | relaxed |
| `(*Uint128).Acquire`, `(*Uint128).TryAcquire` | acquire | relaxed |
| `(*Uint128).Release`, `(*Uint128).TryRelease` | release | relaxed |
| `(*Uint128).AcqRel`, `(*Uint128).TryAcqRel`  | acq_rel | relaxed |

Notes :

- Certains backends peuvent être plus forts que demandé.
  - Sur `amd64`, les opérations `LOCK` sont au minimum acquire-release en succès comme en échec.
  - Sur arm64, les méthodes CAS avec sémantique acquire la conservent en cas d’échec de la comparaison.

### Barrières manuelles

Dans de rares cas, l'appelant peut avoir besoin d'une arête d'ordonnancement explicite en dehors des primitives CAS128. `dwcas` fournit trois barrières manuelles :

- `dwcas.BarrierAcquire` : sur `arm64` émet `DMB ISHLD` ; sur `amd64` c'est uniquement une barrière compilateur.
- `dwcas.BarrierRelease` : sur `arm64` émet `DMB ISHST` ; sur `amd64` c'est uniquement une barrière compilateur.
- `dwcas.BarrierFull` : sur `arm64` émet `DMB ISH` ; sur `amd64` c'est uniquement une barrière compilateur.

## Exigence d'alignement

Le récepteur doit pointer vers une zone accessible en écriture et alignée sur 16 octets. Sur arm64, les chargements et écritures du backend par défaut exigent une mémoire Normal avec les attributs Inner et Outer Write-Back, comme celle du tas Go. Un chargement ou un CAS en échec peut réécrire la valeur observée.

- Les builds par défaut n'effectuent aucune vérification à l'exécution.
- Garde de debug optionnelle : compiler avec `-tags=dwcasdebug` pour déclencher un panic sur pointeur nil ou mal aligné.

Utilisez `dwcas.New` pour allouer un `*Uint128` aligné sur 16 octets dans la mémoire gérée par Go.

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

- Sur `amd64`, le CAS et les lectures par défaut utilisent `LOCK CMPXCHG16B`. Les écritures par défaut sont retentées si la comparaison échoue.
  - Pour les charges à forte proportion de lectures, `-tags=dwcas_avx` utilise une instruction `VMOVDQA` alignée par lecture ou écriture indépendante. Le système doit activer AVX et le processeur doit garantir l’atomicité des accès vectoriels alignés de 16 octets en mémoire write-back.
- `arm64` :
  - Les accès relâchés utilisent `LDP` et `STP`. Un chargement acquire ajoute `DMB ISHLD` après `LDP`. Une écriture release ajoute `DMB ISH` avant `STP`. Le CAS utilise la famille `CASP` de LSE.
  - LL/SC optionnel : `-tags=dwcas_llsc` (base portable via `LDXP`/`LDAXP` avec `STXP` ou `STLXP`).

## Notes de sécurité

`dwcas` utilise `unsafe` et de l'assemblage spécifique à l'architecture.

- Gardez `*Uint128` atteignable en tant que pointeur Go. Ne le convertissez pas en `uintptr` puis de nouveau, et ne le stockez pas dans une mémoire non suivie.
- Une valeur `Uint128` copiée ne transporte pas de garanties d'alignement. L'alignement est une propriété de l'adresse passée à ces méthodes.

## Licence

MIT — voir [LICENSE](./LICENSE).

©2025 Hayabusa Cloud Co., Ltd.
