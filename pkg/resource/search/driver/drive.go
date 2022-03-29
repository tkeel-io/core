package driver

import (
	"context"
	"strings"

	pb "github.com/tkeel-io/core/api/core/v1"
)

type Type string

type SearchEngine interface {
	BuildIndex(ctx context.Context, index, content string) error
	Search(ctx context.Context, request SearchRequest) (SearchResponse, error)
	Delete(ctx context.Context, id string) error
}

type SelectDriveOption func() Type

func Parse(drive string) SelectDriveOption {
	switch strings.ToLower(drive) {
	case "elasticsearch", "es":
		return Elasticsearch
	default:
		return NoopDriver
	}
}

type SearchRequest struct {
	Source    string                `protobuf:"bytes,1,opt,name=source,proto3" json:"source,omitempty"`
	Owner     string                `protobuf:"bytes,2,opt,name=owner,proto3" json:"owner,omitempty"`
	Query     string                `protobuf:"bytes,3,opt,name=query,proto3" json:"query,omitempty"`
	Page      *pb.Pager             `protobuf:"bytes,4,opt,name=page,proto3" json:"page,omitempty"`
	Condition []*pb.SearchCondition `protobuf:"bytes,5,rep,name=condition,proto3" json:"condition,omitempty"`
}

type SearchResponse struct {
	Total  int64                    `json:"total,omitempty"`
	Data   []map[string]interface{} `json:"data,omitempty"`
	Raw    []byte                   `json:"raw,omitempty"`
	Limit  int32                    `json:"limit"`
	Offset int32                    `json:"offset"`
}

type Generator func(map[string]interface{}) (SearchEngine, error)

// search engine register table.
var registerDrivers = map[Type]Generator{}

func GetDriver(driverType Type) (Generator, bool) {
	driverIns, has := registerDrivers[driverType]
	return driverIns, has
}
