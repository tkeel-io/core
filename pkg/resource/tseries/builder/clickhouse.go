package builder

import (
	"context"
	"github.com/tkeel-io/core/pkg/resource"
	"github.com/tkeel-io/core/pkg/resource/transport"
	"github.com/tkeel-io/core/pkg/resource/tseries"
	"github.com/tkeel-io/core/pkg/resource/tseries/clickhouse"
)

func init() {
	tseries.Register("clickhouse", NewTimeSeriesEntry)
}

type TimeSeriesEntry struct {
	clickhouse.Clickhouse
	ts transport.Transport
}

func (r *TimeSeriesEntry) Init(metadata resource.Metadata) error {
	ck := &clickhouse.Clickhouse{}
	if err := ck.Init(metadata); err != nil {
		return err
	}
	ts, err := transport.NewClickHouseTransport(context.Background(), ck)
	if err != nil {
		panic(err)
	}
	r.Clickhouse = *ck
	r.ts = ts
	return err
}

func (r *TimeSeriesEntry) Write(ctx context.Context, req *tseries.TSeriesRequest) (*tseries.TSeriesResponse, error) {
	return &tseries.TSeriesResponse{}, r.ts.Send(ctx, req)
}

func NewTimeSeriesEntry() tseries.TimeSerier {
	return &TimeSeriesEntry{}
}
