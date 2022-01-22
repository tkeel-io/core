package noop

import (
	"context"

	"github.com/tkeel-io/core/pkg/resource/tseries"
	"github.com/tkeel-io/kit/log"
	"go.uber.org/zap"
)

type noop struct{}

func (n *noop) Write(ctx context.Context, req *tseries.TSeriesRequest) (*tseries.TSeriesResponse, error) {
	log.Debug("insert time series data, noop.", zap.Any("data", req.Data), zap.Any("metadata", req.Metadata))
	return &tseries.TSeriesResponse{}, nil
}

func init() {
	tseries.Register("noop", func(properties map[string]interface{}) (tseries.TimeSerier, error) {
		return &noop{}, nil
	})
}
