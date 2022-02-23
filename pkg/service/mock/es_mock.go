package mock

import (
	"context"

	pb "github.com/tkeel-io/core/api/core/v1"
)

type SearchMock struct {
}

func NewSearchMock() pb.SearchHTTPServer {
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
		Limit:  200,
		Offset: 0,
	}, nil
}
