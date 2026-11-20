# dwcas

[![Go Reference](https://pkg.go.dev/badge/code.hybscloud.com/dwcas.svg)](https://pkg.go.dev/code.hybscloud.com/dwcas)
[![Go Report Card](https://goreportcard.com/badge/code.hybscloud.com/dwcas)](https://goreportcard.com/report/code.hybscloud.com/dwcas)
[![Codecov](https://codecov.io/gh/hayabusa-cloud/dwcas/graph/badge.svg)](https://codecov.io/gh/hayabusa-cloud/dwcas)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

Idiomas: [English](README.md) | [简体中文](README.zh-CN.md) | [日本語](README.ja.md) | **Español** | [Français](README.fr.md)

Cargas, escrituras y compare-and-swap (CAS) atómicos de 128 bits para Go.

## De un vistazo

- CAS atómico único de 128 bits en `amd64`/`arm64`.
- Diseñado para algoritmos lock-free (estado versionado, mitigación ABA, actualizaciones compuestas).
- En arm64, la compilación predeterminada requiere FEAT_LSE2. Para CPU anteriores, usa `-tags=dwcas_llsc`.

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
fmt.Println(prev, swapped, v.LoadRelaxed())
```

## APIs

### Tipo principal

- `type Uint128 struct { Lo uint64; Hi uint64 }`

### Cargas y escrituras

- `LoadRelaxed() Uint128` y `LoadAcquire() Uint128` devuelven el par en una sola lectura atómica.
- `StoreRelaxed(value Uint128)` y `StoreRelease(value Uint128)` reemplazan ambas palabras de forma atómica.

### Métodos de comparación e intercambio

Los cuatro métodos `Try` reciben `(old, new Uint128)` y devuelven un `bool`. Un resultado `false` indica que el valor difería de `old`. En arm64 con LL/SC, una comparación coincidente se reintenta si se pierde la reserva, por lo que una llamada carece de un límite fijo de pasos.

Los cuatro métodos compare-exchange también devuelven el par observado.

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

Los métodos CAS especifican por separado el orden en caso de éxito y de fallo.

| Método | Éxito | Fallo |
|--------|-------|-------|
| `(*Uint128).Relaxed`, `(*Uint128).TryRelaxed` | relaxed | relaxed |
| `(*Uint128).Acquire`, `(*Uint128).TryAcquire` | acquire | relaxed |
| `(*Uint128).Release`, `(*Uint128).TryRelease` | release | relaxed |
| `(*Uint128).AcqRel`, `(*Uint128).TryAcqRel`  | acq_rel | relaxed |

Notas:

- Algunos backends son más fuertes de lo solicitado.
  - En `amd64`, las operaciones con `LOCK` son al menos acquire-release tanto en éxito como en fallo.
  - En arm64, los métodos CAS con semántica acquire la mantienen cuando falla la comparación.

### Barreras manuales

En casos raros, el usuario puede necesitar una arista de orden explícita fuera de las primitivas CAS128. `dwcas` proporciona tres barreras manuales:

- `dwcas.BarrierAcquire`: en `arm64` emite `DMB ISHLD`; en `amd64` es solo una barrera del compilador.
- `dwcas.BarrierRelease`: en `arm64` emite `DMB ISHST`; en `amd64` es solo una barrera del compilador.
- `dwcas.BarrierFull`: en `arm64` emite `DMB ISH`; en `amd64` es solo una barrera del compilador.

## Requisito de alineación

El receptor debe apuntar a memoria escribible alineada a 16 bytes. En la compilación predeterminada de arm64, las cargas y escrituras requieren memoria Normal con atributos Inner y Outer Write-Back, como la del heap de Go. Una carga o un CAS fallido pueden volver a escribir el valor observado.

- Las compilaciones por defecto no realizan comprobaciones en tiempo de ejecución.
- Protección de depuración opcional: compila con `-tags=dwcasdebug` para hacer panic en punteros nulos o desalineados.

Usa `dwcas.New` para reservar un `*Uint128` alineado a 16 bytes en memoria gestionada por Go.

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

- En `amd64`, el CAS y las cargas predeterminadas usan `LOCK CMPXCHG16B`. Las escrituras predeterminadas se reintentan si falla la comparación.
  - Para cargas de trabajo con muchas lecturas, `-tags=dwcas_avx` usa una instrucción `VMOVDQA` alineada por cada lectura o escritura independiente. El sistema operativo debe habilitar AVX y la CPU debe garantizar la atomicidad de los accesos vectoriales alineados de 16 bytes a memoria write-back.
- `arm64`:
  - Las cargas y escrituras relajadas usan `LDP` y `STP`. Una carga acquire añade `DMB ISHLD` después de `LDP`. Una escritura release añade `DMB ISH` antes de `STP`. El CAS usa la familia `CASP` de LSE.
  - LL/SC opcional: `-tags=dwcas_llsc` (línea base portable vía `LDXP`/`LDAXP` con `STXP` o `STLXP`).

## Notas de seguridad

`dwcas` usa `unsafe` y ensamblador específico de arquitectura.

- Mantén `*Uint128` alcanzable como puntero de Go. No lo conviertas a `uintptr` y de vuelta, y no lo almacenes en memoria no rastreada.
- Un valor `Uint128` copiado no conserva garantías de alineación. La alineación es una propiedad de la dirección que pasas a estos métodos.

## Licencia

MIT — ver [LICENSE](./LICENSE).

©2025 Hayabusa Cloud Co., Ltd.
