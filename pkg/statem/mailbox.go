package statem

import (
	"errors"
	"sync"
)

type mailbox struct {
	size     int
	headInx  int
	capacity int
	msgQueue []Message

	lock *sync.Mutex
}

var (
	errMailboxOverflow      = errors.New("mailbox.queue full")
	errMailboxInvalidResize = errors.New("resize invalid size")
)

func newMailbox(capacity int) *mailbox {
	return &mailbox{
		size:     0,
		headInx:  0,
		capacity: capacity,
		lock:     &sync.Mutex{},
		msgQueue: make([]Message, capacity),
	}
}

func (mb *mailbox) Get() Message {
	var msg Message

	mb.lock.Lock()
	defer mb.lock.Unlock()

	if mb.size > 0 {
		mb.size--
		msg = mb.msgQueue[mb.headInx]
		mb.headInx = (mb.headInx + 1) % mb.capacity
	}

	return msg
}

func (mb *mailbox) Put(msg Message) error {
	mb.lock.Lock()
	defer mb.lock.Unlock()

	if mb.capacity == mb.size {
		return errMailboxOverflow
	}

	index := (mb.headInx + mb.size) % mb.capacity
	mb.msgQueue[index] = msg

	mb.size++

	return nil
}

func (mb *mailbox) Size() int {
	return mb.size
}

func (mb *mailbox) Resize(capacity int) error {
	mb.lock.Lock()
	defer mb.lock.Unlock()

	if capacity < mb.capacity {
		return errMailboxInvalidResize
	} else if capacity == mb.capacity {
		// do not resize.
		return nil
	}

	msgs := make([]Message, mb.size, capacity)
	for index := 0; index < mb.size; index++ {
		msgs[index] = mb.msgQueue[(mb.headInx+index)%mb.capacity]
	}

	mb.headInx = 0
	mb.capacity = capacity
	mb.msgQueue = msgs

	return nil
}
