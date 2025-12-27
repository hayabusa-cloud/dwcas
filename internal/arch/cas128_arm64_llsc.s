// ©Hayabusa Cloud Co., Ltd. 2025. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

//go:build arm64 && dwcas_llsc

#include "textflag.h"

// LL/SC variant for arm64 using LDXP/STXP (relaxed) and STLXP (release).
//
// Notes:
// - Assumes ptr is 16-byte aligned and valid (no checks; hot path).
// - On store-exclusive failure (contention), we retry.
// - On mismatch, return the observed value immediately (no extra loads).

// Cas128Relaxed(ptr *uint64, oldLo, oldHi, newLo, newHi uint64) (prevLo, prevHi uint64, swapped bool)
TEXT ·Cas128Relaxed(SB), NOSPLIT, $0-57
	MOVD	ptr+0(FP), R8
	MOVD	oldLo+8(FP), R0
	MOVD	oldHi+16(FP), R1
	MOVD	newLo+24(FP), R2
	MOVD	newHi+32(FP), R3

loop_llsc_relaxed:
	LDXP	(R8), (R4, R5)
	CMP	R4, R0
	BNE	not_eq_llsc_relaxed
	CMP	R5, R1
	BNE	not_eq_llsc_relaxed
	STXP	(R2, R3), (R8), R6
	CBNZ	R6, loop_llsc_relaxed

	MOVD	R4, prevLo+40(FP)
	MOVD	R5, prevHi+48(FP)
	CSET	EQ, R6
	MOVB	R6, swapped+56(FP)
	RET

not_eq_llsc_relaxed:
	MOVD	R4, prevLo+40(FP)
	MOVD	R5, prevHi+48(FP)
	MOVB	ZR, swapped+56(FP)
	RET

// Cas128Acquire(ptr *uint64, oldLo, oldHi, newLo, newHi uint64) (prevLo, prevHi uint64, swapped bool)
TEXT ·Cas128Acquire(SB), NOSPLIT, $0-57
	MOVD	ptr+0(FP), R8
	MOVD	oldLo+8(FP), R0
	MOVD	oldHi+16(FP), R1
	MOVD	newLo+24(FP), R2
	MOVD	newHi+32(FP), R3

loop_llsc_acquire:
	LDXP	(R8), (R4, R5)
	CMP	R4, R0
	BNE	not_eq_llsc_acquire
	CMP	R5, R1
	BNE	not_eq_llsc_acquire
	STXP	(R2, R3), (R8), R6
	CBNZ	R6, loop_llsc_acquire

	// Acquire ordering on success only (ISHLD).
	DMB	$0x9

	MOVD	R4, prevLo+40(FP)
	MOVD	R5, prevHi+48(FP)
	CSET	EQ, R6
	MOVB	R6, swapped+56(FP)
	RET

not_eq_llsc_acquire:
	MOVD	R4, prevLo+40(FP)
	MOVD	R5, prevHi+48(FP)
	MOVB	ZR, swapped+56(FP)
	RET

// Cas128Release(ptr *uint64, oldLo, oldHi, newLo, newHi uint64) (prevLo, prevHi uint64, swapped bool)
TEXT ·Cas128Release(SB), NOSPLIT, $0-57
	MOVD	ptr+0(FP), R8
	MOVD	oldLo+8(FP), R0
	MOVD	oldHi+16(FP), R1
	MOVD	newLo+24(FP), R2
	MOVD	newHi+32(FP), R3

loop_llsc_release:
	LDXP	(R8), (R4, R5)
	CMP	R4, R0
	BNE	not_eq_llsc_release
	CMP	R5, R1
	BNE	not_eq_llsc_release
	STLXP	(R2, R3), (R8), R6
	CBNZ	R6, loop_llsc_release

	MOVD	R4, prevLo+40(FP)
	MOVD	R5, prevHi+48(FP)
	CSET	EQ, R6
	MOVB	R6, swapped+56(FP)
	RET

not_eq_llsc_release:
	MOVD	R4, prevLo+40(FP)
	MOVD	R5, prevHi+48(FP)
	MOVB	ZR, swapped+56(FP)
	RET

// Cas128AcqRel(ptr *uint64, oldLo, oldHi, newLo, newHi uint64) (prevLo, prevHi uint64, swapped bool)
TEXT ·Cas128AcqRel(SB), NOSPLIT, $0-57
	MOVD	ptr+0(FP), R8
	MOVD	oldLo+8(FP), R0
	MOVD	oldHi+16(FP), R1
	MOVD	newLo+24(FP), R2
	MOVD	newHi+32(FP), R3

loop_llsc_acqrel:
	LDXP	(R8), (R4, R5)
	CMP	R4, R0
	BNE	not_eq_llsc_acqrel
	CMP	R5, R1
	BNE	not_eq_llsc_acqrel
	STLXP	(R2, R3), (R8), R6
	CBNZ	R6, loop_llsc_acqrel

	// Acquire ordering on success only (ISHLD).
	DMB	$0x9

	MOVD	R4, prevLo+40(FP)
	MOVD	R5, prevHi+48(FP)
	CSET	EQ, R6
	MOVB	R6, swapped+56(FP)
	RET

not_eq_llsc_acqrel:
	MOVD	R4, prevLo+40(FP)
	MOVD	R5, prevHi+48(FP)
	MOVB	ZR, swapped+56(FP)
	RET
