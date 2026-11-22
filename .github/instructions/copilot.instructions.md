# dwcas Copilot Instructions

## Package Overview

dwcas provides 128-bit atomic loads, stores, and compare-and-swap on amd64 and arm64.

## Architecture

### amd64 Implementation
- CAS and the default pair loads and stores use `LOCK CMPXCHG16B`. An unconditional store retries on a mismatch.
- With `dwcas_avx`, standalone loads and stores use aligned 128-bit `VMOVDQA` on the receiver. The target must guarantee atomic aligned 16-byte vector accesses to write-back memory and the OS must enable AVX. Ordinary write-back memory supplies acquire and release ordering.

### arm64 Implementation
- Default builds target FEAT_LSE2. Aligned 128-bit relaxed loads and stores use LDP and STP. Acquire loads add DMB ISHLD after LDP. Release stores add DMB ISH before STP.
- On LSE, both result forms use CASP, CASPA, CASPL, or CASPAL according to the method's ordering.
- With `dwcas_llsc`, loads and pair-returning mismatches validate both sampled words with a successful store-exclusive. A boolean mismatch needs only one differing word.

## Memory Ordering Semantics

| Method | Meaning |
|--------|---------|
| `Relaxed` | No ordering guarantees |
| `Acquire` | A successful CAS orders subsequent reads and writes |
| `Release` | A successful CAS orders preceding reads and writes |
| `AcqRel` | Both acquire and release |

## Critical Review Points

### Alignment
- All 128-bit operations require writable, 16-byte-aligned memory. Default arm64 loads and stores also require Normal memory with Inner and Outer Write-Back cacheability.
- `New()` guarantees alignment via over-allocation
- `PlaceAlignedUint128()` computes aligned offset in byte slices

### Pointer Provenance
- Use `unsafe.Add()` (Go 1.17+) not uintptr arithmetic
- Maintains pointer provenance for go vet compliance

### Assembly Correctness
- amd64: `LOCK CMPXCHG16B` atomically compares RDX:RAX with memory, swaps with RCX:RBX
- arm64: Register pairs must be consecutive even-odd (X0-X1, X2-X3)
- Pair-returning methods return a coherent previous value on either outcome. Each `Try` method returns only success or failure.

## Common Issues to Flag

1. **Non-atomic access to Uint128**: Reading/writing Lo/Hi separately is NOT atomic
2. **Alignment violations**: Pointer must be 16-byte aligned
3. **arm64 ordering**: Verify the CAS or exclusive sequence implements the requested success ordering
4. **Register clobbering**: Follow the Go assembly ABI for the target architecture

## Testing Requirements

- Race detector cannot verify dwcas correctness (low-level primitives)
- Contention tests verify atomicity under concurrent access
- Cross-architecture builds: amd64, arm64, riscv64, loong64 (stubs)
