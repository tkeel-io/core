package search

import (
	"context"
	"crypto/tls"
	"net/http"
	"reflect"

	"github.com/pkg/errors"
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

func condition2boolQuery(condition []*pb.SearchCondition, boolQuery *elastic.BoolQuery) {
	for _, condition := range condition {
		switch condition.Operator {
		case "$lt":
			boolQuery = boolQuery.Must(elastic.NewRangeQuery(condition.Field).Lt(condition.Value.AsInterface()))
		case "$lte":
			boolQuery = boolQuery.Must(elastic.NewRangeQuery(condition.Field).Lte(condition.Value.AsInterface()))
		case "$gt":
			boolQuery = boolQuery.Must(elastic.NewRangeQuery(condition.Field).Gt(condition.Value.AsInterface()))
		case "$gte":
			boolQuery = boolQuery.Must(elastic.NewRangeQuery(condition.Field).Gte(condition.Value.AsInterface()))
		case "$neq":
			boolQuery = boolQuery.MustNot(elastic.NewTermQuery(condition.Field, condition.Value.AsInterface()))
		case "$eq":
			boolQuery = boolQuery.Must(elastic.NewTermQuery(condition.Field, condition.Value.AsInterface()))
		default:
			boolQuery = boolQuery.Must(elastic.NewMatchQuery(condition.Field, condition.Value.AsInterface()))
		}
	}
}

func (es *ESClient) Search(ctx context.Context, req *pb.SearchRequest) (out *pb.SearchResponse, err error) {
	out = &pb.SearchResponse{}
	out.Items = make([]*structpb.Value, 0)
	searchQuery := es.client.Search().Index(EntityIndex)

	boolQuery := elastic.NewBoolQuery()
	if req.Condition != nil {
		condition2boolQuery(req.Condition, boolQuery)
	}
	if req.Query != "" {
		boolQuery = boolQuery.Must(elastic.NewMultiMatchQuery(req.Query))
	}

	if req.Page == nil {
		req.Page = &pb.Pager{
			Limit:   10,
			Offset:  0,
			Sort:    "",
			Reverse: false,
		}
	}
	searchQuery = searchQuery.Query(boolQuery)
	if req.Page.Sort != "" {
		searchQuery = searchQuery.Sort(req.Page.Sort, req.Page.Reverse)
	}

	searchQuery = searchQuery.From(int(req.Page.Offset)).Size(int(req.Page.Limit))

	searchResult, err := searchQuery.Pretty(true).Do(ctx)
	if err != nil {
		return
	}
	var ttyp map[string]interface{}
	for _, item := range searchResult.Each(reflect.TypeOf(ttyp)) {
		if t, ok := item.(map[string]interface{}); ok {
			tt, _ := structpb.NewValue(t)
			out.Items = append(out.Items, tt)
		}
	}
	out.Total = searchResult.TotalHits()
	if req.Page != nil {
		out.Limit = req.Page.Limit
		out.Offset = req.Page.Offset
	}
	return out, errors.Wrap(err, "search failed")
}

func NewESClient(url ...string) pb.SearchHTTPServer {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint
	client, err := elastic.NewClient(elastic.SetURL(url...), elastic.SetSniff(false), elastic.SetBasicAuth("admin", "admin"))
	if err != nil {
		panic(err)
	}

	return &ESClient{client: client}
}
