package noop

import (
	"context"

	logf "github.com/tkeel-io/core/pkg/logfield"

	pb "github.com/tkeel-io/core/api/core/v1"
	"github.com/tkeel-io/core/pkg/resource"
	"github.com/tkeel-io/core/pkg/resource/tseries"
	"github.com/tkeel-io/kit/log"
)

type noop struct{}

func newNoop() tseries.TimeSerier {
	return &noop{}
}

func (n *noop) Init(meta resource.Metadata) error {
	log.L().Info("initialize timeseries.Noop")
	return nil
}

func (n *noop) Write(ctx context.Context, req *tseries.TSeriesRequest) (*tseries.TSeriesResponse, error) {
	log.L().Debug("insert time series data, noop.", logf.Any("data", req.Data), logf.Any("metadata", req.Metadata))
	return &tseries.TSeriesResponse{}, nil
}
func (n *noop) Query(ctx context.Context, req *pb.GetTSDataRequest) (*pb.GetTSDataResponse, error) {
	return nil, nil
}

func init() {
	tseries.Register("noop", newNoop)
}
