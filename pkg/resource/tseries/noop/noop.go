package noop

import (
	"context"
	"os"

	zfield "github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/core/pkg/resource/tseries"
	"github.com/tkeel-io/core/pkg/util"
	"github.com/tkeel-io/kit/log"
	"go.uber.org/zap"
)

type noop struct {
	id string
}

func (n *noop) Write(ctx context.Context, req *tseries.TSeriesRequest) (*tseries.TSeriesResponse, error) {
	log.Debug("insert time series data, noop.", zap.Any("data", req.Data), zap.Any("metadata", req.Metadata))
	return &tseries.TSeriesResponse{}, nil
}

func init() {
	zfield.SuccessStatusEvent(os.Stdout, "Register Resource<TSDB.noop> successful")
	tseries.Register("noop", func(map[string]interface{}) (tseries.TimeSerier, error) {
		id := util.UUID()
		log.Info("create TSDB.noop instance", zfield.ID(id))
		return &noop{id: id}, nil
	})
}
