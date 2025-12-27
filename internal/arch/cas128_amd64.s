// ©Hayabusa Cloud Co., Ltd. 2025. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

//go:build amd64

#include "textflag.h"

// Cas128Relaxed(ptr *uint64, oldLo, oldHi, newLo, newHi uint64) (prevLo, prevHi uint64, swapped bool)
// Implements 128-bit compare-and-swap using LOCK CMPXCHG16B.
//
// Pre-conditions (assumed, not checked): ptr is 16-byte aligned and points to 16 bytes.
TEXT ·Cas128Relaxed(SB), NOSPLIT, $16-57
	// Save callee-saved RBX (required by Go ABI).
	MOVQ	BX, 0(SP)

	MOVQ	ptr+0(FP), DI    // DI = ptr
	MOVQ	oldLo+8(FP), AX  // RDX:RAX expected old value (Lo)
	MOVQ	oldHi+16(FP), DX // expected old value (Hi)
	MOVQ	newLo+24(FP), BX // RCX:RBX desired new value (Lo)
	MOVQ	newHi+32(FP), CX // desired new value (Hi)

	LOCK
	CMPXCHG16B	(DI)

	// RDX:RAX now contains the observed value (old on success, current on failure).
	MOVQ	AX, prevLo+40(FP)
	MOVQ	DX, prevHi+48(FP)

	SETEQ	AL
	MOVB	AL, swapped+56(FP)

	MOVQ	0(SP), BX
	RET

// Cas128Acquire(ptr *uint64, oldLo, oldHi, newLo, newHi uint64) (prevLo, prevHi uint64, swapped bool)
TEXT ·Cas128Acquire(SB), NOSPLIT, $16-57
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

// Cas128Release(ptr *uint64, oldLo, oldHi, newLo, newHi uint64) (prevLo, prevHi uint64, swapped bool)
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

// Cas128AcqRel(ptr *uint64, oldLo, oldHi, newLo, newHi uint64) (prevLo, prevHi uint64, swapped bool)
TEXT ·Cas128AcqRel(SB), NOSPLIT, $16-57
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
