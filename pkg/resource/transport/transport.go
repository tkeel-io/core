package transport

import (
	"context"
	"time"

	batchqueue "github.com/tkeel-io/core/pkg/util/batch_queue"
)

var (
	readyInsertMessage   = 5000
	intervalConsumerTime = 3 * time.Second
)

type SinkTransport struct {
	sink batchqueue.BatchSink
	fn   batchqueue.ProcessFn
}

func NewSinkTransport(ctx context.Context, name string, fn batchqueue.ProcessFn) (Transport, error) {
	ts := &SinkTransport{fn: fn}

	opts := &batchqueue.Config{
		Name:                  name,
		DoSinkFn:              fn,
		MaxBatching:           readyInsertMessage,
		MaxPendingMessages:    uint(readyInsertMessage),
		BatchingMaxFlushDelay: intervalConsumerTime,
	}
	sink, err := batchqueue.NewBatchSink(ctx, opts)
	ts.sink = sink
	return ts, err
}

func (s *SinkTransport) Send(ctx context.Context, msg interface{}) error {
	return s.sink.Send(ctx, msg)
}

func (s *SinkTransport) Close() {
	s.sink.Close()
}
