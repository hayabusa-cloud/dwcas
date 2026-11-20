// ©Hayabusa Cloud Co., Ltd. 2026. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

//go:build amd64 && dwcas_avx

#include "textflag.h"

// The receiver is 16-byte aligned. The ABI stack slot may be unaligned.
// Ordinary write-back memory supplies acquire and release ordering.
#define LOAD_PAIR \
	MOVQ ptr+0(FP), DI; \
	VMOVDQA (DI), X0; \
	VMOVDQU X0, value_Lo+8(FP); \
	RET

#define STORE_PAIR \
	MOVQ ptr+0(FP), DI; \
	VMOVDQU value_Lo+8(FP), X0; \
	VMOVDQA X0, (DI); \
	RET

TEXT ·LoadRelaxed(SB), NOSPLIT, $0-24
	LOAD_PAIR

TEXT ·LoadAcquire(SB), NOSPLIT, $0-24
	LOAD_PAIR

TEXT ·StoreRelaxed(SB), NOSPLIT, $0-24
	STORE_PAIR

TEXT ·StoreRelease(SB), NOSPLIT, $0-24
	STORE_PAIR
