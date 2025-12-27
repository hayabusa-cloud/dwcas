// ©Hayabusa Cloud Co., Ltd. 2025. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

//go:build amd64

#include "textflag.h"

// Pure compiler barriers for amd64.
//
// On x86-64 (TSO), LOCK CMPXCHG16B already provides a full CPU barrier.
// These functions exist solely to prevent compile-time reordering.
// The opaque assembly call boundary forces the Go compiler to complete
// all pending memory operations before the call and not reorder across it.

TEXT ·BarrierAcquire(SB), NOSPLIT, $0-0
	RET

TEXT ·BarrierRelease(SB), NOSPLIT, $0-0
	RET

TEXT ·BarrierFull(SB), NOSPLIT, $0-0
	RET
