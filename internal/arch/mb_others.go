// ©Hayabusa Cloud Co., Ltd. 2025. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

//go:build !amd64 && !arm64

package arch

func BarrierAcquire() {
	panic("dwcas: unsupported architecture")
}

func BarrierRelease() {
	panic("dwcas: unsupported architecture")
}

func BarrierFull() {
	panic("dwcas: unsupported architecture")
}
