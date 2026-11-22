// ©Hayabusa Cloud Co., Ltd. 2026. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

//go:build arm64

package arch

//go:noescape
func Cas128Relaxed(ptr *Pair, old, new Pair) (prev Pair, swapped bool)

//go:noescape
func Cas128Acquire(ptr *Pair, old, new Pair) (prev Pair, swapped bool)

//go:noescape
func Cas128Release(ptr *Pair, old, new Pair) (prev Pair, swapped bool)

//go:noescape
func Cas128AcqRel(ptr *Pair, old, new Pair) (prev Pair, swapped bool)

//go:noescape
func TryRelaxed(ptr *Pair, old, new Pair) (swapped bool)

//go:noescape
func TryAcquire(ptr *Pair, old, new Pair) (swapped bool)

//go:noescape
func TryRelease(ptr *Pair, old, new Pair) (swapped bool)

//go:noescape
func TryAcqRel(ptr *Pair, old, new Pair) (swapped bool)
