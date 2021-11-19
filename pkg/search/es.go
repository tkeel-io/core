package search

import (
	"context"
	"reflect"

	pb "github.com/tkeel-io/core/api/core/v1"

	"github.com/olivere/elastic/v7"
	"google.golang.org/protobuf/types/known/structpb"
)

const EntityIndex = "entity"

type ESClient struct {
	client *elastic.Client
}

func interface2string(in interface{}) (out string) {
	if in == nil {
		return
	}
	switch inString := in.(type) {
	case string:
		out = inString
	default:
		out = ""
	}
	return
}
func (es *ESClient) Index(ctx context.Context, req *pb.IndexObject) (out *pb.IndexResponse, err error) {
	var indexID string
	out = &pb.IndexResponse{}
	out.Status = "SUCCESS"

	switch kv := req.Obj.AsInterface().(type) {
	case map[string]interface{}:
		indexID = interface2string(kv["id"])
	default:
		return
	}
	objBytes, _ := req.Obj.MarshalJSON()
	es.client.Index().Index(EntityIndex).Id(indexID).BodyString(string(objBytes)).Do(context.Background())
	return
}

func (es *ESClient) Search(ctx context.Context, req *pb.SearchRequest) (out *pb.SearchResponse, err error) {
	out = &pb.SearchResponse{}
	out.Items = make([]*structpb.Value, 0)
	searchResult, err := es.client.Search().Index(EntityIndex).Query(elastic.NewMultiMatchQuery(req.Data)).Pretty(true).Do(ctx)
	if err != nil {
		return
	}
	var ttyp map[string]interface{}
	for _, item := range searchResult.Each(reflect.TypeOf(ttyp)) {
		if t, ok := item.(map[string]interface{}); ok {
			tt, _ := structpb.NewValue(t)
			out.Items = append(out.Items, tt)
			out.TotalCount++
		}
	}
	return
}

func NewESClient(url ...string) pb.SearchHTTPServer {
	client, err := elastic.NewClient(elastic.SetURL(url...), elastic.SetSniff(false))
	if err != nil {
		panic(err)
	}

	return &ESClient{client: client}
}
