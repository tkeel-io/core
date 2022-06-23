package builder

import (
	"context"

	"github.com/tkeel-io/core/pkg/resource"
	"github.com/tkeel-io/core/pkg/resource/rawdata"
	"github.com/tkeel-io/core/pkg/resource/rawdata/clickhouse"
	"github.com/tkeel-io/core/pkg/resource/transport"
)

func init() {
	rawdata.Register("clickhouse", NewRawDataEntry)
}

type RawDataEntry struct {
	clickhouse.Clickhouse
	ts transport.Transport
}

func (r *RawDataEntry) Init(metadata resource.Metadata) error {
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

func (r *RawDataEntry) Write(ctx context.Context, req *rawdata.Request) (err error) {
	return r.ts.Send(ctx, req)
}

func NewRawDataEntry() rawdata.Service {
	return &RawDataEntry{}
}
