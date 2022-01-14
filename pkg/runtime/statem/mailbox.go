/*
Copyright 2021 The tKeel Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package statem

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

func (mb *mailbox) Empty() bool {
	mb.lock.Lock()
	defer mb.lock.Unlock()
	return mb.size == 0
}

func (mb *mailbox) Size() int {
	return mb.size
}

func (mb *mailbox) Capcity() int {
	return mb.capcity
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
