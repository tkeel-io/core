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

package json

import (
	"testing"

	jsoniter "github.com/json-iterator/go"

	"github.com/stretchr/testify/assert"
	"github.com/tkeel-io/collectjs"
	"github.com/tkeel-io/collectjs/pkg/json/jsonparser"
	"github.com/tkeel-io/tdtl"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func TestEncodeJSON(t *testing.T) {
	kvalues := map[string]tdtl.Node{
		"id":         tdtl.StringNode("iotd-device123"),
		"name":       tdtl.StringNode("device123"),
		"type":       tdtl.StringNode("DEVICE"),
		"temp":       tdtl.IntNode(25),
		"cpu_used":   tdtl.FloatNode(0.25),
		"interfaces": tdtl.New(`["eth0"]`),
		"conns": tdtl.New(`{
			"hello":     "world",
			"helloss":   "world",
			"hellosss":  "world",
			"hellossss": "world",
			"hellosx":   "world",
			"test": {
				"conns":      {"hello": "world"},
				"id":         "iotd-device123",
				"name":       "device123",
				"type":       "DEVICE",
				"temp":       25,
				"cpu_used":   0.25,
				"interfaces": ["eth888", "xxxx"],
			},
		}`),
	}

	bytes, err := EncodeJSON(kvalues)
	assert.Nil(t, err)
	t.Log(string(bytes))
}

func BenchmarkMarshalJson(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		json.Marshal(map[string]interface{}{
			"id":         "iotd-device123",
			"name":       "device123",
			"type":       "DEVICE",
			"temp":       25,
			"cpu_used":   0.25,
			"interfaces": []string{"eth0"},
			"conns": map[string]interface{}{
				"hello":     "world",
				"helloss":   "world",
				"hellosss":  "world",
				"hellossss": "world",
				"hellosx":   "world",
				"test": map[string]interface{}{
					"conns": map[string]interface{}{
						"hello": "world",
					},
					"id":         "iotd-device123",
					"name":       "device123",
					"type":       "DEVICE",
					"temp":       25,
					"cpu_used":   0.25,
					"interfaces": []string{"eth888", "xxxx"},
				},
			},
			"conns1": map[string]interface{}{
				"hello":     "world",
				"helloss":   "world",
				"hellosss":  "world",
				"hellossss": "world",
				"hellosx":   "world",
				"test": map[string]interface{}{
					"conns": map[string]interface{}{
						"hello": "world",
					},
					"id":         "iotd-device123",
					"name":       "device123",
					"type":       "DEVICE",
					"temp":       25,
					"cpu_used":   0.25,
					"interfaces": []string{"eth888", "xxxx"},
				},
			},
			"conns2": map[string]interface{}{
				"hello":     "world",
				"helloss":   "world",
				"hellosss":  "world",
				"hellossss": "world",
				"hellosx":   "world",
				"test": map[string]interface{}{
					"conns": map[string]interface{}{
						"hello": "world",
					},
					"id":         "iotd-device123",
					"name":       "device123",
					"type":       "DEVICE",
					"temp":       25,
					"cpu_used":   0.25,
					"interfaces": []string{"eth888", "xxxx"},
				},
			},
			"conns3": map[string]interface{}{
				"hello":     "world",
				"helloss":   "world",
				"hellosss":  "world",
				"hellossss": "world",
				"hellosx":   "world",
				"test": map[string]interface{}{
					"conns": map[string]interface{}{
						"hello": "world",
					},
					"id":         "iotd-device123",
					"name":       "device123",
					"type":       "DEVICE",
					"temp":       25,
					"cpu_used":   0.25,
					"interfaces": []string{"eth888", "xxxx"},
				},
			},
		})
	}
}

func BenchmarkCollectJson(b *testing.B) {
	kvalues := map[string]tdtl.Node{
		"id":         tdtl.StringNode("iotd-device123"),
		"name":       tdtl.StringNode("device123"),
		"type":       tdtl.StringNode("DEVICE"),
		"temp":       tdtl.IntNode(25),
		"cpu_used":   tdtl.FloatNode(0.25),
		"interfaces": tdtl.New(`["eth0"]`),
		"conns": tdtl.New(`{
			"hello":     "world",
			"helloss":   "world",
			"hellosss":  "world",
			"hellossss": "world",
			"hellosx":   "world",
			"test": {
				"conns":      {"hello": "world"},
				"id":         "iotd-device123",
				"name":       "device123",
				"type":       "DEVICE",
				"temp":       25,
				"cpu_used":   0.25,
				"interfaces": ["eth888", "xxxx"],
			},
		}`),
		"conns1": tdtl.New(`{
			"hello":     "world",
			"helloss":   "world",
			"hellosss":  "world",
			"hellossss": "world",
			"hellosx":   "world",
			"test": {
				"conns":      {"hello": "world"},
				"id":         "iotd-device123",
				"name":       "device123",
				"type":       "DEVICE",
				"temp":       25,
				"cpu_used":   0.25,
				"interfaces": ["eth888", "xxxx"],
			},
		}`),
		"conns2": tdtl.New(`{
			"hello":     "world",
			"helloss":   "world",
			"hellosss":  "world",
			"hellossss": "world",
			"hellosx":   "world",
			"test": {
				"conns":      {"hello": "world"},
				"id":         "iotd-device123",
				"name":       "device123",
				"type":       "DEVICE",
				"temp":       25,
				"cpu_used":   0.25,
				"interfaces": ["eth888", "xxxx"],
			},
		}`),
		"conns3": tdtl.New(`{
			"hello":     "world",
			"helloss":   "world",
			"hellosss":  "world",
			"hellossss": "world",
			"hellosx":   "world",
			"test": {
				"conns":      {"hello": "world"},
				"id":         "iotd-device123",
				"name":       "device123",
				"type":       "DEVICE",
				"temp":       25,
				"cpu_used":   0.25,
				"interfaces": ["eth888", "xxxx"],
			},
		}`),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		EncodeJSON(kvalues)
	}
}

func TestObject(t *testing.T) {
	collect := collectjs.ByteNew([]byte(`{}`))
	t.Log(collect.GetDataType() == jsonparser.Object.String())
}

func TestJsonParser(t *testing.T) {
	bytes, err := jsonparser.Set([]byte(`{}}`), []byte(`"tom"`), "__name", "_def")
	t.Log(err)
	t.Log(string(bytes))
}

func TestParserType(t *testing.T) {
	tt, err := jsonparser.ParseType([]byte(`"ssssssss"`))
	t.Log(err)
	t.Log(tt.String())
}
