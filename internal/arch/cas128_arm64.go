// ©Hayabusa Cloud Co., Ltd. 2025. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

//go:build arm64 && !dwcas_llsc

package arch

//go:noescape
func Cas128Relaxed(ptr *uint64, oldLo, oldHi, newLo, newHi uint64) (prevLo, prevHi uint64, swapped bool)

//go:noescape
func Cas128Acquire(ptr *uint64, oldLo, oldHi, newLo, newHi uint64) (prevLo, prevHi uint64, swapped bool)

//go:noescape
func Cas128Release(ptr *uint64, oldLo, oldHi, newLo, newHi uint64) (prevLo, prevHi uint64, swapped bool)

//go:noescape
func Cas128AcqRel(ptr *uint64, oldLo, oldHi, newLo, newHi uint64) (prevLo, prevHi uint64, swapped bool)
