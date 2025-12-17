//go:build amd64 || arm64

package dwcas_test

import (
	"testing"

	"code.hybscloud.com/dwcas"
)

func TestBarrierSymbolsExist(t *testing.T) {
	// Compile and execute the calls to ensure the symbols are present and wired.
	dwcas.BarrierAcquire()
	dwcas.BarrierRelease()
	dwcas.BarrierFull()
}
