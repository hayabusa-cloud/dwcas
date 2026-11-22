# dwcas

[![Go Reference](https://pkg.go.dev/badge/code.hybscloud.com/dwcas.svg)](https://pkg.go.dev/code.hybscloud.com/dwcas)
[![Codecov](https://codecov.io/gh/hayabusa-cloud/dwcas/graph/badge.svg)](https://codecov.io/gh/hayabusa-cloud/dwcas)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

Idiomas: [English](README.md) | [简体中文](README.zh-CN.md) | **Español** | [日本語](README.ja.md) | [Français](README.fr.md)

Cargas, escrituras y compare-and-swap (CAS) atómicos de 128 bits para Go.

## De un vistazo

- CAS atómico único de 128 bits en `amd64`/`arm64`.
- Diseñado para algoritmos lock-free (estado versionado, mitigación ABA, actualizaciones compuestas).
- En arm64, la compilación predeterminada requiere FEAT_LSE2. Para CPU anteriores, usa `-tags=dwcas_llsc`.

## Instalación

```bash
go get code.hybscloud.com/dwcas
```

## Ejemplos

### Comparar e intercambiar un par

`dwcas.New` reserva un par alineado a 16 bytes. `AcqRel` devuelve el par observado e indica si lo reemplazó.

```go
v := dwcas.New(1, 2)

old := dwcas.Uint128{Lo: 1, Hi: 2}
newv := dwcas.Uint128{Lo: 3, Hi: 4}

prev, swapped := v.AcqRel(old, newv)
fmt.Println(prev, swapped, v.LoadRelaxed()) // {1 2} true {3 4}
```

### Publicar un rango de bytes

El productor publica en una sola operación el desplazamiento (`Lo`) y la longitud (`Hi`) de un bloque de datos no vacío, ambos en bytes. Una longitud de cero indica que el descriptor está vacío.

```go
buffer := make([]byte, 64)
published := dwcas.New(0, 0)

// Productor
const offset = 8
n := copy(buffer[offset:], "hello")
published.StoreRelease(dwcas.Uint128{Lo: offset, Hi: uint64(n)})

// Lector
span := published.LoadAcquire()
if span.Hi != 0 {
	fmt.Printf("%s\n", buffer[span.Lo:span.Lo+span.Hi])
}
```

Cuando el lector observa el par publicado, puede leer los bytes que escribió el productor. El productor conserva intacta esa zona hasta que el lector termina.

La aplicación define la disposición de los campos de `published`. En un anillo de búferes proporcionados a `io_uring` en Linux, una escritura de 128 bits en el primer descriptor también escribiría el índice de cola de 16 bits situado a 14 bytes del inicio del descriptor. `uring.BufferRing.Add` escribe la dirección, la longitud y el identificador del búfer. `Advance` actualiza el índice para publicar los búferes. [Estructura del anillo en Linux](https://github.com/torvalds/linux/blob/v6.18/include/uapi/linux/io_uring.h#L816-L836)

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

En arm64, estas barreras ordenan los accesos ordinarios a la memoria compartida de tipo write-back dentro del dominio inner-shareable.

- `dwcas.BarrierAcquire` ordena las lecturas anteriores antes de las lecturas y escrituras posteriores (`DMB ISHLD`).
- `dwcas.BarrierRelease` ordena las escrituras anteriores antes de las posteriores (`DMB ISHST`). Para ordenar también las lecturas anteriores, publica con una escritura release o con un CAS release que tenga éxito.
- `dwcas.BarrierFull` ordena todas las lecturas y escrituras anteriores antes de las posteriores (`DMB ISH`).

En amd64, las tres funciones actúan como barreras del compilador. El procesador puede adelantar una lectura a una escritura anterior en otra dirección.

## Requisito de alineación

El receptor debe apuntar a memoria escribible alineada a 16 bytes. En la compilación predeterminada de arm64, las cargas y escrituras requieren memoria Normal con atributos Inner y Outer Write-Back, como la del heap de Go. Una carga o un CAS fallido pueden volver a escribir el valor observado. Usa los métodos atómicos del par para acceder a él de forma concurrente.

En la compilación predeterminada, quien llama garantiza la alineación del receptor. Con `-tags=dwcasdebug`, los métodos atómicos provocan un panic si el receptor es `nil` o está mal alineado.

Usa `dwcas.New` para reservar un `*Uint128` alineado a 16 bytes en memoria gestionada por Go.

`CanPlaceAlignedUint128` indica si `off` es un valor no negativo y quedan al menos 31 bytes en el slice a partir de ese desplazamiento. Si la comprobación falla, `PlaceAlignedUint128` provoca un panic. Si tiene éxito, devuelve el número de bytes utilizados (entre 16 y 31, incluido el relleno) y el puntero alineado.

```go
buf := make([]byte, 256)
off := 7

n, v := dwcas.PlaceAlignedUint128(buf, off)
*v = dwcas.Uint128{Lo: 1, Hi: 2}

off += n
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
- Si la dirección de un objeto de Go solo se guarda en `Lo` o `Hi`, el recolector de basura puede liberar el objeto. Una cola puede mantener los punteros en un slice alcanzable y guardar en el par el índice de la celda y su número de generación.
- Un valor `Uint128` copiado no conserva garantías de alineación. La alineación es una propiedad de la dirección que pasas a estos métodos.

## Licencia

MIT — ver [LICENSE](./LICENSE).

©2025 Hayabusa Cloud Co., Ltd.
