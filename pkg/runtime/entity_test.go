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
	v1 "github.com/tkeel-io/core/api/core/v1"
	xjson "github.com/tkeel-io/core/pkg/util/json"
	"github.com/tkeel-io/tdtl"
)

func TestNewEntity(t *testing.T) {
	en, err := NewEntity("en-123", []byte(`{"properties": {"temp": 20}}`))
	assert.Nil(t, err)
	t.Log(string(en.Raw()))
}

func TestEntity_HandleScheme(t *testing.T) {
	en, err := NewEntity("en-123", []byte(`{"properties": {"temp": 20}}`))
	assert.Nil(t, err)

	feeds := []*Feed{
		{
			Event: &v1.ProtoEvent{
				Metadata: map[string]string{},
			},
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
		{
			Event: &v1.ProtoEvent{
				Metadata: map[string]string{
					v1.MetaPathConstructor: string(v1.PCScheme),
				},
			},
			Patches: []Patch{
				{
					Path:  "scheme.attributes.define.fields.serial-N",
					Op:    xjson.OpReplace,
					Value: tdtl.New(`{"id":"serial-N","type":"string","name":"序列号N","weight":0,"enabled":false,"enabled_search":false,"enabled_time_series":false,"description":"设备批次","define":{"default_value":"xxxxxxxn","rw":"w"},"last_time":1652164204383}`),
				},
				{
					Path:  "scheme.attributes.define.fields.serial-3",
					Op:    xjson.OpReplace,
					Value: tdtl.New(`{"id":"serial-3","type":"string","name":"序列号3","weight":0,"enabled":false,"enabled_search":false,"enabled_time_series":false,"description":"设备批次","define":{"default_value":"xxxxxxx3","rw":"r"},"last_time":1652164204383}`),
				},
			},
		},
	}

	tests := []struct {
		name string
		feed *Feed
		want map[string]string
	}{
		{
			"0",
			feeds[0],
			map[string]string{
				"properties": `{"metrics":{"cpu_used":0.78,"mem_used":0.28,"interfaces":[0.28],"temp":209}}`,
			},
		},
		{
			"1",
			feeds[1],
			map[string]string{
				"scheme.attributes.define.fields.serial-N": `{"id":"serial-N","type":"string","name":"序列号N","weight":0,"enabled":false,"enabled_search":false,"enabled_time_series":false,"description":"设备批次","define":{"default_value":"xxxxxxxn","rw":"w"},"last_time":1652164204383}`,
				"scheme.attributes.define.fields.serial-3": `{"id":"serial-3","type":"string","name":"序列号3","weight":0,"enabled":false,"enabled_search":false,"enabled_time_series":false,"description":"设备批次","define":{"default_value":"xxxxxxx3","rw":"r"},"last_time":1652164204383}`,
			},
		},
	}

	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := en.Handle(ctx, tt.feed)
			for path, val := range tt.want {
				cc := tdtl.New(got.State)
				if !assert.JSONEq(t, val, cc.Get(path).String()) {
					t.Errorf("Path= %v, \nHandle() = %v, \nwant %v", path, cc.Get(path).String(), val)
				}
			}
		})
	}
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

func Test_makeSubPath(t *testing.T) {
	tests := []struct {
		name    string
		dest    string
		src     string
		path    string
		want    string
		want1   string
		wantErr bool
	}{
		{"1", "{}", "{}", "a.b", "{}", "a.b", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := makeSubPath([]byte(tt.dest), []byte(tt.src), tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("makeSubPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.want != string(got) {
				t.Errorf("makeSubPath() got = %v, want %v", string(got), tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("makeSubPath() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
