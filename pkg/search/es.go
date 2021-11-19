package search

import (
	"context"
	"fmt"
	"github.com/olivere/elastic/v7"
	pb "github.com/tkeel-io/core/api/core/v1"

)

type ESClient struct {
	client *elastic.Client
}

func (es *ESClient) Index(ctx context.Context, req *pb.IndexObject) (out *pb.IndexResponse, err error) {
	fmt.Println(req.Obj.String())
	//	indexID := so["id"].(string)
	es.client.Index().Index("entity").Id("abc").BodyJson(req.Obj.String()).Do(context.Background())
	panic("implement me")
}

func (es *ESClient) Search(context.Context, *pb.SearchRequest) (*pb.SearchResponse, error) {
	panic("implement me")
}

func (es *ESClient) mustEmbedUnimplementedSearchServer() {
	panic("implement me")
}

func NewESClient(url ...string) *ESClient {
	client, err := elastic.NewClient(elastic.SetURL(url...), elastic.SetSniff(false))
	if err != nil {
		panic(err)
	}

	return &ESClient{client: client}
}


