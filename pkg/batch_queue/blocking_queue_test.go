package batchqueue

import (
	"fmt"
	"testing"
	"time"

	"github.com/smartystreets/assertions/should"
	"github.com/smartystreets/gunit"
)

func TestBlockingQueueUnit(t *testing.T) {
	//gunit.RunSequential(new(BlockingQueueUnit), t)
	gunit.Run(new(BlockingQueueUnit), t)
}

type BlockingQueueUnit struct {
	*gunit.Fixture
}

func (this *BlockingQueueUnit) Setup() {
	this.So(nil, should.BeNil)
}

func (this *BlockingQueueUnit) TestBlockingQueue() {
	q := NewBlockingQueue(10)

	this.AssertEqual(0, q.Size())
	this.AssertEqual(nil, q.Poll())
	this.AssertEqual(nil, q.Peek())
	this.AssertEqual(nil, q.PeekLast())

	q.Put("test")
	this.AssertEqual(1, q.Size())

	this.AssertEqual("test", q.Peek())
	this.AssertEqual("test", q.PeekLast())
	this.AssertEqual(1, q.Size())

	this.AssertEqual("test", q.Take())
	this.AssertEqual(nil, q.Peek())
	this.AssertEqual(nil, q.PeekLast())
	this.AssertEqual(0, q.Size())

	ch := make(chan string)

	go func() {
		// Stays blocked until item is available
		ch <- q.Take().(string)
	}()

	time.Sleep(100 * time.Millisecond)

	select {
	case <-ch:
		this.Error("Should not have had a value at this point")
	default:
		// Good, no value yet
	}

	q.Put("test-2")

	x := <-ch
	this.AssertEqual("test-2", x)

	// Fill the queue
	for i := 0; i < 10; i++ {
		q.Put(fmt.Sprintf("i-%d", i))
		this.AssertEqual(i+1, q.Size())
	}

	for i := 0; i < 10; i++ {
		this.AssertEqual(fmt.Sprintf("i-%d", i), q.Take())
	}

	close(ch)
}

func (this *BlockingQueueUnit) TestBlockingQueueWaitWhenFull() {
	q := NewBlockingQueue(3)

	q.Put("test-1")
	q.Put("test-2")
	q.Put("test-3")
	this.AssertEqual(3, q.Size())
	this.AssertEqual("test-1", q.Peek())
	this.AssertEqual("test-3", q.PeekLast())

	ch := make(chan bool)

	go func() {
		// Stays blocked until space is available
		q.Put("test-4")
		ch <- true
	}()

	time.Sleep(100 * time.Millisecond)

	select {
	case <-ch:
		this.Error("Should not have had a value at this point")
	default:
		// Good, no value yet
	}

	this.AssertEqual("test-1", q.Poll())

	// Now the go-routine should have completed
	<-ch
	this.AssertEqual(3, q.Size())

	this.AssertEqual("test-2", q.Take())
	this.AssertEqual("test-3", q.Take())
	this.AssertEqual("test-4", q.Take())

	close(ch)
}

func (this *BlockingQueueUnit) TestBlockingQueueIterator() {
	q := NewBlockingQueue(10)

	q.Put(1)
	q.Put(2)
	q.Put(3)
	this.AssertEqual(3, q.Size())

	i := 1
	for it := q.Iterator(); it.HasNext(); {
		this.AssertEqual(i, it.Next())
		i++
	}
}
