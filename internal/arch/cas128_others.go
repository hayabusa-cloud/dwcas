// ©Hayabusa Cloud Co., Ltd. 2025. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

//go:build !amd64 && !arm64

package arch

func LoadRelaxed(ptr *Pair) (value Pair) {
	panic("dwcas: unsupported architecture")
}

func LoadAcquire(ptr *Pair) (value Pair) {
	panic("dwcas: unsupported architecture")
}

func StoreRelaxed(ptr *Pair, value Pair) {
	panic("dwcas: unsupported architecture")
}

func StoreRelease(ptr *Pair, value Pair) {
	panic("dwcas: unsupported architecture")
}

func TryRelaxed(ptr *Pair, old, new Pair) bool {
	panic("dwcas: unsupported architecture")
}

func TryAcquire(ptr *Pair, old, new Pair) bool {
	panic("dwcas: unsupported architecture")
}

func TryRelease(ptr *Pair, old, new Pair) bool {
	panic("dwcas: unsupported architecture")
}

func TryAcqRel(ptr *Pair, old, new Pair) bool {
	panic("dwcas: unsupported architecture")
}

func Cas128Relaxed(ptr *Pair, old, new Pair) (prev Pair, swapped bool) {
	panic("dwcas: unsupported architecture")
}

func Cas128Acquire(ptr *Pair, old, new Pair) (prev Pair, swapped bool) {
	panic("dwcas: unsupported architecture")
}

func Cas128Release(ptr *Pair, old, new Pair) (prev Pair, swapped bool) {
	panic("dwcas: unsupported architecture")
}

func Cas128AcqRel(ptr *Pair, old, new Pair) (prev Pair, swapped bool) {
	panic("dwcas: unsupported architecture")
}
