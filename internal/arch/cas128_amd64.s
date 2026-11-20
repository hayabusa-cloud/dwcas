// ©Hayabusa Cloud Co., Ltd. 2025. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

//go:build amd64

#include "textflag.h"

// LOCK CMPXCHG16B supplies all four orderings. On return DX:AX holds the
// observed pair and ZF records whether it matched the expected pair.
#define COMPARE_PAIR \
	MOVQ ptr+0(FP), DI; \
	MOVQ old_Lo+8(FP), AX; \
	MOVQ old_Hi+16(FP), DX; \
	MOVQ new_Lo+24(FP), BX; \
	MOVQ new_Hi+32(FP), CX; \
	LOCK; \
	CMPXCHG16B (DI)

#define RETURN_PAIR \
	MOVQ AX, prev_Lo+40(FP); \
	MOVQ DX, prev_Hi+48(FP); \
	SETEQ swapped+56(FP); \
	RET

TEXT ·Cas128Relaxed(SB), NOSPLIT, $0-57
	COMPARE_PAIR
	RETURN_PAIR

TEXT ·Cas128Acquire(SB), NOSPLIT, $0-57
	COMPARE_PAIR
	RETURN_PAIR

TEXT ·Cas128Release(SB), NOSPLIT, $0-57
	COMPARE_PAIR
	RETURN_PAIR

TEXT ·Cas128AcqRel(SB), NOSPLIT, $0-57
	COMPARE_PAIR
	RETURN_PAIR

TEXT ·TryRelaxed(SB), NOSPLIT, $0-41
	COMPARE_PAIR
	SETEQ swapped+40(FP)
	RET

TEXT ·TryAcquire(SB), NOSPLIT, $0-41
	COMPARE_PAIR
	SETEQ swapped+40(FP)
	RET

TEXT ·TryRelease(SB), NOSPLIT, $0-41
	COMPARE_PAIR
	SETEQ swapped+40(FP)
	RET

TEXT ·TryAcqRel(SB), NOSPLIT, $0-41
	COMPARE_PAIR
	SETEQ swapped+40(FP)
	RET
