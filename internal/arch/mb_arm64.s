// ©Hayabusa Cloud Co., Ltd. 2025. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

//go:build arm64

#include "textflag.h"

TEXT ·BarrierAcquire(SB), NOSPLIT, $0-0
	// DMB ISHLD
	DMB	$0x9
	RET

TEXT ·BarrierRelease(SB), NOSPLIT, $0-0
	// DMB ISHST
	DMB	$0xA
	RET

TEXT ·BarrierFull(SB), NOSPLIT, $0-0
	// DMB ISH
	DMB	$0xB
	RET
