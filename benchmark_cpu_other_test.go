// ©Hayabusa Cloud Co., Ltd. 2026. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

//go:build !unix && (amd64 || arm64)

package dwcas_test

func processCPUNanos() (int64, error) { return 0, nil }
