// ©Hayabusa Cloud Co., Ltd. 2025. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

//go:build arm64 && !dwcas_llsc

#include "textflag.h"

// LSE pair-CAS backend using CASPD.
//
// Go's arm64 assembler (Go 1.25) does not currently accept ordered pair-CAS
// mnemonics (e.g. CASPAL/CASPALD), so we use CASPD and add explicit barriers
// when a method requires ordering.
//
// Notes:
// - Assumes ptr is 16-byte aligned and valid (no checks; hot path).
// - CASPD overwrites the "expected" registers (R0:R1) with the memory value on failure,
//   so we preserve the expected old pair in R4:R5 before issuing CASPD.

// Cas128Relaxed(ptr *uint64, oldLo, oldHi, newLo, newHi uint64) (prevLo, prevHi uint64, swapped bool)
TEXT ·Cas128Relaxed(SB), NOSPLIT, $0-57
	MOVD	ptr+0(FP), R8
	MOVD	oldLo+8(FP), R0
	MOVD	oldHi+16(FP), R1
	MOVD	newLo+24(FP), R2
	MOVD	newHi+32(FP), R3

	MOVD	R0, R4
	MOVD	R1, R5

	CASPD	(R0, R1), (R8), (R2, R3)

	MOVD	R0, prevLo+40(FP)
	MOVD	R1, prevHi+48(FP)

	CMP	R0, R4
	CCMP	EQ, R1, R5, $0
	CSET	EQ, R6
	MOVB	R6, swapped+56(FP)
	RET

// Cas128Acquire(ptr *uint64, oldLo, oldHi, newLo, newHi uint64) (prevLo, prevHi uint64, swapped bool)
TEXT ·Cas128Acquire(SB), NOSPLIT, $0-57
	MOVD	ptr+0(FP), R8
	MOVD	oldLo+8(FP), R0
	MOVD	oldHi+16(FP), R1
	MOVD	newLo+24(FP), R2
	MOVD	newHi+32(FP), R3

	MOVD	R0, R4
	MOVD	R1, R5

	CASPD	(R0, R1), (R8), (R2, R3)

	MOVD	R0, prevLo+40(FP)
	MOVD	R1, prevHi+48(FP)

	CMP	R0, R4
	CCMP	EQ, R1, R5, $0
	BNE	cas128_acquire_fail

	// Acquire ordering on success only (ISHLD).
	DMB	$0x9
	CSET	EQ, R6
	MOVB	R6, swapped+56(FP)
	RET

cas128_acquire_fail:
	MOVB	ZR, swapped+56(FP)
	RET

// Cas128Release(ptr *uint64, oldLo, oldHi, newLo, newHi uint64) (prevLo, prevHi uint64, swapped bool)
TEXT ·Cas128Release(SB), NOSPLIT, $0-57
	MOVD	ptr+0(FP), R8
	MOVD	oldLo+8(FP), R0
	MOVD	oldHi+16(FP), R1
	MOVD	newLo+24(FP), R2
	MOVD	newHi+32(FP), R3

	MOVD	R0, R4
	MOVD	R1, R5

	// Release ordering for prior writes (ISHST).
	// This must occur before the conditional store performed by CASPD.
	DMB	$0xA

	CASPD	(R0, R1), (R8), (R2, R3)

	MOVD	R0, prevLo+40(FP)
	MOVD	R1, prevHi+48(FP)

	CMP	R0, R4
	CCMP	EQ, R1, R5, $0
	CSET	EQ, R6
	MOVB	R6, swapped+56(FP)
	RET

// Cas128AcqRel(ptr *uint64, oldLo, oldHi, newLo, newHi uint64) (prevLo, prevHi uint64, swapped bool)
TEXT ·Cas128AcqRel(SB), NOSPLIT, $0-57
	MOVD	ptr+0(FP), R8
	MOVD	oldLo+8(FP), R0
	MOVD	oldHi+16(FP), R1
	MOVD	newLo+24(FP), R2
	MOVD	newHi+32(FP), R3

	MOVD	R0, R4
	MOVD	R1, R5

	// Release ordering for prior writes (ISHST).
	DMB	$0xA

	CASPD	(R0, R1), (R8), (R2, R3)

	MOVD	R0, prevLo+40(FP)
	MOVD	R1, prevHi+48(FP)

	CMP	R0, R4
	CCMP	EQ, R1, R5, $0
	BNE	cas128_acqrel_fail

	// Acquire ordering on success only (ISHLD).
	DMB	$0x9
	CSET	EQ, R6
	MOVB	R6, swapped+56(FP)
	RET

cas128_acqrel_fail:
	MOVB	ZR, swapped+56(FP)
	RET
