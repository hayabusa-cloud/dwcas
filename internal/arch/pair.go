// ©Hayabusa Cloud Co., Ltd. 2026. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package arch

// Pair is an unnamed struct so dwcas.Uint128 values pass directly across
// the assembly boundary. Lo occupies the first eight bytes.
type Pair = struct {
	Lo uint64
	Hi uint64
}
