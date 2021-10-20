package entities

import (
	"errors"
	"sync"
)

type mailbox struct {
	size     int
	headInx  int
	capcity  int
	msgQueue []Message

	lock *sync.Mutex
}

var (
	errMailboxOverflow      = errors.New("mailbox.queue full")
	errMailboxInvalidResize = errors.New("resize invalid size")
)

func newMailbox(capcity int) *mailbox {
	return &mailbox{
		size:     0,
		headInx:  0,
		capcity:  capcity,
		lock:     &sync.Mutex{},
		msgQueue: make([]Message, capcity),
	}
}

func (mb *mailbox) Get() Message {
	var msg Message

	mb.lock.Lock()
	defer mb.lock.Unlock()

	if mb.size > 0 {
		mb.size--
		msg = mb.msgQueue[mb.headInx]
		mb.headInx = (mb.headInx + 1) % mb.capcity
	}

	return msg
}

func (mb *mailbox) Put(msg Message) error {
	mb.lock.Lock()
	defer mb.lock.Unlock()

	if mb.capcity == mb.size {
		return errMailboxOverflow
	}

	index := (mb.headInx + mb.size) % mb.capcity
	mb.msgQueue[index] = msg

	mb.size++

	return nil
}

func (mb *mailbox) Size() int {
	return mb.size
}

func (mb *mailbox) Resize(capcity int) error {
	mb.lock.Lock()
	defer mb.lock.Unlock()

	if capcity < mb.capcity {
		return errMailboxInvalidResize
	} else if capcity == mb.capcity {
		// do not resize.
		return nil
	}

	msgs := make([]Message, mb.size, capcity)
	for index := 0; index < mb.size; index++ {
		msgs[index] = mb.msgQueue[(mb.headInx+index)%mb.capcity]
	}

	mb.headInx = 0
	mb.capcity = capcity
	mb.msgQueue = msgs

	return nil
}
