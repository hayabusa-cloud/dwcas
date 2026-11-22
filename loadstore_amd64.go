// ©Hayabusa Cloud Co., Ltd. 2026. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

//go:build amd64

package dwcas

import "code.hybscloud.com/dwcas/internal/arch"

// Both amd64 backends already give pair loads acquire ordering and pair
// stores release ordering.

// LoadRelaxed atomically loads the pair with relaxed ordering.
func (p *Uint128) LoadRelaxed() Uint128 {
	checkAligned(p)
	return arch.LoadAcquire((*arch.Pair)(p))
}

// StoreRelaxed atomically stores value with relaxed ordering.
func (p *Uint128) StoreRelaxed(value Uint128) {
	checkAligned(p)
	arch.StoreRelease((*arch.Pair)(p), value)
}
