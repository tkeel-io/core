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
	"encoding/json"
	"fmt"
	"testing"

	structpb "google.golang.org/protobuf/types/known/structpb"

	"github.com/stretchr/testify/assert"
	pb "github.com/tkeel-io/core/api/core/v1"
	"github.com/tkeel-io/core/pkg/config"
)

func Test_condition2boolQuery(t *testing.T) {

}

func Test_defaultPage(t *testing.T) {
	page := &pb.Pager{Offset: 20}

	defaultPage(page)
	assert.Equal(t, int64(10), page.Limit)
}

func TestESClient_Search(t *testing.T) {

	json1 := `{"a":{"b":{"c":4}}, "c":1}`
	json2 := `{"a":{"b":{"d":4}}}`

	tt := make(map[string]interface{})

	json.Unmarshal([]byte(json1), &tt)
	json.Unmarshal([]byte(json2), &tt)
	fmt.Println(tt)

	config := config.ESConfig{
		Endpoints: []string{"http://192.168.123.9:31770"},
		Username:  "admin",
		Password:  "admin",
	}
	es := NewElasticsearchEngine(config)
	value := structpb.NewBoolValue(false)
	req := SearchRequest{
		Source: "source",
		Owner:  "usr-a683a762f176f6b6a6c6fee42546",
		Query:  "",
		Page: &pb.Pager{
			Limit:   10,
			Offset:  0,
			Reverse: false,
		},

		Condition: []*pb.SearchCondition{&pb.SearchCondition{

			Value: value,
		}},
	}
	res, err := es.Search(context.Background(), req)
	if err == nil {
		t.Log(res)
		t.Log(res.Data)
	}

}
