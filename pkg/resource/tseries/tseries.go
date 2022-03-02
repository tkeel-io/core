package tseries

import (
	"context"
	"time"

	pb "github.com/tkeel-io/core/api/core/v1"
	"github.com/tkeel-io/core/pkg/resource"
)

var registeredTS = make(map[string]TSGenerator)

type TSeriesData struct { //nolint
	Measurement string
	Tags        map[string]string
	Fields      map[string]string
	Value       string
	Timestamp   int64
}

type TSeriesRequest struct { //nolint
	Data     interface{}       `json:"data"`
	Metadata map[string]string `json:"metadata"`
}

type TSeriesResponse struct { //nolint
	Data     []byte            `json:"data"`
	Metadata map[string]string `json:"metadata"`
}

type TSeriesQueryRequest struct { //nolint
	Id          string `json:"id"`
	StartTime   int64  `protobuf:"varint,2,opt,name=start_time,json=startTime,proto3" json:"start_time,omitempty"`
	EndTime     int64  `protobuf:"varint,3,opt,name=end_time,json=endTime,proto3" json:"end_time,omitempty"`
	Identifiers string `protobuf:"bytes,4,opt,name=identifiers,proto3" json:"identifiers,omitempty"`
}

type TSData struct {
	Time  time.Time              `json:"time"`
	Value map[string]interface{} `json:"value"`
}

type TSeriesQueryResponse struct { //nolint
	Items []*pb.TsResponse `json:"items"`
	Id    string           `json:"id"`
	Total int32            `json:"total"`
}

type TimeSerier interface {
	Init(resource.Metadata) error
	Write(ctx context.Context, req *TSeriesRequest) (*TSeriesResponse, error)
	Query(ctx context.Context, req *pb.GetTsDataRequest) (*pb.GetTsDataResponse, error)
}

type TSGenerator func() TimeSerier

func NewTimeSerier(name string) TimeSerier {
	if generator, has := registeredTS[name]; has {
		return generator()
	}
	return registeredTS["noop"]()
}

func Register(name string, handler TSGenerator) {
	registeredTS[name] = handler
}
