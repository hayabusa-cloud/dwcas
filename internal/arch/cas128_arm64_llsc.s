// ©Hayabusa Cloud Co., Ltd. 2025. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

//go:build arm64 && dwcas_llsc

#include "textflag.h"

#define LOAD_CAS_ARGS \
	MOVD ptr+0(FP), R8; \
	LDP old_Lo+8(FP), (R0, R1); \
	LDP new_Lo+24(FP), (R2, R3)

// A successful exclusive store validates the observed pair on either outcome.
// Check Lo first so a mismatch can go straight to a relaxed write-back.
#define PAIR_CAS(LOAD, STORE) \
	LOAD_CAS_ARGS; \
	retry: LOAD (R8), (R4, R5); \
	CMP R4, R0; \
	BNE mismatch; \
	CMP R5, R1; \
	BNE mismatch; \
	STORE (R2, R3), (R8), R6; \
	CBNZ R6, retry; \
	STP (R4, R5), prev_Lo+40(FP); \
	MOVD $1, R6; \
	MOVB R6, swapped+56(FP); \
	RET; \
	mismatch: STXP (R4, R5), (R8), R6; \
	CBNZ R6, retry; \
	STP (R4, R5), prev_Lo+40(FP); \
	MOVB ZR, swapped+56(FP); \
	RET

// A differing word witnesses false. Retry a lost reservation only on a match.
#define TRY_CAS(LOAD, STORE) \
	LOAD_CAS_ARGS; \
	retry: LOAD (R8), (R4, R5); \
	CMP R4, R0; \
	CCMP EQ, R5, R1, $0; \
	BNE mismatch; \
	STORE (R2, R3), (R8), R6; \
	CBNZ R6, retry; \
	MOVD $1, R6; \
	MOVB R6, swapped+40(FP); \
	RET; \
	mismatch: CLREX $15; \
	MOVB ZR, swapped+40(FP); \
	RET

TEXT ·Cas128Relaxed(SB), NOSPLIT, $0-57
	PAIR_CAS(LDXP, STXP)

TEXT ·Cas128Acquire(SB), NOSPLIT, $0-57
	PAIR_CAS(LDAXP, STXP)

TEXT ·Cas128Release(SB), NOSPLIT, $0-57
	PAIR_CAS(LDXP, STLXP)

TEXT ·Cas128AcqRel(SB), NOSPLIT, $0-57
	PAIR_CAS(LDAXP, STLXP)

TEXT ·TryRelaxed(SB), NOSPLIT, $0-41
	TRY_CAS(LDXP, STXP)

TEXT ·TryAcquire(SB), NOSPLIT, $0-41
	TRY_CAS(LDAXP, STXP)

TEXT ·TryRelease(SB), NOSPLIT, $0-41
	TRY_CAS(LDXP, STLXP)

TEXT ·TryAcqRel(SB), NOSPLIT, $0-41
	TRY_CAS(LDAXP, STLXP)
