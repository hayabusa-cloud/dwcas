// ©Hayabusa Cloud Co., Ltd. 2026. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

//go:build amd64

package arch

// LOCK CMPXCHG16B implements all four orderings. Each result type shares one
// assembly entry. The named pair returns keep both Go forwarding layers inlineable.

//go:noescape
func Cas128Release(ptr *Pair, old, new Pair) (prev Pair, swapped bool)

//go:noescape
func TryRelease(ptr *Pair, old, new Pair) (swapped bool)

func Cas128Relaxed(ptr *Pair, old, new Pair) (prev Pair, swapped bool) {
	prev, swapped = Cas128Release(ptr, old, new)
	return
}

func Cas128Acquire(ptr *Pair, old, new Pair) (prev Pair, swapped bool) {
	prev, swapped = Cas128Release(ptr, old, new)
	return
}

func Cas128AcqRel(ptr *Pair, old, new Pair) (prev Pair, swapped bool) {
	prev, swapped = Cas128Release(ptr, old, new)
	return
}

func TryRelaxed(ptr *Pair, old, new Pair) bool {
	return TryRelease(ptr, old, new)
}

func TryAcquire(ptr *Pair, old, new Pair) bool {
	return TryRelease(ptr, old, new)
}

func TryAcqRel(ptr *Pair, old, new Pair) bool {
	return TryRelease(ptr, old, new)
}
