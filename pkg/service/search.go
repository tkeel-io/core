package service

import (
	"context"

	pb "github.com/tkeel-io/core/api/core/v1"
)

type SearchService struct {
	pb.UnimplementedSearchServer
}

func NewSearchService() *SearchService {
	return &SearchService{}
}

func (s *SearchService) Index(ctx context.Context, req *pb.IndexObject) (*pb.IndexResponse, error) {
	return &pb.IndexResponse{}, nil
}
func (s *SearchService) Search(ctx context.Context, req *pb.SearchRequest) (*pb.SearchResponse, error) {
	return &pb.SearchResponse{}, nil
}
