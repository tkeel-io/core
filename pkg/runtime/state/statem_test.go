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

package state

import (
	"context"
	"testing"
	"time"

	"github.com/shamaton/msgpack/v2"
	"github.com/stretchr/testify/assert"
	"github.com/tkeel-io/collectjs"
	"github.com/tkeel-io/collectjs/pkg/json/jsonparser"
	"github.com/tkeel-io/core/pkg/repository/dao"
	"github.com/tkeel-io/core/pkg/util"
	xjson "github.com/tkeel-io/core/pkg/util/json"
	"github.com/tkeel-io/tdtl"
)

func TestNewStatem(t *testing.T) {
	base := dao.Entity{
		ID:          "device123",
		Type:        "DEVICE",
		Owner:       "admin",
		Source:      "dm",
		Version:     0,
		LastTime:    util.UnixMilli(),
		Properties:  map[string]tdtl.Node{"temp": tdtl.IntNode(25)},
		ConfigBytes: nil,
	}

	sm, err := NewState(context.Background(), &base, nil, nil, nil)
	assert.Nil(t, err)
	assert.Equal(t, "device123", sm.GetID())
}

func TestGJson(t *testing.T) {
	bytes, _ := jsonparser.Set([]byte(``), []byte(`"sss"`), "aa.a")
	t.Log(string(bytes))
}

func TestState_Patch(t *testing.T) {
	stateIns := State{ID: "test", Props: make(map[string]tdtl.Node)}

	stateIns.Patch(xjson.OpAdd, "aa.b.c.c[0]", []byte(`123`))

	t.Log(stateIns.Props)
}

var state1 = &dao.Entity{
	ID:       "d225317a-9ac9-4f52-bf6f-331cd04bbd98",
	Type:     "template",
	Owner:    "usr-33737945c2b718db4c309d633d2f",
	Source:   "device",
	Version:  1,
	LastTime: time.Now().UnixNano() / 1e6,
	ConfigBytes: []byte(`{
		"attributes": {
			"define": {
				"fields": {
					"serial-2": {
						"define": {
							"default_value": "xxxxxxxx",
							"rw": "r"
						},
						"description": "设备批次",
						"enabled": false,
						"enabled_search": false,
						"enabled_time_series": false,
						"id": "serial-2",
						"last_time": 1645525768071,
						"name": "序列号",
						"type": "string",
						"weight": 0
					},
					"version": {
						"define": {
							"default_value": "xxxxxxxx",
							"rw": "r"
						},
						"description": "设备固件版本",
						"enabled": false,
						"enabled_search": false,
						"enabled_time_series": false,
						"id": "version",
						"last_time": 1645525801931,
						"name": "版本",
						"type": "string",
						"weight": 0
					}
				}
			},
			"description": "",
			"enabled": true,
			"enabled_search": true,
			"enabled_time_series": true,
			"id": "attributes",
			"last_time": 1645525768071,
			"name": "",
			"type": "struct",
			"weight": 0
		},
		"commands": {
			"define": {
				"fields": {
					"ota": {
						"define": {
							"fields": {},
							"input": {
								"define": {},
								"id": "ota_send",
								"name": "ota_send",
								"type": "json"
							},
							"mode": "sync",
							"output": {
								"define": {
									"eq": "ok"
								},
								"id": "ota_return",
								"name": "ota_return",
								"type": "string"
							}
						},
						"description": "在线升级",
						"enabled": false,
						"enabled_search": false,
						"enabled_time_series": false,
						"id": "ota",
						"last_time": 1645525862295,
						"name": "在线升级",
						"type": "struct",
						"weight": 0
					}
				}
			},
			"description": "",
			"enabled": true,
			"enabled_search": true,
			"enabled_time_series": true,
			"id": "commands",
			"last_time": 1645525862295,
			"name": "",
			"type": "struct",
			"weight": 0
		},
		"telemetry": {
			"define": {
				"fields": {
					"electricity": {
						"define": {
							"ext": {
								"alias": "EM_BI",
								"ratio_of_transformation": "0.001"
							},
							"max": "1000",
							"min": "0",
							"step": "0.1",
							"unit": "A",
							"unitName": "安"
						},
						"description": "A相电流",
						"enabled": false,
						"enabled_search": false,
						"enabled_time_series": false,
						"id": "electricity",
						"last_time": 1645525836294,
						"name": "电流",
						"type": "int",
						"weight": 0
					},
					"voltage": {
						"define": {
							"ext": {
								"alias": "EM_BI",
								"ratio_of_transformation": "0.001"
							},
							"max": "1000",
							"min": "0",
							"step": "0.1",
							"unit": "v",
							"unitName": "伏"
						},
						"description": "A相电压",
						"enabled": false,
						"enabled_search": false,
						"enabled_time_series": false,
						"id": "voltage",
						"last_time": 1645525824385,
						"name": "电压",
						"type": "int",
						"weight": 0
					}
				}
			},
			"description": "",
			"enabled": true,
			"enabled_search": true,
			"enabled_time_series": true,
			"id": "telemetry",
			"last_time": 1645525824385,
			"name": "",
			"type": "struct",
			"weight": 0
		}
	}`),
	PropertyBytes: []byte(`{
		"basicInfo": {
			"description": "test",
			"name": "template1"
		},
		"sysField": {
			"_createdAt": 1645525739581,
			"_id": "d225317a-9ac9-4f52-bf6f-331cd04bbd98",
			"_owner": "usr-33737945c2b718db4c309d633d2f",
			"_source": "device",
			"_updatedAt": 1645525739581
		},
		"attributes": {
			"metrics": {
				"temp": 26,
				"cpu_used": 0.23,
				"mem_used": 0.78
			}
		}
	}`),
}

func BenchmarkStateJson(b *testing.B) {
	stateInts := state1
	//  patch property.
	for n := 0; n < b.N; n++ {
		stateInts.PropertyBytes, _ = collectjs.Set(stateInts.PropertyBytes, "attributes.metrics.temp", []byte(`40`))
		collectjs.Get(stateInts.PropertyBytes, "attributes.metrics.temp")
	}
}

func BenchmarkStateMap(b *testing.B) {
	// 实体的状态时整个map[string]json.
	bytes, err := msgpack.Marshal(state1)
	assert.Nil(b, err)
	var en dao.Entity

	err = dao.GetEntityCodec().Decode(bytes, &en)
	assert.Nil(b, err)

	b.ResetTimer()
	// patch property.
	stateIns := State{ID: en.ID, Props: en.Properties}
	for n := 0; n < b.N; n++ {
		stateIns.Patch(xjson.OpReplace, "attributes.metrics.temp", []byte(`40`))
		stateIns.Patch(xjson.OpCopy, "attributes.metrics.temp", nil)
	}
}

type A struct {
	Props map[string]string
}

type B struct {
	A
	caches map[string]map[string]string
}

func NewB() *B {
	b := &B{
		A:      A{Props: make(map[string]string)},
		caches: make(map[string]map[string]string),
	}
	b.caches["A"] = b.Props
	return b
}

func TestMapReference(t *testing.T) {
	b := NewB()
	b.Props["name"] = "tom"
	t.Log(b.caches)
}

func TestPatch(t *testing.T) {
	stateIns := State{ID: "xxxxxxxx", Props: map[string]tdtl.Node{
		"temp":    tdtl.New("20"),
		"metrics": tdtl.New(`{"name": "tom"}`),
	}}

	val, err := stateIns.Get("metrics.name")
	t.Log(err)
	t.Log(val.String())
	t.Log(string(val.Raw()))

	stateIns.Patch(xjson.OpReplace, "sysField._spacePath", []byte(`"tom"`))

	t.Log(stateIns)
}
