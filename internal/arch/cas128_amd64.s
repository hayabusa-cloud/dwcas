// ©Hayabusa Cloud Co., Ltd. 2025. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

//go:build amd64

#include "textflag.h"

// On x86-64 TSO, LOCK CMPXCHG16B provides sequentially consistent semantics,
// which is stronger than any of Relaxed/Acquire/Release/AcqRel.
// All orderings are therefore equivalent. Cas128Release is the canonical
// implementation (most common in lock-free producer/publish patterns).

// Cas128Release is the canonical implementation.
// Pre-conditions (assumed, not checked): ptr is 16-byte aligned.
TEXT ·Cas128Release(SB), NOSPLIT, $16-57
	MOVQ	BX, 0(SP)
	MOVQ	ptr+0(FP), DI
	MOVQ	oldLo+8(FP), AX
	MOVQ	oldHi+16(FP), DX
	MOVQ	newLo+24(FP), BX
	MOVQ	newHi+32(FP), CX
	LOCK
	CMPXCHG16B	(DI)
	MOVQ	AX, prevLo+40(FP)
	MOVQ	DX, prevHi+48(FP)
	SETEQ	AL
	MOVB	AL, swapped+56(FP)
	MOVQ	0(SP), BX
	RET

// All other orderings tail-call to Cas128Release.
TEXT ·Cas128Relaxed(SB), NOSPLIT, $0-57
	JMP	·Cas128Release(SB)

TEXT ·Cas128Acquire(SB), NOSPLIT, $0-57
	JMP	·Cas128Release(SB)

TEXT ·Cas128AcqRel(SB), NOSPLIT, $0-57
	JMP	·Cas128Release(SB)
