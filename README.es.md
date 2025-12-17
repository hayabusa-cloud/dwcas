# dwcas

[![Go Reference](https://pkg.go.dev/badge/github.com/hayabusa-cloud/dwcas.svg)](https://pkg.go.dev/github.com/hayabusa-cloud/dwcas)
[![Go Report Card](https://goreportcard.com/badge/github.com/hayabusa-cloud/dwcas)](https://goreportcard.com/report/github.com/hayabusa-cloud/dwcas)
[![Codecov](https://codecov.io/gh/hayabusa-cloud/dwcas/graph/badge.svg)](https://codecov.io/gh/hayabusa-cloud/dwcas)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

Idiomas: [English](README.md) | [简体中文](README.zh-CN.md) | [日本語](README.ja.md) | **Español** | [Français](README.fr.md)

Primitiva portable de compare-and-swap (CAS) de 128 bits (doble palabra) para Go.

## De un vistazo

- CAS atómico único de 128 bits en `amd64`/`arm64`.
- Diseñado para algoritmos lock-free (estado versionado, mitigación ABA, actualizaciones compuestas).
- En `arm64` usa por defecto una ruta rápida LSE pair-CAS; LL/SC está disponible mediante build tag por compatibilidad.

## Instalación

```bash
go get code.hybscloud.com/dwcas
```

## Inicio rápido

```text
// New returns a 16-byte aligned *Uint128 suitable for 128-bit compare-and-swap.
v := dwcas.New(1, 2)

old := dwcas.Uint128{Lo: 1, Hi: 2}
newv := dwcas.Uint128{Lo: 3, Hi: 4}

prev, swapped := v.AcqRel(old, newv)
fmt.Println(prev, swapped, v.Lo, v.Hi)
```

## APIs

### Tipo principal

- `type Uint128 struct { Lo uint64; Hi uint64 }`

### Métodos compare-exchange

Cada método de `*dwcas.Uint128` es un CAS de 128 bits en forma compare-exchange.

Todos los métodos devuelven:

- `prev`: el valor observado en memoria en el momento del intento de CAS
- `swapped`: `true` si el intercambio ocurrió

Métodos:

- `(*Uint128) Relaxed(old, new Uint128) (prev Uint128, swapped bool)`
- `(*Uint128) Acquire(old, new Uint128) (prev Uint128, swapped bool)`
- `(*Uint128) Release(old, new Uint128) (prev Uint128, swapped bool)`
- `(*Uint128) AcqRel(old, new Uint128) (prev Uint128, swapped bool)`

### Asignación y colocación

- `func New(lo, hi uint64) *Uint128`
- `func CanPlaceAlignedUint128(p []byte, off int) bool`
- `func PlaceAlignedUint128(p []byte, off int) (n int, u128 *Uint128)`

## Orden de memoria

El orden se especifica por método, con **orden en éxito** potencialmente distinto del **orden en fallo**:

| Método | Éxito | Fallo |
|--------|-------|-------|
| `dwcas.Relaxed` | relaxed | relaxed |
| `dwcas.Acquire` | acquire | relaxed |
| `dwcas.Release` | release | relaxed |
| `dwcas.AcqRel`  | acq_rel | relaxed |

Notas:

- Algunos backends son más fuertes de lo solicitado.
  - En `amd64`, las operaciones con `LOCK` son al menos acquire-release tanto en éxito como en fallo.
  - En `arm64` (LSE por defecto), `Release`/`AcqRel` colocan una barrera release antes del CAS, lo que también ordena un intento fallido.

### Barreras manuales

En casos raros, el usuario puede necesitar una arista de orden explícita fuera de las primitivas CAS128. `dwcas` proporciona tres barreras manuales:

- `dwcas.BarrierAcquire`: en `arm64` emite `DMB ISHLD`; en `amd64` es solo una barrera del compilador.
- `dwcas.BarrierRelease`: en `arm64` emite `DMB ISHST`; en `amd64` es solo una barrera del compilador.
- `dwcas.BarrierFull`: en `arm64` emite `DMB ISH`; en `amd64` es solo una barrera del compilador.

## Requisito de alineación

La dirección de `*dwcas.Uint128` usada con estos métodos debe estar alineada a 16 bytes. El puntero también debe ser no nulo.

- Las compilaciones por defecto no realizan comprobaciones en tiempo de ejecución.
- Protección de depuración opcional: compila con `-tags=dwcasdebug` para hacer panic en punteros nulos o desalineados.

Usa `dwcas.New` si necesitas un `*Uint128` alineado a 16 bytes en el heap.

Si necesitas colocar un `*Uint128` dentro de un buffer de bytes proporcionado por el usuario (por ejemplo, en un arena manual), usa los helpers de colocación. El peor caso de bytes restantes requeridos desde `off` es 31.

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

## Arquitecturas y backends compatibles

- `amd64`: implementado mediante `CMPXCHG16B`.
- `arm64`:
  - por defecto: LSE pair-CAS (familia CASP) (mejor rendimiento).
  - LL/SC opcional: `-tags=dwcas_llsc` (línea base portable vía `LDAXP`/`STLXP`).

## Notas de seguridad

`dwcas` usa `unsafe` y ensamblador específico de arquitectura.

- Mantén `*Uint128` alcanzable como puntero de Go. No lo conviertas a `uintptr` y de vuelta, y no lo almacenes en memoria no rastreada.
- Un valor `Uint128` copiado no conserva garantías de alineación. La alineación es una propiedad de la dirección que pasas a estos métodos.

## Licencia

MIT — ver [LICENSE](./LICENSE).

©2025 Hayabusa Cloud Co., Ltd.
