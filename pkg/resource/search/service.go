package search

import (
	"context"
	"github.com/tkeel-io/core/pkg/config"

	pb "github.com/tkeel-io/core/api/core/v1"
	"github.com/tkeel-io/core/pkg/resource/search/driver"
)

var Service = newService()

type service struct {
	drivers map[driver.Type]driver.Engine
}

func newService() *service {

	return &service{drivers: map[driver.Type]driver.Engine{
		driver.Elasticsearch: driver.NewElasticsearchEngine(config.Get().Elasticsearch.Url),
	}}
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
