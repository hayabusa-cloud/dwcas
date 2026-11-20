// ©Hayabusa Cloud Co., Ltd. 2026. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

//go:build arm64 && dwcas_llsc

#include "textflag.h"

// A successful write-back validates the two words as one coherent sample.
#define LOAD_PAIR(LOAD) \
	MOVD ptr+0(FP), R8; \
	retry: LOAD (R8), (R0, R1); \
	STXP (R0, R1), (R8), R2; \
	CBNZ R2, retry; \
	STP (R0, R1), value_Lo+8(FP); \
	RET

#define STORE_PAIR(STORE) \
	MOVD ptr+0(FP), R8; \
	LDP value_Lo+8(FP), (R0, R1); \
	retry: LDXP (R8), (ZR, R3); \
	STORE (R0, R1), (R8), R2; \
	CBNZ R2, retry; \
	RET

TEXT ·LoadRelaxed(SB), NOSPLIT, $0-24
	LOAD_PAIR(LDXP)

TEXT ·LoadAcquire(SB), NOSPLIT, $0-24
	LOAD_PAIR(LDAXP)

TEXT ·StoreRelaxed(SB), NOSPLIT, $0-24
	STORE_PAIR(STXP)

TEXT ·StoreRelease(SB), NOSPLIT, $0-24
	STORE_PAIR(STLXP)
