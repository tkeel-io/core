package service

import (
	"context"

	"github.com/pkg/errors"
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
	out, err := s.searchClient.Index(ctx, req)
	if err != nil {
		return out, errors.Wrap(err, "index failed")
	}
	return out, nil
}
func (s *SearchService) Search(ctx context.Context, req *pb.SearchRequest) (*pb.SearchResponse, error) {
	out, err := s.searchClient.Search(ctx, req)
	if err != nil {
		return out, errors.Wrap(err, "search failed")
	}
	return out, nil
}
