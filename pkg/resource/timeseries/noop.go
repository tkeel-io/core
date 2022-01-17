package timeseries

import (
	"context"

	"github.com/tkeel-io/core/pkg/resource"
	"github.com/tkeel-io/kit/log"

	"go.uber.org/zap"
)

var EngineNoop Engine

type noop struct{}

func newNoop() Actuator {
	return &noop{}
}

func (n *noop) Init(meta resource.TimeSeriesMetadata) error {
	log.Info("initialize timeseries.Noop")
	return nil
}

func (n *noop) Write(ctx context.Context, req *WriteRequest) *Response {
	log.Debug("insert time series data, noop.", zap.Any("data", req.Data), zap.Any("metadata", req.Metadata))
	return &Response{}
}

func (n *noop) Query(ctx context.Context, req QueryRequest) *Response {
	return &Response{}
}
