//go:build !dwcasdebug

package dwcas

// checkAligned is compiled out in normal builds.
func checkAligned(_ *Uint128) {}
