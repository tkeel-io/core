/*
 * Copyright (C) 2019 Yunify, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this work except in compliance with the License.
 * You may obtain a copy of the License in the LICENSE file, or at:
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

/**
SinkBatch
1. 线程池，限定最大并发数
2. 累计 message, 定时提交
3. 可以 Flush，手动触发提交
*/

package batchqueue

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/tkeel-io/core/pkg/logger"
)

var log = logger.NewLogger("core.batch-queue")

type sinkBatchState int

const (
	sinkBatchInit sinkBatchState = iota
	sinkBatchReady
	sinkBatchClosing
	sinkBatchClosed
)

type processState int

const (
	processIdle processState = iota
	processInProgress
)

type BatchSink interface {
	Send(ctx context.Context, msg interface{}) error
	Flush(ctx context.Context) error
	Close()
}

type sendRequest struct {
	ctx context.Context
	msg interface{}
}

type closeRequest struct {
	waitGroup *sync.WaitGroup
}

type flushRequest struct {
	waitGroup *sync.WaitGroup
	err       error
}

type pendingItem struct {
	sync.Mutex
	sequenceID uint64
	batchData  []interface{}
	callback   []CallbackFn
	status     processState
	err        error
}

func (pending *pendingItem) GetSequenceID() uint64 {
	return pending.sequenceID
}

func (pending *pendingItem) Callback() {
	// lock the pending item
	pending.Lock()
	defer pending.Unlock()
	for _, fn := range pending.callback {
		fn(pending.sequenceID, pending.err)
	}
}

func (pending *pendingItem) Release() {
	pending.batchData = nil
}

type CallbackFn func(sequenceID uint64, e error)

type ProcessFn func(msgs []interface{}) (err error)

type batchSink struct {
	batchBuilder     *BatchBuilder
	batchFlushTicker *time.Ticker

	// Channel where app is posting messages to be published
	eventsChan chan interface{}

	pendingQueue BlockingQueue

	processFn ProcessFn
	state     sinkBatchState
	sinkName  string
	//	lock      sync.Mutex

	conf *Config

	sendCnt int64

	ctx context.Context
}

type Config struct {
	Name     string
	DoSinkFn ProcessFn
	// BatchingMaxMessages set the maximum number of messages permitted in a batch. (default: 1000)
	MaxBatching int
	// MaxPendingMessages set the max size of the queue.
	MaxPendingMessages uint
	// BatchingMaxFlushDelay set the time period within which the messages sent will be batched (default: 10ms)
	BatchingMaxFlushDelay time.Duration
}

func (c *Config) GetBatchingMaxFlushDelay() time.Duration {
	if c.BatchingMaxFlushDelay != 0 {
		c.BatchingMaxFlushDelay = defaultBatchingMaxFlushDelay
	}
	return c.BatchingMaxFlushDelay
}

func (c *Config) GetMaxPendingMessages() int {
	if c.MaxPendingMessages != 0 {
		c.MaxPendingMessages = defaultMaxPendingMessages
	}
	return int(c.MaxPendingMessages)
}

func (c *Config) GetMaxBatching() uint {
	return uint(c.MaxBatching)
}

const (
	defaultMaxPendingMessages    = 5
	defaultBatchingMaxFlushDelay = 10 * time.Millisecond
)

func NewBatchSink(ctx context.Context, conf *Config) (BatchSink, error) {
	if nil == conf {
		return nil, errors.New("reconfiguration required")
	}

	p := &batchSink{
		ctx:              ctx,
		conf:             conf,
		sinkName:         conf.Name,
		processFn:        conf.DoSinkFn,
		state:            sinkBatchInit,
		eventsChan:       make(chan interface{}, 1),
		batchBuilder:     NewBatchBuilder(conf.GetMaxBatching()),
		pendingQueue:     NewBlockingQueue(conf.GetMaxPendingMessages()),
		batchFlushTicker: time.NewTicker(conf.GetBatchingMaxFlushDelay()),
	}

	p.state = sinkBatchReady

	go p.runEventsLoop()

	return p, nil
}

func (p *batchSink) runEventsLoop() {
	for {
		select {
		case i := <-p.eventsChan:
			switch v := i.(type) {
			case *sendRequest:
				p.internalSend(v)
			case *flushRequest:
				p.internalFlush(v)
			case *closeRequest:
				p.internalClose(v)
				return
			}

		case <-p.batchFlushTicker.C:
			p.internalFlushCurrentBatch()
		}
	}
}

func (p *batchSink) internalSend(request *sendRequest) {
	msg := request.msg

	isFull := p.batchBuilder.Add(msg)
	if isFull {
		// The current batch is full then flush it.
		p.internalFlushCurrentBatch()
	}
	p.sendCnt++
}

func (p *batchSink) internalFlushCurrentBatch() {
	batchData, sequenceID := p.batchBuilder.Flush()
	if batchData == nil {
		return
	}

	item := pendingItem{
		batchData:  batchData,
		sequenceID: sequenceID,
		callback:   []CallbackFn{},
		status:     processInProgress,
	}
	p.pendingQueue.Put(&item)

	go func(item *pendingItem) {
		p.callbackReceipt(item, p.processFn(item.batchData))
	}(&item)
}

func (p *batchSink) internalFlush(fr *flushRequest) {
	p.internalFlushCurrentBatch()

	pi, ok := p.pendingQueue.PeekLast().(*pendingItem)
	if !ok {
		fr.waitGroup.Done()
		return
	}

	// lock the pending request while adding requests
	// since the ReceivedSendReceipt func iterates over this list
	pi.Lock()
	pi.callback = append(pi.callback, func(sequenceID uint64, e error) {
		fr.err = e
		fr.waitGroup.Done()
	})
	pi.Unlock()
}

func (p *batchSink) internalClose(req *closeRequest) {
	defer req.waitGroup.Done()
	if p.state != sinkBatchReady {
		return
	}

	p.state = sinkBatchClosing

	p.state = sinkBatchClosed
	p.batchFlushTicker.Stop()

	wg := sync.WaitGroup{}
	wg.Add(1)
	fr := &flushRequest{&wg, nil}
	p.internalFlush(fr)
	wg.Wait()
}

func (p *batchSink) callbackReceipt(item *pendingItem, err error) {
	log.Debugf("Response receipt:%d", item.sequenceID)
	item.status = processIdle
	item.err = err
	p.sendCnt -= int64(len(item.batchData))

	for {
		pi, ok := p.pendingQueue.Peek().(*pendingItem)

		if !ok {
			break
		}
		if pi.status == processInProgress {
			// p.log.Bg().Debug("Response receipt unexpected",
			//	logf.Any("pendingSequenceId", pi.sequenceID),
			//	logf.Any("responseSequenceId", item.sequenceID))
			break
		}

		// We can remove the item which is done
		p.pendingQueue.Poll()

		// Trigger the callback and release item
		pi.Callback()
		pi.Release()
	}
}

func (p *batchSink) Send(ctx context.Context, msg interface{}) error {
	var err error
	sr := &sendRequest{
		ctx: ctx,
		msg: msg,
	}
	p.internalSend(sr)

	return err
}

func (p *batchSink) Flush(ctx context.Context) error {
	wg := sync.WaitGroup{}
	wg.Add(1)

	fr := &flushRequest{&wg, nil}
	p.eventsChan <- fr

	wg.Wait()
	return fr.err
}

func (p *batchSink) Close() {
	if p.state != sinkBatchReady {
		// SinkBench is closing
		return
	}

	wg := sync.WaitGroup{}
	wg.Add(1)

	cp := &closeRequest{&wg}
	p.eventsChan <- cp

	wg.Wait()
}
