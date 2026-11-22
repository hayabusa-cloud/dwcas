// ©Hayabusa Cloud Co., Ltd. 2026. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

//go:build unix && (amd64 || arm64)

package dwcas_test

import "syscall"

func processCPUNanos() (int64, error) {
	var usage syscall.Rusage
	if err := syscall.Getrusage(syscall.RUSAGE_SELF, &usage); err != nil {
		return 0, err
	}
	return usage.Utime.Nano() + usage.Stime.Nano(), nil
}
