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

package runtime

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "github.com/tkeel-io/core/api/core/v1"
	xjson "github.com/tkeel-io/core/pkg/util/json"
	"github.com/tkeel-io/tdtl"
)

func TestNewEntity(t *testing.T) {
	en, err := NewEntity("en-123", []byte(`{"properties": {"temp": 20}}`))
	assert.Nil(t, err)
	t.Log(string(en.Raw()))
}

func TestEntity_Handle(t *testing.T) {
	en, err := NewEntity("en-123", []byte(`{"properties": {"temp": 20}}`))
	assert.Nil(t, err)

	in := []*Feed{
		{
			Event: &v1.ProtoEvent{
				Metadata: map[string]string{}},
			Patches: []Patch{
				{
					Path:  "properties.temp",
					Value: tdtl.New("50"),
					Op:    xjson.OpReplace,
				},
				{
					Path:  "properties.metrics.cpu_used",
					Value: tdtl.New("0.78"),
					Op:    xjson.OpReplace,
				},
				{
					Path:  "properties.metrics.mem_used",
					Value: tdtl.New("0.28"),
					Op:    xjson.OpReplace,
				},
				{
					Path:  "properties.metrics.interfaces",
					Value: tdtl.New("0.28"),
					Op:    xjson.OpAdd,
				},
				{
					Path: "properties.temp",
					Op:   xjson.OpRemove,
				},
				{
					Path:  "properties.metrics",
					Value: tdtl.New(`{"temp": 209}`),
					Op:    xjson.OpMerge,
				},
			},
		},
	}

	for _, test := range in {
		t.Run("test", func(t *testing.T) {
			result := en.Handle(context.TODO(), test)
			assert.Nil(t, result.Err)
			t.Log("result", result)
		})
	}

	t.Log(string(en.Raw()))
}

func TestMerge(t *testing.T) {
	cc := tdtl.New("{}")
	cc.Merge(tdtl.New([]byte(`{"sss":{"id":"sss","type":"struct","name":"","weight":0,"enabled":true,"enabled_search":true,"enabled_time_series":false,"description":"","define":{"fields":{"aaa":{"id":"aaa","type":"struct","name":"","weight":0,"enabled":true,"enabled_search":true,"enabled_time_series":false,"description":"","define":{"fields":{}},"last_time":0}}},"last_time":0}}`)))
	t.Log(cc.String())
}

func TestAdd(t *testing.T) {
	cc := tdtl.New(`{"id":"device123","type":"DEVICE","owner":"admin","source":"CORE","version":0,"last_time":0,"mappers":null,"template_id":"","properties":{"sysField":{"_spacePath":"tomas"},"temp":300},"scheme":{}}`)
	cc.Append("properties.mems", tdtl.New(`0.4`))

	var v interface{}
	json.Unmarshal(cc.Raw(), &v)
}
