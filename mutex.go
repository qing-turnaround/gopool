package pool

import (
	"runtime"
	"sync"
	"sync/atomic"
)

// Ants 锁
type spinLock uint32

const maxBackoff = 16

func (sl *spinLock) Lock() {
	backoff := 1
	for !atomic.CompareAndSwapUint32((*uint32)(sl), 0, 1) {
		for i := 0; i < backoff; i++ {
			runtime.Gosched()
		}
		if backoff < maxBackoff {
			backoff <<= 1
		}
	}
}

func (sl *spinLock) Unlock() {
	atomic.StoreUint32((*uint32)(sl), 0)
}

func newSpinLock() sync.Locker {
	return new(spinLock)
}

func newMutexLock() sync.Locker {
	return new(sync.Mutex)
}
