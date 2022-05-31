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
	"net/url"
	"reflect"

	"github.com/goinggo/mapstructure"
	pb "github.com/tkeel-io/core/api/core/v1"
	xerrors "github.com/tkeel-io/core/pkg/errors"
	logf "github.com/tkeel-io/core/pkg/logfield"

	elastic "github.com/olivere/elastic/v7"
	"github.com/pkg/errors"
	"github.com/tkeel-io/kit/log"
)

const DriverTypeElasticsearch Type = "elasticsearch"

const (
	EntityIndex        = "entity"
	DefaultLimit int32 = 20
	MaxLimit     int32 = 500
)

type ESConfig struct {
	Username  string   `json:"username" mapstructure:"username"`
	Password  string   `json:"password" mapstructure:"password"`
	Endpoints []string `json:"endpoints" mapstructure:"endpoints"`
}

type ESClient struct {
	Client *elastic.Client
}

func NewElasticsearchEngine(cfgJSON map[string]interface{}) (SearchEngine, error) {
	var cfg ESConfig
	if err := mapstructure.Decode(cfgJSON, &cfg); nil != err {
		log.L().Error("decode elasticsearch configuration", logf.Error(err), logf.Value(cfgJSON))
		return nil, errors.Wrap(err, "decode elasticsearch configuration")
	}

	addHTTPScheme(cfg.Endpoints)
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint
	client, err := elastic.NewClient(
		elastic.SetURL(cfg.Endpoints...),
		elastic.SetSniff(false),
		elastic.SetBasicAuth(cfg.Username, cfg.Password),
	)
	if err != nil {
		return nil, err
	}

	// ping connection.
	if len(cfg.Endpoints) == 0 {
		log.L().Error("please check your configuration with elasticsearch")
		return nil, errors.Wrap(xerrors.ErrEmptyParam, "elasticsearch broker endpoints empty")
	}

	info, _, err := client.Ping(cfg.Endpoints[0]).Do(context.Background())
	if nil != err {
		log.L().Error("ping elasticsearch cluster", logf.Error(err))
		return nil, errors.Wrap(err, "ping elasticsearch cluster")
	}

	log.L().Info("use ElasticsearchDriver version:", logf.Value(info.Version.Number))
	client.Index().Index(EntityIndex).Id("core_init").BodyString(`{"id":"core_init"}`).Do(context.Background())
	// 1. 检查索引是否存在 2. 字段是否符合要求 3. 创建索引
	client.PutMapping().Index(EntityIndex).BodyString(`{"properties":{"sysField":{"properties":{"_subscribeAddr":{"type":"text", "fields":{"keyword":{"type":"keyword", "ignore_above":4096}}}}}}}`).Do(context.Background())
	return &ESClient{Client: client}, nil
}

func (es *ESClient) BuildIndex(ctx context.Context, index, body string) error {
	if _, err := es.Client.Index().Index(EntityIndex).
		Id(index).BodyString(body).Refresh("true").Do(ctx); err != nil {
		return errors.Wrap(err, "set index in es error")
	}
	return nil
}

func (es *ESClient) Delete(ctx context.Context, id string) error {
	_, err := es.Client.Delete().Index(EntityIndex).Id(id).Refresh("true").Do(ctx)
	if nil != err {
		if elastic.IsNotFound(err) {
			return errors.Wrap(xerrors.ErrEntityNotFound, "elasticsearch delete by id")
		}
	}
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
	searchQuery = searchQuery.Sort(req.Page.Sort, !req.Page.Reverse)
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
			boolQuery = boolQuery.Must(elastic.NewTermQuery(condition.Field+".keyword", condition.Value.AsInterface()))
		case "$eq":
			switch valueItem := condition.Value.AsInterface().(type) {
			case bool:
				boolQuery = boolQuery.Must(elastic.NewTermQuery(condition.Field, valueItem))
			default:
				boolQuery = boolQuery.Must(elastic.NewTermQuery(condition.Field+".keyword", valueItem))
			}
		case "$prefix":
			boolQuery = boolQuery.Must(elastic.NewPrefixQuery(condition.Field+".keyword", condition.Value.GetStringValue()))
		case "$wildcard":
			wildcard := condition.Value.GetStringValue()
			if wildcard != "" {
				boolQuery = boolQuery.Must(elastic.NewWildcardQuery(condition.Field+".keyword", "*"+wildcard+"*"))
			}
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
		page.Limit = DefaultLimit
	} else if page.Limit > MaxLimit {
		page.Limit = MaxLimit
	}

	if page.Sort == "" {
		page.Sort = "id.keyword"
	}
	return page
}

func Elasticsearch() Type {
	return DriverTypeElasticsearch
}

func init() {
	registerDrivers[DriverTypeElasticsearch] = NewElasticsearchEngine
}

func addHTTPScheme(endpoints []string) []string {
	for index := range endpoints {
		urlIns := url.URL{}
		urlIns.Scheme = "http"
		urlIns.Host = endpoints[index]

		// set endpoint.
		endpoints[index] = urlIns.String()
	}
	return endpoints
}
