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
	sink    batchqueue.BatchSink
	fn      batchqueue.ProcessFn
	encoder Encoder
}

func NewSinkTransport(ctx context.Context, name string, fn batchqueue.ProcessFn, encoder Encoder) (Transport, error) {
	ts := &SinkTransport{fn: fn, encoder: encoder}

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
	m, err := s.encoder(msg)
	if err != nil {
		return err
	}
	return s.sink.Send(ctx, m)
}

func (s *SinkTransport) Close() {
	s.sink.Close()
}
