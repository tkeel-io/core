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

package driver

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"net/http"
	"reflect"

	pb "github.com/tkeel-io/core/api/core/v1"
	"github.com/tkeel-io/core/pkg/config"

	"github.com/olivere/elastic/v7"
	"github.com/pkg/errors"
	"github.com/tkeel-io/kit/log"
)

const ElasticsearchDriver Type = "elasticsearch"

const EntityIndex = "entity"

type ESClient struct {
	Client *elastic.Client
}

func NewElasticsearchEngine(config config.ESConfig) SearchEngine {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec
	client, err := elastic.NewClient(
		elastic.SetURL(config.Address...),
		elastic.SetSniff(false),
		elastic.SetBasicAuth(config.Username, config.Password),
	)
	if err != nil {
		log.Fatal(err)
	}

	// ping connection.
	if len(config.Address) == 0 {
		log.Fatal("please check your configuration with elasticsearch")
	}
	info, _, err := client.Ping(config.Address[0]).Do(context.Background())
	if nil != err {
		log.Fatal(err)
	}
	log.Info("use ElasticsearchDriver version:", info.Version.Number)
	return &ESClient{Client: client}
}

func (es *ESClient) BuildIndex(ctx context.Context, id, body string) error {
	if _, err := es.Client.Index().Index(EntityIndex).Id(id).BodyString(body).Do(ctx); err != nil {
		return errors.Wrap(err, "set index in es error")
	}
	return nil
}

func (es *ESClient) Delete(ctx context.Context, id string) error {
	_, err := es.Client.Delete().Index(EntityIndex).Id(id).Do(ctx)
	return errors.Wrap(err, "elasticsearch delete by id")
}

func (es *ESClient) DeleteByQuery(ctx context.Context, query map[string]interface{}) error {
	var bytes bytes.Buffer
	if err := json.NewEncoder(&bytes).Encode(query); err != nil {
		return errors.Wrap(err, "json encoding query")
	} else if _, err = es.Client.DeleteByQuery(EntityIndex, bytes.String()).DoAsync(ctx); err != nil {
		return errors.Wrap(err, "elasticsearch deleye by query")
	}

	return nil
}

func (es *ESClient) Search(ctx context.Context, req SearchRequest) (SearchResponse, error) {
	resp := SearchResponse{}
	boolQuery := elastic.NewBoolQuery()
	searchQuery := es.Client.Search().Index(EntityIndex)

	if req.Condition != nil {
		condition2boolQuery(req.Condition, boolQuery)
	}
	if req.Query != "" {
		boolQuery = boolQuery.Must(elastic.NewMultiMatchQuery(req.Query))
	}

	req.Page = defaultPage(req.Page)
	// searchQuery = searchQuery.Sort(req.Page.Sort, req.Page.Reverse).
	searchQuery = searchQuery.Query(boolQuery).From(int(req.Page.Offset)).Size(int(req.Page.Limit))

	searchResult, err := searchQuery.Pretty(true).Do(ctx)
	if err != nil {
		return resp, errors.Wrap(err, "query search failed")
	}

	var data []map[string]interface{}
	for _, item := range searchResult.Each(reflect.TypeOf(map[string]interface{}{})) {
		if t, ok := item.(map[string]interface{}); ok {
			data = append(data, t)
		}
	}

	resp.Total = searchResult.TotalHits()
	resp.Data = data
	resp.Raw, _ = json.Marshal(data)
	if req.Page != nil {
		resp.Limit = req.Page.Limit
		resp.Offset = req.Page.Offset
	}

	return resp, nil
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

func Elasticsearch() Type {
	return ElasticsearchDriver
}
