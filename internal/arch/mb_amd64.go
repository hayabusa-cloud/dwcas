// ©Hayabusa Cloud Co., Ltd. 2025. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

//go:build amd64

package arch

import "sync/atomic"

// compilerBarrierSink is used to force a compiler barrier without emitting an
// expensive CPU fence on amd64.
//
// An atomic load is sufficient to prevent compile-time reordering across the
// call site, while relying on amd64's already-strong CPU ordering.
var compilerBarrierSink uint32

func BarrierAcquire() {
	_ = atomic.LoadUint32(&compilerBarrierSink)
}

func BarrierRelease() {
	_ = atomic.LoadUint32(&compilerBarrierSink)
}

func BarrierFull() {
	_ = atomic.LoadUint32(&compilerBarrierSink)
}
