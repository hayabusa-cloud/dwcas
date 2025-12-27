// ©Hayabusa Cloud Co., Ltd. 2025. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

//go:build amd64

package arch

// Pure compiler barriers for amd64.
//
// On x86-64 (TSO), LOCK CMPXCHG16B already provides a full CPU barrier.
// These functions exist solely to prevent compile-time reordering.

//go:noinline
func BarrierAcquire()

//go:noinline
func BarrierRelease()

//go:noinline
func BarrierFull()
