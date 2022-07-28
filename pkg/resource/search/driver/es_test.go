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
	"context"
	"fmt"
	"github.com/olivere/elastic/v7"
	"google.golang.org/protobuf/types/known/structpb"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	pb "github.com/tkeel-io/core/api/core/v1"
)

func printQuery(query elastic.Query) (string, error) {
	src, err := query.Source()
	if err != nil {
		return "", err
	}
	data, err := json.Marshal(src)
	if err != nil {
		return "", err
	}
	got := string(data)
	fmt.Printf("query: %s\n", got)
	return got, nil
}

func printHits(result *elastic.SearchResult) []interface{} {
	total := result.Hits.TotalHits.Value
	num := len(result.Hits.Hits)
	val := make([]interface{}, num, num)
	fmt.Printf("total: %d\n", total)
	for i, hit := range result.Hits.Hits {
		s := hit.Source
		fmt.Printf("result: %d %s\n", i, s)
		val[i] = string(s)
	}
	return val
}

func printProfile(result *elastic.SearchResult) {
	printProfileRet := func(r *elastic.ProfileResult) {
		fmt.Printf("--Description: %s \n--Type: %s \n--NodeTime: %s \n--NodeTimeNanos: %d\n",
			r.Description,
			r.Type,
			r.NodeTime,
			r.NodeTimeNanos)
	}

	for _, val := range result.Profile.Shards {
		fmt.Printf("-->profile: ID: %s\n", val.ID)
		for _, search := range val.Searches {
			fmt.Println("-Searches: ")
			for _, q := range search.Query {
				printProfileRet(&q)
			}
			fmt.Printf("--Collector: %s \n--RewriteTime: %d\n",
				search.Collector,
				search.RewriteTime)
		}
		for _, agg := range val.Aggregations {
			fmt.Println("-Aggregations: ")
			printProfileRet(&agg)
		}
	}
}

func Test_condition2boolQuery(t *testing.T) {
}

func Test_defaultPage(t *testing.T) {
	page := &pb.Pager{Offset: 20}

	defaultPage(page)
	assert.Equal(t, DefaultLimit, page.Limit)
}

func TestESClient_Search(t *testing.T) {
	urlText := "es://admin:admin@tkeel-middleware-elasticsearch-master:9200"
	urlIns, err := url.Parse(urlText)
	if nil != err {
		t.Log(err)
		return
	}

	cfgJSON := make(map[string]interface{})
	cfgJSON["username"] = urlIns.User.Username()
	cfgJSON["password"], _ = urlIns.User.Password()
	cfgJSON["endpoints"] = strings.Split(urlIns.Host, ",")

	se, err := NewElasticsearchEngine(cfgJSON)
	if err != nil {
		t.Log(err)
		return
	}
	req := SearchRequest{}
	req.Page = &pb.Pager{
		Limit:   1,
		Offset:  0,
		Sort:    "owner.keyword",
		Reverse: false,
	}
	resp, err := se.Search(context.Background(), req)

	assert.Nil(t, err)
	t.Log(resp.Total)
	t.Log(len(resp.Data))
	//	t.Log(resp.Data)
}

func TestESClient_Search2(t *testing.T) {
	val, err := structpb.NewValue("template")
	req := SearchRequest{
		Source: "device",
		Owner:  "usr-9dd24b66b6ff21ce9114ea0afbca",
		Query:  "青云",
		Page: &pb.Pager{
			Limit:   20,
			Offset:  0,
			Sort:    "",
			Reverse: false,
		},
		Condition: []*pb.SearchCondition{
			{
				Field:    "type",
				Operator: "$eq",
				Value:    val,
			},
		},
	}
	// http://user1:secret1@localhost:9200
	client, err := elastic.NewClient(elastic.SetURL("http://10.10.98.254:9200"))
	if err != nil {
		t.Fatal(err)
	}

	boolQuery := elastic.NewBoolQuery()

	searchQuery := client.Search().Index(EntityIndex)

	if req.Condition != nil {
		condition2boolQuery(req.Condition, boolQuery)
	}
	if req.Query != "" {
		queryKeyWords := strings.Split(req.Query, " ")

		if len(queryKeyWords) == 1 {
			boolQuery.Should(elastic.NewMultiMatchQuery(req.Query).Boost(1000))
			boolQuery.Should(elastic.NewWildcardQuery("keywords", fmt.Sprintf("*%s*",
				queryKeyWords[0])).
				Boost(10))
		} else {
			for _, val := range queryKeyWords {
				boolQuery.Must(elastic.NewWildcardQuery("keywords", fmt.Sprintf("*%s*", val)))
				//boolQuery.Should(elastic.NewRegexpQuery(name, fmt.Sprintf("*%s*", val)))
			}
		}
	}
	if _, err = printQuery(boolQuery); err != nil {
		t.Fatal(err)
	} else {
		req.Page = defaultPage(req.Page)
		searchQuery = searchQuery.Sort(req.Page.Sort, !req.Page.Reverse)
		searchQuery = searchQuery.Query(boolQuery).From(int(req.Page.Offset)).Size(int(req.Page.Limit)).Profile(true)

		searchResult, err := searchQuery.Pretty(true).Do(context.Background())
		if err != nil {
			t.Error(err)
		} else {
			printHits(searchResult)
			printProfile(searchResult)
		}
	}
}
