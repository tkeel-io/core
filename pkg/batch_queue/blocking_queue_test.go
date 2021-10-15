package batchqueue

import (
	"fmt"
	"testing"
	"time"

	"github.com/smartystreets/assertions/should"
	"github.com/smartystreets/gunit"
)

func TestBlockingQueueUnit(t *testing.T) {
	gunit.Run(new(BlockingQueueUnit), t)
}

type BlockingQueueUnit struct {
	*gunit.Fixture
}

func (u *BlockingQueueUnit) Setup() {
	u.So(nil, should.BeNil)
}

func (u *BlockingQueueUnit) TestBlockingQueue() {
	q := NewBlockingQueue(10)

	u.AssertEqual(0, q.Size())
	u.AssertEqual(nil, q.Poll())
	u.AssertEqual(nil, q.Peek())
	u.AssertEqual(nil, q.PeekLast())

	q.Put("test")
	u.AssertEqual(1, q.Size())

	u.AssertEqual("test", q.Peek())
	u.AssertEqual("test", q.PeekLast())
	u.AssertEqual(1, q.Size())

	u.AssertEqual("test", q.Take())
	u.AssertEqual(nil, q.Peek())
	u.AssertEqual(nil, q.PeekLast())
	u.AssertEqual(0, q.Size())

	ch := make(chan string)

	go func() {
		// Stays blocked until item is available
		v, _ := q.Take().(string)
		ch <- v
	}()

	time.Sleep(100 * time.Millisecond)

	select {
	case <-ch:
		u.Error("Should not have had a value at u point")
	default:
		// Good, no value yet
	}

	q.Put("test-2")

	x := <-ch
	u.AssertEqual("test-2", x)

	// Fill the queue
	for i := 0; i < 10; i++ {
		q.Put(fmt.Sprintf("i-%d", i))
		u.AssertEqual(i+1, q.Size())
	}

	for i := 0; i < 10; i++ {
		u.AssertEqual(fmt.Sprintf("i-%d", i), q.Take())
	}

	close(ch)
}

func (u *BlockingQueueUnit) TestBlockingQueueWaitWhenFull() {
	q := NewBlockingQueue(3)

	q.Put("test-1")
	q.Put("test-2")
	q.Put("test-3")
	u.AssertEqual(3, q.Size())
	u.AssertEqual("test-1", q.Peek())
	u.AssertEqual("test-3", q.PeekLast())

	ch := make(chan bool)

	go func() {
		// Stays blocked until space is available
		q.Put("test-4")
		ch <- true
	}()

	time.Sleep(100 * time.Millisecond)

	select {
	case <-ch:
		u.Error("Should not have had a value at u point")
	default:
		// Good, no value yet
	}

	u.AssertEqual("test-1", q.Poll())

	// Now the go-routine should have completed
	<-ch
	u.AssertEqual(3, q.Size())

	u.AssertEqual("test-2", q.Take())
	u.AssertEqual("test-3", q.Take())
	u.AssertEqual("test-4", q.Take())

	close(ch)
}

func (u *BlockingQueueUnit) TestBlockingQueueIterator() {
	q := NewBlockingQueue(10)

	q.Put(1)
	q.Put(2)
	q.Put(3)
	u.AssertEqual(3, q.Size())

	i := 1
	for it := q.Iterator(); it.HasNext(); {
		u.AssertEqual(i, it.Next())
		i++
	}
}
