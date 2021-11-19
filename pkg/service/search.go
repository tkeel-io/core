package service

import (
	"context"

	pb "github.com/tkeel-io/core/api/core/v1"
)

type SearchService struct {
	pb.UnimplementedSearchServer
	searchClient pb.SearchHTTPServer
}

func NewSearchService(searchClient pb.SearchHTTPServer) *SearchService {
	return &SearchService{searchClient: searchClient}
}

func (s *SearchService) Index(ctx context.Context, req *pb.IndexObject) (*pb.IndexResponse, error) {
	return s.searchClient.Index(ctx, req)
}
func (s *SearchService) Search(ctx context.Context, req *pb.SearchRequest) (*pb.SearchResponse, error) {
	return s.searchClient.Search(ctx, req)
}
