package search

import (
	"context"

	pb "github.com/tkeel-io/core/api/core/v1"
	"github.com/tkeel-io/core/pkg/search/driver"
)

type searchEngine interface {
}

type service struct {
	driver map[driver.Type]searchEngine
}

func (s *service) Search(ctx context.Context, request *pb.SearchRequest) (*pb.SearchResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (s *service) DeleteByID(ctx context.Context, request *pb.DeleteByIDRequest) (*pb.DeleteByIDResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (s *service) Index(ctx context.Context, in *pb.IndexObject) (*pb.IndexResponse, error) {
	//TODO implement me
	panic("implement me")
}
