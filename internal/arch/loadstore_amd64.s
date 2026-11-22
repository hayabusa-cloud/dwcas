// ©Hayabusa Cloud Co., Ltd. 2026. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

//go:build amd64 && !dwcas_avx

#include "textflag.h"

// Use zero for both expected and desired so either outcome preserves memory.
TEXT ·LoadAcquire(SB), NOSPLIT, $0-24
	MOVQ	ptr+0(FP), DI
	XORL	AX, AX
	XORL	DX, DX
	XORL	BX, BX
	XORL	CX, CX
	LOCK
	CMPXCHG16B	(DI)
	MOVQ	AX, value_Lo+8(FP)
	MOVQ	DX, value_Hi+16(FP)
	RET

// Seed the comparison with ordinary loads. CMPXCHG16B validates both
// words together and supplies the next expected pair on a mismatch.
TEXT ·StoreRelease(SB), NOSPLIT, $0-24
	MOVQ	ptr+0(FP), DI
	MOVQ	value_Lo+8(FP), BX
	MOVQ	value_Hi+16(FP), CX
	MOVQ	(DI), AX
	MOVQ	8(DI), DX
retry:
	LOCK
	CMPXCHG16B	(DI)
	JNE	retry
	RET
