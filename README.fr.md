# dwcas

[![Go Reference](https://pkg.go.dev/badge/code.hybscloud.com/dwcas.svg)](https://pkg.go.dev/code.hybscloud.com/dwcas)
[![Codecov](https://codecov.io/gh/hayabusa-cloud/dwcas/graph/badge.svg)](https://codecov.io/gh/hayabusa-cloud/dwcas)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

Langues : [English](README.md) | [简体中文](README.zh-CN.md) | [Español](README.es.md) | [日本語](README.ja.md) | **Français**

Chargements, écritures et compare-and-swap (CAS) atomiques de 128 bits pour Go.

## En bref

- CAS 128 bits atomique en une seule opération sur `amd64`/`arm64`.
- Conçu pour les algorithmes lock-free (état versionné, mitigation ABA, mises à jour composites).
- Sur arm64, la compilation par défaut cible FEAT_LSE2. Pour les processeurs plus anciens, utilisez `-tags=dwcas_llsc`.

## Installation

```bash
go get code.hybscloud.com/dwcas
```

## Exemples

### Comparer et échanger une paire

`dwcas.New` alloue une paire alignée sur 16 octets. `AcqRel` renvoie la paire observée et indique si l’échange a réussi.

```go
v := dwcas.New(1, 2)

old := dwcas.Uint128{Lo: 1, Hi: 2}
newv := dwcas.Uint128{Lo: 3, Hi: 4}

prev, swapped := v.AcqRel(old, newv)
fmt.Println(prev, swapped, v.LoadRelaxed()) // {1 2} true {3 4}
```

### Publier une plage d’octets

Le producteur publie ensemble la position de départ (`Lo`) et la longueur (`Hi`) d’un bloc de données non vide, toutes deux exprimées en octets. Une longueur nulle indique un descripteur vide.

```go
buffer := make([]byte, 64)
published := dwcas.New(0, 0)

// Producteur
const offset = 8
n := copy(buffer[offset:], "hello")
published.StoreRelease(dwcas.Uint128{Lo: offset, Hi: uint64(n)})

// Lecteur
span := published.LoadAcquire()
if span.Hi != 0 {
	fmt.Printf("%s\n", buffer[span.Lo:span.Lo+span.Hi])
}
```

Dès que le lecteur observe la paire publiée, il peut lire les octets écrits par le producteur. Le producteur conserve leur contenu intact jusqu’à la fin de la lecture.

Si les goroutines se synchronisent uniquement au moyen de cette paire, `-race` peut signaler un conflit d’accès au tampon, car les opérations sur la paire utilisent du code assembleur sans instrumentation.

L’application définit la disposition des champs de `published`. Dans un anneau de tampons fournis à `io_uring` sous Linux, une écriture de 128 bits dans le premier descripteur écrirait aussi l’indice de fin de l’anneau, un champ de 16 bits situé à 14 octets du début du descripteur. `uring.BufferRing.Add` écrit l’adresse, la longueur et l’identifiant du tampon. `Advance` publie les tampons en mettant cet indice à jour. [Structure de l’anneau sous Linux](https://github.com/torvalds/linux/blob/v6.18/include/uapi/linux/io_uring.h#L816-L836)

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

Sur arm64, ces barrières ordonnent les accès ordinaires à la mémoire partagée de type write-back dans le domaine de partage interne (Inner Shareable).

- `dwcas.BarrierAcquire` place les lectures antérieures avant les lectures et écritures ultérieures (`DMB ISHLD`).
- `dwcas.BarrierRelease` place les écritures antérieures avant les suivantes (`DMB ISHST`). Pour ordonner aussi les lectures antérieures, publiez avec une écriture release ou un CAS release qui réussit.
- `dwcas.BarrierFull` place toutes les lectures et écritures antérieures avant les lectures et écritures ultérieures (`DMB ISH`).

Sur amd64, les trois fonctions agissent comme des barrières de compilation. Le processeur peut réordonner une lecture par rapport à une écriture antérieure à une autre adresse.

## Exigence d'alignement

Le récepteur doit pointer vers une zone accessible en écriture et alignée sur 16 octets. Sur arm64, les chargements et écritures du backend par défaut exigent une mémoire Normal avec les attributs Inner et Outer Write-Back, comme celle du tas Go. Un chargement ou un CAS en échec peut réécrire la valeur observée. Pour les accès concurrents, utilisez les méthodes atomiques de la paire.

Par défaut, l’appelant garantit l’alignement du récepteur. Avec `-tags=dwcasdebug`, les méthodes atomiques déclenchent un panic si le récepteur est `nil` ou mal aligné.

Utilisez `dwcas.New` pour allouer un `*Uint128` aligné sur 16 octets dans la mémoire gérée par Go.

`CanPlaceAlignedUint128` indique si `off` est positif ou nul et s’il reste au moins 31 octets dans la tranche à partir de cette position. Si la vérification échoue, `PlaceAlignedUint128` déclenche un panic. En cas de succès, la fonction renvoie le nombre d’octets utilisés (de 16 à 31, alignement compris) et le pointeur aligné.

```go
buf := make([]byte, 256)
off := 7

n, v := dwcas.PlaceAlignedUint128(buf, off)
*v = dwcas.Uint128{Lo: 1, Hi: 2}

off += n
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
- Si l’adresse d’un objet Go est conservée uniquement dans `Lo` ou `Hi`, le ramasse-miettes peut libérer l’objet. Une file peut conserver les pointeurs dans une tranche qui reste accessible et stocker dans la paire l’indice de la case et son numéro de génération.
- Une valeur `Uint128` copiée ne transporte pas de garanties d'alignement. L'alignement est une propriété de l'adresse passée à ces méthodes.

## Licence

MIT — voir [LICENSE](./LICENSE).

©2025 Hayabusa Cloud Co., Ltd.
