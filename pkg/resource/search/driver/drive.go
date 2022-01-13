package driver

import (
	"context"

	pb "github.com/tkeel-io/core/api/core/v1"
)

type Type string

type Engine interface {
	BuildIndex(ctx context.Context, index, content string) error
	Search(ctx context.Context, request SearchRequest) (SearchResponse, error)
	Delete(ctx context.Context, id string) error
}

type SelectDriveOption func() Type

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
	Limit  int64                    `json:"limit"`
	Offset int64                    `json:"offset"`
}
