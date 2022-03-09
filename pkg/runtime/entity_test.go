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
	"testing"

	"github.com/stretchr/testify/assert"
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

	in := []*Result{
		{
			State: en.Raw(),
			Event: nil,
			Patches: []Patch{
				{
					Path:  "properties.temp",
					Value: tdtl.New("50"),
					Op:    OpReplace,
				},
				{
					Path:  "properties.metrics.cpu_used",
					Value: tdtl.New("0.78"),
					Op:    OpReplace,
				},
				{
					Path:  "properties.metrics.mem_used",
					Value: tdtl.New("0.28"),
					Op:    OpReplace,
				},
				{
					Path:  "properties.metrics.interfaces",
					Value: tdtl.New("0.28"),
					Op:    OpAdd,
				},
				{
					Path: "properties.temp",
					Op:   OpRemove,
				},
				{
					Path:  "properties.metrics",
					Value: tdtl.New(`{"temp": 209}`),
					Op:    OpMerge,
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
