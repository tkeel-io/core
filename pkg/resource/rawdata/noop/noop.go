package noop

import (
	"context"

	pb "github.com/tkeel-io/core/api/core/v1"
	"github.com/tkeel-io/core/pkg/resource"
	"github.com/tkeel-io/core/pkg/resource/rawdata"
)

type Noop struct {
}

func (n *Noop) Init(resource.Metadata) error {
	return nil
}

func (n *Noop) Write(ctx context.Context, req *rawdata.RawDataRequest) error {
	return nil
}

func (n *Noop) Query(ctx context.Context, req *pb.GetRawdataRequest) (*pb.GetRawdataResponse, error) {
	return nil, nil
}

func NewNoop() rawdata.RawDataService {
	return &Noop{}
}

func init() {
	rawdata.Register("noop", NewNoop)
}
