// ©Hayabusa Cloud Co., Ltd. 2026. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package dwcas_test

import "code.hybscloud.com/dwcas"

var casMethods = [...]struct {
	name string
	try  func(*dwcas.Uint128, dwcas.Uint128, dwcas.Uint128) bool
	fn   func(*dwcas.Uint128, dwcas.Uint128, dwcas.Uint128) (dwcas.Uint128, bool)
}{
	{"Relaxed", (*dwcas.Uint128).TryRelaxed, (*dwcas.Uint128).Relaxed},
	{"Acquire", (*dwcas.Uint128).TryAcquire, (*dwcas.Uint128).Acquire},
	{"Release", (*dwcas.Uint128).TryRelease, (*dwcas.Uint128).Release},
	{"AcqRel", (*dwcas.Uint128).TryAcqRel, (*dwcas.Uint128).AcqRel},
}
