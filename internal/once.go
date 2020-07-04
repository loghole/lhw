package internal

import (
	"sync"
	"sync/atomic"
)

type Once struct {
	done uint32
}

func (o *Once) DoWG(wg *sync.WaitGroup, f func()) {
	if atomic.CompareAndSwapUint32(&o.done, 0, 1) {
		f()
		atomic.StoreUint32(&o.done, 0)
	}

	wg.Done()
}
