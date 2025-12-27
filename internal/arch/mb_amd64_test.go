// ©Hayabusa Cloud Co., Ltd. 2025. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

//go:build amd64

package arch

import (
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestBarriersSmoke(t *testing.T) {
	BarrierAcquire()
	BarrierRelease()
	BarrierFull()
}

func TestBarriersTightLoop(t *testing.T) {
	for i := 0; i < 1<<18; i++ {
		BarrierAcquire()
		BarrierRelease()
		BarrierFull()
	}
}

func TestBarriersConcurrentExercise(t *testing.T) {
	workers := runtime.GOMAXPROCS(0) * 4
	if workers < 4 {
		workers = 4
	}

	iters := 1 << 15
	var wg sync.WaitGroup
	wg.Add(workers)
	for w := 0; w < workers; w++ {
		go func() {
			defer wg.Done()
			for i := 0; i < iters; i++ {
				BarrierAcquire()
				BarrierRelease()
				BarrierFull()
			}
		}()
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("barrier exercise timed out")
	}
}

func TestBarrierUsagePatternNoLitmus(t *testing.T) {
	// This test intentionally does not attempt to validate CPU ordering.
	// It only ensures the barrier functions are executed in a typical publish/
	// consume control-flow shape without relying on flaky litmus behavior.

	var data atomic.Uint64
	var flag atomic.Uint32
	const want = 0xfeedbeef

	done := make(chan struct{})
	go func() {
		data.Store(want)
		BarrierRelease()
		flag.Store(1)
		close(done)
	}()

	deadline := time.Now().Add(2 * time.Second)
	for flag.Load() == 0 {
		if time.Now().After(deadline) {
			t.Fatal("timed out waiting for publisher")
		}
		runtime.Gosched()
	}

	BarrierAcquire()
	if got := data.Load(); got != want {
		t.Fatalf("unexpected data: got=%d want=%d", got, want)
	}
	<-done
}
