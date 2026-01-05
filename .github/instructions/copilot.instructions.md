# dwcas Copilot Instructions

## Package Overview

dwcas provides double-word compare-and-swap (128-bit CAS) primitives for Go on amd64 and arm64.

## Architecture

### amd64 Implementation
- Uses `LOCK CMPXCHG16B` instruction
- Sequentially consistent by design (no explicit barriers needed)
- Requires 16-byte alignment

### arm64 Implementation
- Uses `CASPD` (LSE) with explicit DMB barriers
- `CASPD` alone is relaxed; ordering requires:
  - `DMB ISHST` ($0xA) for release semantics before
  - `DMB ISHLD` ($0x9) for acquire semantics after
- Go assembler doesn't support ordered variants (CASPAL/CASPALD)

## Memory Ordering Semantics

| Method | Meaning |
|--------|---------|
| `Relaxed` | No ordering guarantees |
| `Acquire` | Loads after CAS see CAS effects |
| `Release` | Stores before CAS visible to CAS |
| `AcqRel` | Both acquire and release |

## Critical Review Points

### Alignment
- All 128-bit CAS operations REQUIRE 16-byte alignment
- `New()` guarantees alignment via over-allocation
- `PlaceAlignedUint128()` computes aligned offset in byte slices

### Pointer Provenance
- Use `unsafe.Add()` (Go 1.17+) not uintptr arithmetic
- Maintains pointer provenance for go vet compliance

### Assembly Correctness
- amd64: `LOCK CMPXCHG16B` atomically compares RDX:RAX with memory, swaps with RCX:RBX
- arm64: Register pairs must be consecutive even-odd (X0-X1, X2-X3)
- Return value is always the previous memory value (enables CAS loop pattern)

## Common Issues to Flag

1. **Non-atomic access to Uint128**: Reading/writing Lo/Hi separately is NOT atomic
2. **Alignment violations**: Pointer must be 16-byte aligned
3. **Barrier omissions on arm64**: Each ordering level needs correct DMB sequence
4. **Register clobbering**: Assembly must preserve caller-saved registers correctly

## Testing Requirements

- Race detector cannot verify dwcas correctness (low-level primitives)
- Contention tests verify atomicity under concurrent access
- Cross-architecture builds: amd64, arm64, riscv64, loong64 (stubs)
