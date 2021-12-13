/*
Copyright 2021 The tKeel Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package search

import (
	"context"
	"crypto/tls"
	"net/http"
	"os"
	"reflect"

	"github.com/pkg/errors"
	pb "github.com/tkeel-io/core/api/core/v1"
	"github.com/tkeel-io/core/pkg/print"

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
	case nil:
		// do nothing.
		return out, nil
	default:
		return out, ErrIndexParamInvalid
	}

	objBytes, _ := req.Obj.MarshalJSON()
	_, err = es.client.Index().Index(EntityIndex).Id(indexID).BodyString(string(objBytes)).Do(context.Background())
	return out, errors.Wrap(err, "es index failed")
}

// reference: https://www.tutorialspoint.com/elasticsearch/elasticsearch_query_dsl.htm#:~:text=In%20Elasticsearch%2C%20searching%20is%20carried%20out%20by%20using,look%20for%20a%20specific%20value%20in%20specific%20field.
// convert condition.
func condition2boolQuery(conditions []*pb.SearchCondition, boolQuery *elastic.BoolQuery) {
	for _, condition := range conditions {
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
	boolQuery := elastic.NewBoolQuery()
	searchQuery := es.client.Search().Index(EntityIndex)

	if req.Condition != nil {
		condition2boolQuery(req.Condition, boolQuery)
	}
	if req.Query != "" {
		boolQuery = boolQuery.Must(elastic.NewMultiMatchQuery(req.Query))
	}

	req.Page = defaultPage(req.Page)
	searchQuery = searchQuery.Query(boolQuery)
	// searchQuery = searchQuery.Sort(req.Page.Sort, req.Page.Reverse).
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

func defaultPage(page *pb.Pager) *pb.Pager {
	if nil == page {
		page = &pb.Pager{}
	}

	if page.Limit == 0 {
		page.Limit = 10
	}
	if page.Sort == "" {
		page.Sort = "id"
	}
	return page
}

func NewESClient(url ...string) pb.SearchHTTPServer {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint
	client, err := elastic.NewClient(elastic.SetURL(url...), elastic.SetSniff(false), elastic.SetBasicAuth("admin", "admin"))
	if err != nil {
		panic(err)
	}

	// ping connection.
	info, _, err := client.Ping(url[0]).Do(context.Background())
	if nil != err {
		panic(err)
	}

	print.InfoStatusEvent(os.Stdout, "use Elasticsearch version<%s>", info.Version.Number)
	return &ESClient{client: client}
}
