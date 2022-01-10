package mock

import (
	"context"

	pb "github.com/tkeel-io/core/api/core/v1"
)

type SearchMock struct {
}

func NewSearchMock() *SearchMock {
	return &SearchMock{}
}

func (s *SearchMock) DeleteByID(context.Context, *pb.DeleteByIDRequest) (*pb.DeleteByIDResponse, error) {
	return nil, nil
}

func (s *SearchMock) Index(context.Context, *pb.IndexObject) (*pb.IndexResponse, error) {
	return nil, nil
}

func (s *SearchMock) Search(context.Context, *pb.SearchRequest) (*pb.SearchResponse, error) {
	return nil, nil
}
