// ©Hayabusa Cloud Co., Ltd. 2026. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

//go:build arm64 && !dwcas_llsc

#include "textflag.h"

// With FEAT_LSE2 and Normal Inner/Outer Write-Back memory, the aligned
// 16-byte access to (R8) is single-copy atomic.
TEXT ·LoadRelaxed(SB), NOSPLIT, $0-24
	MOVD	ptr+0(FP), R8
	LDP	(R8), (R0, R1)
	STP	(R0, R1), value_Lo+8(FP)
	RET

TEXT ·LoadAcquire(SB), NOSPLIT, $0-24
	MOVD	ptr+0(FP), R8
	LDP	(R8), (R0, R1)
	DMB	$0x9 // ISHLD
	STP	(R0, R1), value_Lo+8(FP)
	RET

TEXT ·StoreRelaxed(SB), NOSPLIT, $0-24
	MOVD	ptr+0(FP), R8
	LDP	value_Lo+8(FP), (R0, R1)
	STP	(R0, R1), (R8)
	RET

TEXT ·StoreRelease(SB), NOSPLIT, $0-24
	MOVD	ptr+0(FP), R8
	LDP	value_Lo+8(FP), (R0, R1)
	DMB	$0xb // ISH orders both earlier reads and earlier writes.
	STP	(R0, R1), (R8)
	RET
