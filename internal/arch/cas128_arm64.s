// ©Hayabusa Cloud Co., Ltd. 2025. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

//go:build arm64 && !dwcas_llsc

#include "textflag.h"

#define LOAD_EXPECTED_AND_DESIRED \
	MOVD ptr+0(FP), R8; \
	LDP old_Lo+8(FP), (R0, R1); \
	LDP new_Lo+24(FP), (R2, R3); \
	MOVD R0, R4; \
	MOVD R1, R5

// Go assembly spells the relaxed instruction CASPD. The ordered encodings
// below use the same operands: X0:X1 expected, X2:X3 desired, [X8] address.
#define CASP_RELAXED CASPD (R0, R1), (R8), (R2, R3)
#define CASP_ACQUIRE WORD $0x48607d02 // CASPA
#define CASP_RELEASE WORD $0x4820fd02 // CASPL
#define CASP_ACQ_REL WORD $0x4860fd02 // CASPAL

// CASP returns the observed pair in R0:R1. R4:R5 holds the original expected pair.
#define RETURN_MATCH(slot) \
	CMP R0, R4; \
	CCMP EQ, R1, R5, $0; \
	CSET EQ, R6; \
	MOVB R6, slot; \
	RET

TEXT ·Cas128Relaxed(SB), NOSPLIT, $0-57
	LOAD_EXPECTED_AND_DESIRED
	CASP_RELAXED
	STP (R0, R1), prev_Lo+40(FP)
	RETURN_MATCH(swapped+56(FP))

TEXT ·Cas128Acquire(SB), NOSPLIT, $0-57
	LOAD_EXPECTED_AND_DESIRED
	CASP_ACQUIRE
	STP (R0, R1), prev_Lo+40(FP)
	RETURN_MATCH(swapped+56(FP))

TEXT ·Cas128Release(SB), NOSPLIT, $0-57
	LOAD_EXPECTED_AND_DESIRED
	CASP_RELEASE
	STP (R0, R1), prev_Lo+40(FP)
	RETURN_MATCH(swapped+56(FP))

TEXT ·Cas128AcqRel(SB), NOSPLIT, $0-57
	LOAD_EXPECTED_AND_DESIRED
	CASP_ACQ_REL
	STP (R0, R1), prev_Lo+40(FP)
	RETURN_MATCH(swapped+56(FP))

TEXT ·TryRelaxed(SB), NOSPLIT, $0-41
	LOAD_EXPECTED_AND_DESIRED
	CASP_RELAXED
	RETURN_MATCH(swapped+40(FP))

TEXT ·TryAcquire(SB), NOSPLIT, $0-41
	LOAD_EXPECTED_AND_DESIRED
	CASP_ACQUIRE
	RETURN_MATCH(swapped+40(FP))

TEXT ·TryRelease(SB), NOSPLIT, $0-41
	LOAD_EXPECTED_AND_DESIRED
	CASP_RELEASE
	RETURN_MATCH(swapped+40(FP))

TEXT ·TryAcqRel(SB), NOSPLIT, $0-41
	LOAD_EXPECTED_AND_DESIRED
	CASP_ACQ_REL
	RETURN_MATCH(swapped+40(FP))
