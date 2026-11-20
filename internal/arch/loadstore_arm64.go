// ©Hayabusa Cloud Co., Ltd. 2026. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

//go:build arm64

package arch

//go:noescape
func LoadRelaxed(ptr *Pair) (value Pair)

//go:noescape
func StoreRelaxed(ptr *Pair, value Pair)
