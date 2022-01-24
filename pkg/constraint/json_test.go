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

package constraint

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

func TestNewNode(t *testing.T) {
	valInt := 23
	valFloat := 55.8

	assert.Equal(t, NewNode(true), BoolNode(true), "BoolNode<true>.")
	assert.Equal(t, NewNode(false), BoolNode(false), "BoolNode<false>.")
	assert.Equal(t, NewNode(int(1)), IntNode(1), "IntNode.")
	assert.Equal(t, NewNode(1.2), FloatNode(1.2), "FloatNode.")
	assert.Equal(t, NewNode(-22.1).To(String), StringNode("-22.1"), "StringNode.")
	assert.Equal(t, NewNode(&valInt), IntNode(valInt), "IntNode PTR.")
	assert.Equal(t, NewNode(&valFloat), FloatNode(valFloat), "FloatNode PTR.")
	assert.Equal(t, NewNode([]byte("test bytes.")), JSONNode("test bytes."), "RawNode PTR.")

	t.Log(NewNode(-22.1).To(String).String())

	time.Sleep(time.Second)
}

func TestEncodeJSON(t *testing.T) {
	kvalues := map[string]Node{
		"id":         NewNode("iotd-device123"),
		"name":       NewNode("device123"),
		"type":       NewNode("DEVICE"),
		"temp":       NewNode(25),
		"cpu_used":   NewNode(0.25),
		"interfaces": NewNode([]string{"eth0"}),
		"conns": NewNode(map[string]interface{}{
			"hello": "world",
		}),
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
	kvalues := map[string]Node{
		"id":         NewNode("iotd-device123"),
		"name":       NewNode("device123"),
		"type":       NewNode("DEVICE"),
		"temp":       NewNode(25),
		"cpu_used":   NewNode(0.25),
		"interfaces": NewNode([]string{"eth0"}),
		"conns": NewNode(map[string]interface{}{
			"hello":     "world",
			"helloss":   "world",
			"hellosss":  "world",
			"hellossss": "world",
			"hellosx":   "world",
			"test": map[string]interface{}{
				"conns":      map[string]interface{}{"hello": "world"},
				"id":         "iotd-device123",
				"name":       "device123",
				"type":       "DEVICE",
				"temp":       25,
				"cpu_used":   0.25,
				"interfaces": []string{"eth888", "xxxx"},
			},
		}),
		"conns1": NewNode(map[string]interface{}{
			"hello":     "world",
			"helloss":   "world",
			"hellosss":  "world",
			"hellossss": "world",
			"hellosx":   "world",
			"test": map[string]interface{}{
				"conns":      map[string]interface{}{"hello": "world"},
				"id":         "iotd-device123",
				"name":       "device123",
				"type":       "DEVICE",
				"temp":       25,
				"cpu_used":   0.25,
				"interfaces": []string{"eth888", "xxxx"},
			},
		}),
		"conns2": NewNode(map[string]interface{}{
			"hello":     "world",
			"helloss":   "world",
			"hellosss":  "world",
			"hellossss": "world",
			"hellosx":   "world",
			"test": map[string]interface{}{
				"conns":      map[string]interface{}{"hello": "world"},
				"id":         "iotd-device123",
				"name":       "device123",
				"type":       "DEVICE",
				"temp":       25,
				"cpu_used":   0.25,
				"interfaces": []string{"eth888", "xxxx"},
			},
		}),
		"conns3": NewNode(map[string]interface{}{
			"hello":     "world",
			"helloss":   "world",
			"hellosss":  "world",
			"hellossss": "world",
			"hellosx":   "world",
			"test": map[string]interface{}{
				"conns":      map[string]interface{}{"hello": "world"},
				"id":         "iotd-device123",
				"name":       "device123",
				"type":       "DEVICE",
				"temp":       25,
				"cpu_used":   0.25,
				"interfaces": []string{"eth888", "xxxx"},
			},
		}),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		EncodeJSON(kvalues)
	}

	bytes, err := EncodeJSON(kvalues)
	assert.Nil(b, err)
	b.Log(string(bytes))
	assert.Equal(b, true, gjson.Valid(string(bytes)))
}

func TestDecodeJson(t *testing.T) {
	jsonText := `{
		"a":"b",
		"arr": [1,2,3],
		"xxx": {"a":"c", "x":20}
	}`

	res, err := DecodeJSON([]byte(jsonText))
	assert.Nil(t, err)
	t.Log("decode result: ", res)
}
