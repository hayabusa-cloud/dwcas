// ©Hayabusa Cloud Co., Ltd. 2026. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

//go:build amd64 && dwcas_avx

#include "textflag.h"

// The receiver is 16-byte aligned. The ABI stack slot may be unaligned.
// Ordinary write-back memory supplies acquire and release ordering.
TEXT ·LoadAcquire(SB), NOSPLIT, $0-24
	MOVQ	ptr+0(FP), DI
	VMOVDQA	(DI), X0
	VMOVDQU	X0, value_Lo+8(FP)
	RET

TEXT ·StoreRelease(SB), NOSPLIT, $0-24
	MOVQ	ptr+0(FP), DI
	VMOVDQU	value_Lo+8(FP), X0
	VMOVDQA	X0, (DI)
	RET
