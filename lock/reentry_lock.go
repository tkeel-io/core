package lock

import (
	"sync"

	"errors"

	"go.uber.org/atomic"
)

const defaultStackDepth = 12

var ErrMaxStackDepthExceeded = errors.New("maximum stack depth exceeded")

// ReEntryLock re-entry lock.
type ReEntryLock struct {
	methodLock    *sync.Mutex
	requestLock   *sync.Mutex
	activeRequest *string
	stackDepth    *atomic.Int32
	maxStackDepth int32
}

func NewReEntryLock(maxStackDepth int32) *ReEntryLock {
	if 0 >= maxStackDepth {
		maxStackDepth = defaultStackDepth
	}
	return &ReEntryLock{
		methodLock:    &sync.Mutex{},
		requestLock:   &sync.Mutex{},
		activeRequest: nil,
		stackDepth:    atomic.NewInt32(int32(0)),
		maxStackDepth: maxStackDepth,
	}
}

func (a *ReEntryLock) Lock(requestID *string) error {
	currentRequest := a.getCurrentID()

	if a.stackDepth.Load() == a.maxStackDepth {
		return ErrMaxStackDepthExceeded
	}

	// sync.Mutex不是递归锁，所以这一可重入针对的是同一个requestID的请求
	if currentRequest == nil || *currentRequest != *requestID {
		// 如果不是同一个requestID，进来了也会阻塞
		a.methodLock.Lock()
		a.setCurrentID(requestID)
		a.stackDepth.Inc()
	} else {
		a.stackDepth.Inc()
	}

	return nil
}

func (a *ReEntryLock) Unlock() {
	a.stackDepth.Dec()
	if a.stackDepth.Load() == 0 {
		a.clearCurrentID()
		a.methodLock.Unlock()
	}
}

func (a *ReEntryLock) getCurrentID() *string {
	a.requestLock.Lock()
	defer a.requestLock.Unlock()

	return a.activeRequest
}

func (a *ReEntryLock) setCurrentID(id *string) {
	a.requestLock.Lock()
	defer a.requestLock.Unlock()

	a.activeRequest = id
}

func (a *ReEntryLock) clearCurrentID() {
	a.requestLock.Lock()
	defer a.requestLock.Unlock()

	a.activeRequest = nil
}
