package dwcas

import "code.hybscloud.com/dwcas/internal/arch"

// BarrierAcquire orders earlier loads before later loads and stores for ordinary
// accesses to write-back shared memory. On arm64 it emits DMB ISHLD.
// On amd64 it is a compiler barrier.
func BarrierAcquire() {
	arch.BarrierAcquire()
}

// BarrierRelease orders earlier stores before later stores for ordinary
// accesses to write-back shared memory. On arm64 it emits DMB ISHST.
// On amd64 it is a compiler barrier.
//
// When publication must also order earlier loads, use a release store
// such as [Uint128.StoreRelease] or a successful release CAS.
func BarrierRelease() {
	arch.BarrierRelease()
}

// BarrierFull emits DMB ISH on arm64 and acts as a compiler barrier on amd64.
// On amd64, the CPU may reorder a load with an earlier store to a different
// address.
func BarrierFull() {
	arch.BarrierFull()
}
