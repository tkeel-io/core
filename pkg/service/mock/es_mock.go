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
	return &pb.DeleteByIDResponse{}, nil
}

func (s *SearchMock) Index(context.Context, *pb.IndexObject) (*pb.IndexResponse, error) {
	return &pb.IndexResponse{Status: ""}, nil
}

func (s *SearchMock) Search(context.Context, *pb.SearchRequest) (*pb.SearchResponse, error) {
	return &pb.SearchResponse{
		PageNum:  2,
		PageSize: 10,
	}, nil
}
