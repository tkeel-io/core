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
	"encoding/base64"
	"fmt"
	"testing"
	"time"

	_ "github.com/ClickHouse/clickhouse-go/v2"
	_ "github.com/tkeel-io/core/pkg/resource/tseries/builder"

	"github.com/stretchr/testify/assert"
	v1 "github.com/tkeel-io/core/api/core/v1"
	"github.com/tkeel-io/core/pkg/placement"
	"github.com/tkeel-io/core/pkg/repository"
	"github.com/tkeel-io/core/pkg/resource"
	"github.com/tkeel-io/core/pkg/resource/tseries"
	tkeelJson "github.com/tkeel-io/core/pkg/util/json"
	"github.com/tkeel-io/core/pkg/util/path"
	"github.com/tkeel-io/tdtl"
)

func TestTTDL(t *testing.T) {
	cc := tdtl.New([]byte(`{"a": 20}`))
	cc.Set("", tdtl.New("{}"))
	t.Log(cc.Error())
	t.Log(string(cc.Raw()))
	t.Log(cc.Get("").String())
}

func Test_adjustTSData(t *testing.T) {
	in := []byte(`{"subOffline":334,"a":"abc"}`)
	out := adjustTSData(in, nil)
	t.Log(string(out))
	in = []byte(`{"ts":1646954803319,"values":{"humidity5":83.0,"temperature5":43.6}}`)
	out = adjustTSData(in, nil)
	t.Log(string(out))

	in = []byte(`{"ModBus-TCP":{"ts":1649215733364,"values":{"wet":42,"temperature":"abc"}},"OPC-UA":{"ts":1649215733364,"values":{"counter":15}}}`)
	out = adjustTSData(in, nil)
	t.Log(string(out))
}

var expr1 = `{"ID":"/core/v1/expressions/usr-57bea3a2d74e21ebbedde8268610/iotd-b10bcaa1-ba98-4e03-bece-6f852feb6edf/properties.telemetry.yc1",
"Path":"properties.telemetry.yc1",
"Name":"", "Type":"eval", "Owner":"usr-57bea3a2d74e21ebbedde8268610",
"EntityID":"iotd-b10bcaa1-ba98-4e03-bece-6f852feb6edf",
"Expression":"iotd-06a96c8d-c166-447c-afd1-63010636b362.properties.telemetry.src1",
"Description":"iotd-06a96c8d-c166-447c-afd1-63010636b362=映射3,yc1=遥测1"}`

var expr2 = `{"ID":"/core/v1/expressions/usr-57bea3a2d74e21ebbedde8268610/iotd-b10bcaa1-ba98-4e03-bece-6f852feb6edf/properties.telemetry.yc2",
"Path":"properties.telemetry.yc2",
"Name":"", "Type":"eval", "Owner":"usr-57bea3a2d74e21ebbedde8268610",
"EntityID":"iotd-b10bcaa1-ba98-4e03-bece-6f852feb6edf",
"Expression":"iotd-06a96c8d-c166-447c-afd1-63010636b362.properties.telemetry.src2",
"Description":"iotd-06a96c8d-c166-447c-afd1-63010636b362=映射3,yc1=遥测1"}`

var state = `{
	"properties": {
		"_version": {"type": "number"},
		"ts": {"type": "number"},
		"telemetry": {
			"FFFF": {
                "ts": 1560000000000,
				"value": "0"
			},
			"GGGG": {
                "ts": 1560000000000,
				"value": 0
			},
			"fire_confidence": {
                "ts": 1560000000000,
				"value": 9.46
			},
			"smoke_confidence": {
                "ts": 1560000000000,
				"value": 9.46
			},
			"switch": {
                "ts": 1560000000000,
				"value": 9.46
			},
			"temperature": {
                "ts": 1560000000000,
				"value": 9.46
			},
			"video_time_offset": {
                "ts": 1560000000000,
				"value": 9.46
			}
		}
	}
}`

var entityStr = `
{
  "id": "iotd-d9f4d19f-9081-4fd2-b0fb-66ca4304d1af",
  "type": "device",
  "owner": "usr-0ae554490ce9653f3d9b16a77963",
  "source": "device",
  "version": 16153151,
  "last_time": 1676340139240,
  "mappers": null,
  "template_id": "iotd-7caf350d-be51-4b16-9a92-86918102d0a8",
  "properties": {
    "basicInfo": {
      "customId": "",
      "description": "",
      "directConnection": false,
      "ext": {},
      "extBusiness": null,
      "name": "test映射",
      "parentId": "iotd-ac009050-ef6e-497b-ad73-90ab53b4206b",
      "parentName": "赣州演示",
      "selfLearn": false,
      "templateId": "iotd-7caf350d-be51-4b16-9a92-86918102d0a8",
      "templateName": "test映射"
    },
    "connectInfo": {
      "_clientId": "",
      "_online": false,
      "_peerHost": "",
      "_protocol": "",
      "_sockPort": "",
      "_userName": ""
    },
    "sysField": {
      "_createdAt": 1664417637948,
      "_enable": true,
      "_id": "iotd-d9f4d19f-9081-4fd2-b0fb-66ca4304d1af",
      "_owner": "usr-0ae554490ce9653f3d9b16a77963",
      "_source": "device",
      "_spacePath": "iotd-ac009050-ef6e-497b-ad73-90ab53b4206b/iotd-d9f4d19f-9081-4fd2-b0fb-66ca4304d1af",
      "_status": "offline",
      "_subscribeAddr": "",
      "_tenantId": "AJJfReg1",
      "_token": "ODUxNmQ5ODQtZjNiNS0zODczLTk5Y2MtN2FhMzliMzMxMzRk",
      "_updatedAt": 1664417637948
    },
    "telemetry": {
      "test": {
        "ts": 1676340075305,
        "value": 37.50578689575195
      },
      "test11": {
        "ts": 1676340075161,
        "value": 1.7777780294418335
      }
    }
  },
  "scheme": {
    "telemetry": {
      "id": "telemetry",
      "type": "struct",
      "name": "telemetry",
      "weight": 0,
      "enabled": true,
      "enabled_search": true,
      "enabled_time_series": true,
      "description": "",
      "define": {
        "fields": {
          "field_float": {
            "define": {
              "ext": {}
            },
            "description": "",
            "enabled": false,
            "enabled_search": false,
            "enabled_time_series": false,
            "id": "test",
            "last_time": 1664417614285,
            "name": "test",
            "type": "float",
            "weight": 0
          },
          "field_bool": {
            "define": {
              "ext": {}
            },
            "description": "",
            "enabled": false,
            "enabled_search": false,
            "enabled_time_series": false,
            "id": "test",
            "last_time": 1664417614285,
            "name": "test",
            "type": "bool",
            "weight": 0
          },
          "test11": {
            "id": "test11",
            "type": "double",
            "name": "test11",
            "weight": 0,
            "enabled": false,
            "enabled_search": false,
            "enabled_time_series": false,
            "description": "",
            "define": {
              "ext": {}
            },
            "last_time": 1668137819874
          },
          "aa": {
            "id": "aa",
            "type": "int",
            "name": "xzxc",
            "weight": 0,
            "enabled": false,
            "enabled_search": false,
            "enabled_time_series": false,
            "description": "",
            "define": {
              "ext": {}
            },
            "last_time": 1669337874747
          }
        }
      },
      "last_time": 1664417614285
    }
  }
}`

type dispatcherMock struct{}

func (d *dispatcherMock) DispatchToLog(ctx context.Context, bytes []byte) error {
	return nil
}

func (d *dispatcherMock) Dispatch(ctx context.Context, event v1.Event) error {
	return nil
}

func TestRuntime_handleComputed(t *testing.T) {
	placement.Initialize()
	placement.Global().Append(placement.Info{
		ID:   "core/1234",
		Flag: true,
	})
	en, err := NewEntity("iotd-06a96c8d-c166-447c-afd1-63010636b362", []byte(state))
	assert.Nil(t, err)

	rt := &Runtime{
		dispatcher: &dispatcherMock{},
		enCache: NewCacheMock(map[string]Entity{
			"iotd-06a96c8d-c166-447c-afd1-63010636b362": en,
		}),
		expressions: map[string]ExpressionInfo{},
		subTree:     path.NewRefTree(),
		evalTree:    path.New(),
	}

	updateExpr(rt, expr1)
	updateExpr(rt, expr2)
	feed := Feed{
		TTL:      0,
		Err:      nil,
		Event:    &v1.ProtoEvent{},
		State:    nil,
		EntityID: "iotd-06a96c8d-c166-447c-afd1-63010636b362",
		Patches:  nil,
		Changes: []Patch{
			{
				Op:    tkeelJson.OpReplace,
				Path:  "properties.telemetry.src1",
				Value: tdtl.New(123),
			},
			{
				Op:    tkeelJson.OpReplace,
				Path:  "properties.telemetry.src2",
				Value: tdtl.New(123),
			},
		},
	}
	got := rt.handleTentacle(context.Background(), &feed)
	t.Log(got)
	got = rt.handleComputed(context.Background(), &feed)
	t.Log(got)

	removeExpr(rt, expr1)
	removeExpr(rt, expr2)
	got = rt.handleTentacle(context.Background(), &feed)
	t.Log(got)
	got = rt.handleComputed(context.Background(), &feed)
	t.Log(got)
}

func TestRuntime_handleEntityEvent(t *testing.T) {
	placement.Initialize()
	placement.Global().Append(placement.Info{
		ID:   "core/1234",
		Flag: true,
	})
	en, err := NewEntity("iotd-06a96c8d-c166-447c-afd1-63010636b362", []byte(state))
	assert.Nil(t, err)

	entity, err := NewEntity("entity-1", []byte(`{
  "id": "iotd-aeb86fb2-e694-431b-b98a-8447c6745817",
  "type": "device",
  "owner": "usr-0008aad7c0f3d28e42f9d5b3448c",
  "source": "device",
  "version": 8707,
  "last_time": 1666788907012,
  "mappers": null,
  "template_id": "iotd-5dfc8ca9-01b1-463a-b60c-5df5f6c7cc81",
  "properties": {
    "basicInfo": {
      "customId": "",
      "description": "",
      "directConnection": true,
      "ext": {},
      "extBusiness": null,
      "name": "智慧工厂02",
      "parentId": "iotd-usr-0008aad7c0f3d28e42f9d5b3448c-defaultGroup",
      "parentName": "默认分组",
      "selfLearn": false,
      "templateId": "iotd-5dfc8ca9-01b1-463a-b60c-5df5f6c7cc81",
      "templateName": "AI 智慧工厂"
    },
    "connectInfo": {
      "_clientId": "IncomingDataPublisher",
      "_online": true,
      "_peerHost": "192.168.100.8",
      "_protocol": "mqtt",
      "_sockPort": "1883",
      "_userName": "iotd-aeb86fb2-e694-431b-b98a-8447c6745817",
      "_owner": "",
      "_timestamp": 1666788893832
    },
    "sysField": {
      "_createdAt": 1663922154648,
      "_enable": true,
      "_id": "iotd-aeb86fb2-e694-431b-b98a-8447c6745817",
      "_owner": "usr-0008aad7c0f3d28e42f9d5b3448c",
      "_source": "device",
      "_spacePath": "iotd-usr-0008aad7c0f3d28e42f9d5b3448c-defaultGroup/iotd-aeb86fb2-e694-431b-b98a-8447c6745817",
      "_status": "offline",
      "_subscribeAddr": "",
      "_tenantId": "H18fhe6d",
      "_token": "MjNjNjlmYjMtYTFmNy0zNGIwLWIyMDUtODkwZGRkZjc1YWUw",
      "_updatedAt": 1663922154648
    },
    "rawData": {
      "path": "iotd-aeb86fb2-e694-431b-b98a-8447c6745817/v1/devices/me/telemetry",
      "type": "telemetry",
      "mark": "upstream",
      "id": "iotd-aeb86fb2-e694-431b-b98a-8447c6745817",
      "ts": 1666788907007,
      "values": "ewoJIkZGRkYiOiAiIjAuMjMiIiwKCSJ2aWRlb190aW1lX29mZnNldCI6ICIiMi4xMiIiCn0K"
    },
    "telemetry": {
      "FFFF": {
        "ts": 1666787329859,
        "value": 0
      },
      "GGGG": {
        "ts": 1666777942241,
        "value": "0"
      },
      "fire_confidence": {
        "ts": 1666777942241,
        "value": "8.83"
      },
      "smoke_confidence": {
        "ts": 1666777942241,
        "value": "8.17"
      },
      "switch": {
        "ts": 1666777942241,
        "value": "3.68"
      },
      "temperature": {
        "ts": 1666777942241,
        "value": "1.66"
      },
      "video_time_offset": {
        "ts": 1666787329859,
        "value": 7.84
      }
    }
  },
  "scheme": {
    "attributes": {
      "define": {
        "fields": {
          "switch": {
            "define": {
              "default_value": "",
              "rw": "rw"
            },
            "description": "",
            "enabled": false,
            "enabled_search": false,
            "enabled_time_series": false,
            "id": "switch",
            "last_time": 1662713902758,
            "name": "警报器开关",
            "type": "int",
            "weight": 0
          }
        }
      }
    },
    "telemetry": {
      "define": {
        "fields": {
          "FFFF": {
            "define": {
              "ext": {}
            },
            "description": "",
            "enabled": false,
            "enabled_search": false,
            "enabled_time_series": false,
            "id": "FFFF",
            "last_time": 1664507581012,
            "name": "FFFF",
            "type": "int",
            "weight": 0
          },
          "GGGG": {
            "define": {
              "ext": {}
            },
            "description": "",
            "enabled": false,
            "enabled_search": false,
            "enabled_time_series": false,
            "id": "GGGG",
            "last_time": 1664507265492,
            "name": "GGGG",
            "type": "int",
            "weight": 0
          },
          "fire_confidence": {
            "define": {
              "ext": {}
            },
            "description": "",
            "enabled": false,
            "enabled_search": false,
            "enabled_time_series": false,
            "id": "fire_confidence",
            "last_time": 1662703177269,
            "name": "置信度（火焰）",
            "type": "float",
            "weight": 0
          },
          "smoke_confidence": {
            "define": {
              "ext": {}
            },
            "description": "",
            "enabled": false,
            "enabled_search": false,
            "enabled_time_series": false,
            "id": "smoke_confidence",
            "last_time": 1662703260929,
            "name": "置信度（烟雾）",
            "type": "float",
            "weight": 0
          },
          "switch": {
            "define": {
              "ext": {}
            },
            "description": "",
            "enabled": false,
            "enabled_search": false,
            "enabled_time_series": false,
            "id": "switch",
            "last_time": 1662717093545,
            "name": "警报器开关",
            "type": "int",
            "weight": 0
          },
          "temperature": {
            "define": {
              "ext": {}
            },
            "description": "",
            "enabled": false,
            "enabled_search": false,
            "enabled_time_series": false,
            "id": "temperature",
            "last_time": 1662703006245,
            "name": "工厂温度",
            "type": "int",
            "weight": 0
          },
          "video_time_offset": {
            "define": {
              "ext": {}
            },
            "description": "",
            "enabled": false,
            "enabled_search": false,
            "enabled_time_series": false,
            "id": "video_time_offset",
            "last_time": 1663836149036,
            "name": "视频帧位置",
            "type": "float",
            "weight": 0
          }
        }
      },
      "description": "",
      "enabled": true,
      "enabled_search": true,
      "enabled_time_series": true,
      "id": "telemetry",
      "last_time": 1662702862347,
      "name": "telemetry",
      "type": "struct",
      "weight": 0
    }
  }
}
`))
	assert.Nil(t, err)

	// default time series.
	cfg := resource.Metadata{
		Name: "clickhouse",
		Properties: map[string]interface{}{
			"urls":     []string{"clickhouse://default:xxxx@127.0.0.1:9000"},
			"database": "core",
			"table":    "timeseries",
		},
	}
	tsdbClient := tseries.NewTimeSerier(cfg.Name)
	err = tsdbClient.Init(cfg)
	assert.Nil(t, err)

	rt := &Runtime{
		dispatcher: &dispatcherMock{},
		entities: map[string]Entity{
			entity.ID(): entity,
		},
		enCache: NewCacheMock(map[string]Entity{
			"iotd-06a96c8d-c166-447c-afd1-63010636b362": en,
		}),
		expressions: map[string]ExpressionInfo{},
		subTree:     path.NewRefTree(),
		evalTree:    path.New(),
		entityResourcer: EntityResource{
			PersistentEntity: func(ctx context.Context, e Entity, feed *Feed) error {
				flushData, tsCount, err := makeTimeSeriesData(ctx, en, feed)
				t.Log(flushData, tsCount, err)

				ret, err := tsdbClient.Write(ctx, flushData)
				if err != nil {
					panic(err)
				}
				t.Log(ret)
				return err
			},
			FlushHandler: func(ctx context.Context, e Entity, feed *Feed) error {
				return nil
			},
			RemoveHandler: func(ctx context.Context, e Entity, feed *Feed) error {
				return nil
			},
		},
	}

	ev := &v1.ProtoEvent{
		Id:        "ev-12345",
		Timestamp: time.Now().UnixNano(),
		Metadata: map[string]string{
			v1.MetaEntityID: "entity-1",
		},
		Data: &v1.ProtoEvent_Patches{
			Patches: &v1.PatchDatas{
				Patches: []*v1.PatchData{
					{
						Path:     "properties.rawData",
						Operator: "replace",
						Value:    []byte(`{"id":"iotd-aeb86fb2-e694-431b-b98a-8447c6745817","ts":1666790831700,"values":"ewogICAgICAgICJ0cyI6IDE1NjAwMDAwLAoJIkZGRkYiOiAwLAoJIkdHR0ciOiAiMCIsCgkiZmlyZV9jb25maWRlbmNlIjogIjkuNDYiLAoJInNtb2tlX2NvbmZpZGVuY2UiOiAiNC42NCIsCgkic3dpdGNoIjogIjEuODIiLAoJInRlbXBlcmF0dXJlIjogIjYuNDQiLAoJInZpZGVvX3RpbWVfb2Zmc2V0IjogIjkuODciCn0=","path":"iotd-aeb86fb2-e694-431b-b98a-8447c6745817/v1/devices/me/telemetry","type":"telemetry","mark":"upstream"}`),
					},
				},
			},
		},
	}

	execer, feed := rt.handleEntityEvent(ev)
	newFeed := execer.Exec(context.Background(), feed)

	t.Log(newFeed)
	select {}
}

func makeExpr(exprRaw string) repository.Expression {
	expr := repository.Expression{}
	expr.Decode([]byte(""), []byte(exprRaw))
	return expr
}

func makeExprInfos(expr repository.Expression) map[string]*ExpressionInfo {
	exprInfo := newExprInfo(&expr)

	exprInfos, err := parseExpression(exprInfo.Expression, 0)
	if nil != err {
		panic(err)
	}
	return exprInfos
}

func updateExpr(rt *Runtime, exprRaw string) repository.Expression {
	expr := repository.Expression{}
	expr.Decode([]byte(""), []byte(exprRaw))
	exprInfo1, err := parseExpression(expr, 1)
	if err != nil {
		panic(err)
	}
	for _, exprIns := range exprInfo1 {
		rt.AppendExpression(*exprIns)
	}
	return expr
}

func removeExpr(rt *Runtime, exprRaw string) repository.Expression {
	expr := repository.Expression{}
	expr.Decode([]byte(""), []byte(exprRaw))
	exprInfo := newExprInfo(&expr)
	rt.RemoveExpression(exprInfo.ID)
	return expr
}

func Test_mergePath(t *testing.T) {
	tests := []struct {
		name       string
		subPath    string
		changePath string
		want       string
	}{
		{"1", "dev1.a.b.*", "dev1.a.b.c", "a.b"},
		{"1", "dev1.a.b.c.d", "dev1.a.b.c", "a.b.c.d"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mergePath(tt.subPath, tt.changePath); got != tt.want {
				t.Errorf("mergePath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRuntime_AppendExpression(t *testing.T) {
	placement.Initialize()
	placement.Global().Append(placement.Info{
		ID:   "core/1234",
		Flag: true,
	})
	en, err := NewEntity("iotd-06a96c8d-c166-447c-afd1-63010636b362", []byte(state))
	assert.Nil(t, err)

	rt := &Runtime{
		dispatcher: &dispatcherMock{},
		enCache: NewCacheMock(map[string]Entity{
			"iotd-06a96c8d-c166-447c-afd1-63010636b362": en,
		}),
		expressions: map[string]ExpressionInfo{},
		subTree:     path.NewRefTree(),
		evalTree:    path.New(),
	}

	ex1 := makeExpr(expr1)
	ex2 := makeExpr(expr2)

	exInfo1 := makeExprInfos(ex1)
	exInfo2 := makeExprInfos(ex2)

	// delivery expression.
	for _, exprItem := range exInfo1 {
		rt.AppendExpression(*exprItem)
	}
	for _, exprItem := range exInfo2 {
		rt.AppendExpression(*exprItem)
	}
	rt.RemoveExpression(ex1.ID)
	rt.RemoveExpression(ex2.ID)
	t.Log(rt)
}

func Test_parsePayload(t *testing.T) {
	t.Logf("It is now %s\n", time.Now())

	tests := []struct {
		name     string
		arg      []byte
		hasErr   bool
		wantSize int
	}{
		{"1", []byte(`{"ts":11111, "values":{"a":1, "b":2}}`), false, 2},
		{"2", []byte(`{"ts":11111, "a":1, "b":2}`), false, 2},
		{"2", []byte(`{ "a":1, "b":2}`), false, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, got1, _, err := parsePayload(tt.arg)
			assert.Nil(t, err, "parsePayload(%v)", tt.arg)
			// assert.Equalf(t, tt.wantTS, got, "parsePayload(%v)", tt.arg)
			assert.Equalf(t, tt.wantSize, len(got1), "parsePayload(%v)", tt.arg)
		})
	}
}

func Example_handleRawData() {
	V := tdtl.New([]byte("{\"_1001_HW_DI060504_PV\":0.000000,\"_1001_HW_DI060505_PV\":0.000000,\"_1001_HW_DI060506_PV\":0.000000,\"_1001_HW_DI060507_PV\":0.000000,\"_1001_HW_DI060508_PV\":1.000000,\"_1001_HW_DI060509_PV\":0.000000,\"_1001_HW_DI060510_PV\":0.000000,\"_1001_HW_DI060511_PV\":0.000000,\"_1001_HW_DI060512_PV\":0.000000,\"_1001_HW_DI060513_PV\":0.000000,\"_1001_HW_DI060514_PV\":0.000000,\"_1001_HW_DI060515_PV\":0.000000,\"_1001_HW_DI060516_PV\":0.000000,\"_1001_HW_DI070501_PV\":1.000000,\"_1001_HW_DI070502_PV\":1.000000,\"_1001_HW_DI070503_PV\":1.000000,\"_1001_HW_DI070504_PV\":1.000000,\"_1001_HW_DI070505_PV\":0.000000,\"_1001_HW_DI070506_PV\":1.000000,\"_1001_HW_DI070507_PV\":0.000000,\"_1001_HW_DI070508_PV\":1.000000,\"_1001_HW_DI070509_PV\":0.000000,\"_1001_HW_DI070510_PV\":1.000000,\"_1001_HW_DI070511_PV\":1.000000,\"_1001_HW_DI070512_PV\":0.000000,\"_1001_HW_DI070513_PV\":1.000000,\"_1001_HW_DI070514_PV\":1.000000,\"_1001_HW_DI070515_PV\":0.000000,\"_1001_HW_DI070516_PV\":0.000000,\"_1001_HW_DI080401_PV\":1.000000,\"_1001_HW_DI080402_PV\":0.000000,\"_1001_HW_DI080403_PV\":1.000000,\"_1001_HW_DI080404_PV\":0.000000,\"_1001_HW_DI080405_PV\":1.000000,\"_1001_HW_DI080406_PV\":1.000000,\"_1001_HW_DI080407_PV\":1.000000,\"_1001_HW_DI080408_PV\":0.000000,\"_1001_HW_DI080409_PV\":1.000000,\"_1001_HW_DI080410_PV\":0.000000,\"_1001_HW_DI080411_PV\":1.000000,\"_1001_HW_DI080412_PV\":0.000000,\"_1001_HW_DI080413_PV\":1.000000,\"_1001_HW_DI080414_PV\":0.000000,\"_1001_HW_DI080415_PV\":1.000000,\"_1001_HW_DI080416_PV\":0.000000,\"_1001_HW_DI080501_PV\":1.000000,\"_1001_HW_DI080502_PV\":0.000000,\"_1001_HW_DI080503_PV\":1.000000,\"_1001_HW_DI080504_PV\":0.000000,\"_1001_HW_DI080505_PV\":1.000000,\"_1001_HW_DI080506_PV\":0.000000,\"_1001_HW_DI080507_PV\":1.000000,\"_1001_HW_DI080508_PV\":0.000000,\"_1001_HW_DI080509_PV\":1.000000,\"_1001_HW_DI080510_PV\":0.000000,\"_1001_HW_DI080511_PV\":1.000000,\"_1001_HW_DI080512_PV\":0.000000,\"_1001_HW_DI080513_PV\":1.000000,\"_1001_HW_DI080514_PV\":0.000000,\"_1001_HW_DI080515_PV\":1.000000,\"_1001_HW_DI080516_PV\":0.000000,\"_1001_HW_DI090301_PV\":1.000000,\"_1001_HW_DI090302_PV\":0.000000,\"_1001_HW_DI090303_PV\":0.000000,\"_1001_HW_DI090304_PV\":0.000000,\"_1001_HW_DI090305_PV\":0.000000,\"_1001_HW_DI090306_PV\":0.000000,\"_1001_HW_DI090307_PV\":0.000000,\"_1001_HW_DI090308_PV\":0.000000,\"_1001_HW_DI090309_PV\":0.000000,\"_1001_HW_DI090310_PV\":0.000000,\"_1001_HW_DI090311_PV\":1.000000,\"_1001_HW_DI090312_PV\":1.000000,\"_1001_HW_DI090313_PV\":0.000000,\"_1001_HW_DI090314_PV\":0.000000,\"_1001_HW_DI090315_PV\":0.000000,\"_1001_HW_DI090316_PV\":0.000000,\"_1001_HW_DI090401_PV\":1.000000,\"_1001_HW_DI090402_PV\":0.000000,\"_1001_HW_DI090403_PV\":0.000000,\"_1001_HW_DI090404_PV\":0.000000,\"_1001_HW_DI090405_PV\":0.000000,\"_1001_HW_DI090406_PV\":0.000000,\"_1001_HW_DI090407_PV\":0.000000,\"_1001_HW_DI090408_PV\":0.000000,\"_1001_HW_DI090409_PV\":0.000000,\"_1001_HW_DI090410_PV\":1.000000,\"_1001_HW_DI090411_PV\":0.000000,\"_1001_HW_DI090412_PV\":1.000000,\"_1001_HW_DI090413_PV\":1.000000,\"_1001_HW_DI090414_PV\":0.000000,\"_1001_HW_DI090415_PV\":0.000000,\"_1001_HW_DI090416_PV\":0.000000,\"_1001_HW_DI090501_PV\":0.000000,\"_1001_HW_DI090502_PV\":0.000000,\"_1001_HW_DI090503_PV\":0.000000,\"_1001_HW_DI090504_PV\":0.000000,\"_1001_HW_DI090505_PV\":1.000000,\"_1001_HW_DI090506_PV\":0.000000,\"_1001_HW_DI090507_PV\":0.000000,\"_1001_HW_DI090508_PV\":0.000000,\"_1001_HW_DI090509_PV\":0.000000,\"_1001_HW_DI090510_PV\":0.000000,\"_1001_HW_DI090511_PV\":1.000000,\"_1001_HW_DI090512_PV\":0.000000,\"_1001_HW_DI090513_PV\":1.000000,\"_1001_HW_DI090514_PV\":0.000000,\"_1001_HW_DI090515_PV\":1.000000,\"_1001_HW_DI090516_PV\":0.000000,\"_1001_HW_DI120501_PV\":0.000000,\"_1001_HW_DI120502_PV\":0.000000,\"_1001_HW_DI120503_PV\":0.000000,\"_1001_HW_DI120504_PV\":0.000000,\"_1001_HW_DI120505_PV\":0.000000,\"_1001_HW_DI120506_PV\":0.000000,\"_1001_HW_DI120507_PV\":0.000000,\"_1001_HW_DI120508_PV\":0.000000,\"_1001_HW_DI120509_PV\":0.000000,\"_1001_HW_DI120510_PV\":0.000000,\"_1001_HW_DI120511_PV\":0.000000,\"_1001_HW_DI120512_PV\":0.000000,\"_1001_HW_DI120513_PV\":0.000000,\"_1001_HW_DI120514_PV\":0.000000,\"_1001_HW_DI120515_PV\":0.000000,\"_1001_HW_DI120516_PV\":0.000000,\"_1001_HW_RT070301_PV\":41.806629,\"_1001_HW_RT070302_PV\":30.728159,\"_1001_HW_RT070303_PV\":37.446587,\"_1001_HW_RT070304_PV\":617.500000,\"_1001_HW_RT070305_PV\":38.235462,\"_1001_HW_RT070306_PV\":617.500000,\"_1001_HW_RT070307_PV\":617.500000,\"_1001_HW_RT070308_PV\":617.500000,\"_1001_HW_RT080201_PV\":37.575916,\"_1001_HW_RT080202_PV\":37.705231,\"_1001_HW_RT080203_PV\":38.558811,\"_1001_HW_RT080204_PV\":37.692287,\"_1001_HW_RT080205_PV\":32.794170,\"_1001_HW_RT080206_PV\":617.500000,\"_1001_HW_RT080207_PV\":617.500000,\"_1001_HW_RT080208_PV\":617.500000,\"_1001_HW_RT090101_PV\":617.500000,\"_1001_HW_RT090102_PV\":617.500000,\"_1001_HW_RT090103_PV\":617.500000,\"_1001_HW_RT090104_PV\":617.500000,\"_1001_HW_RT090105_PV\":617.500000,\"_1001_HW_RT090106_PV\":617.500000,\"_1001_HW_RT090107_PV\":617.500000,\"_1001_HW_RT090108_PV\":21.393274,\"_1001_HW_RT100201_PV\":16.850239,\"_1001_HW_RT100203_PV\":28.263302,\"_1001_HW_RT100207_PV\":118.363228,\"_1001_HW_RT100208_PV\":30.986353,\"_1001_HW_RT100301_PV\":69.629410,\"_1001_HW_RT100302_PV\":198.922623,\"_1001_HW_RT100303_PV\":21.444784,\"_1001_HW_RT100304_PV\":22.449375,\"_1001_HW_RT100305_PV\":26.483381,\"_1001_HW_RT100306_PV\":22.372091,\"_1001_HW_RT100307_PV\":22.848717,\"_1001_HW_RT100308_PV\":22.784311,\"_1001_HW_RT100401_PV\":34.964619,\"_1001_HW_RT100402_PV\":36.670815,\"_1001_HW_RT100403_PV\":39.684258,\"_1001_HW_RT100404_PV\":42.570431,\"_1001_HW_RT100405_PV\":50.034515,\"_1001_HW_RT100406_PV\":50.683060,\"_1001_HW_RT100407_PV\":51.539288,\"_1001_HW_RT100408_PV\":30.728159,\"_1001_HW_RT100501_PV\":49.321236,\"_1001_HW_RT100502_PV\":46.793304,\"_1001_HW_RT100503_PV\":47.972828,\"_1001_HW_RT100504_PV\":23.222336,\"_1001_HW_RT100505_PV\":24.098541,\"_1001_HW_RT100506_PV\":36.205418,\"_1001_HW_RT100507_PV\":28.069777,\"_1001_HW_RT100508_PV\":31.050894,\"_1001_HW_RT110201_PV\":30.883068,\"_1001_HW_RT110202_PV\":30.753973,\"_1001_HW_RT110203_PV\":31.218731,\"_1001_HW_RT110204_PV\":26.328642,\"_1001_HW_RT110205_PV\":25.928921,\"_1001_HW_RT110206_PV\":30.883068,\"_1001_HW_RT110207_PV\":73.224396,\"_1001_HW_RT110208_PV\":20.504822,\"_1001_HW_RT120201_PV\":31.541525,\"_1001_HW_RT120202_PV\":31.141277,\"_1001_HW_RT120203_PV\":59.759583,\"_1001_HW_RT120204_PV\":49.632477,\"_1001_HW_RT120205_PV\":80.504448,\"_1001_HW_RT120206_PV\":73.119759,\"_1001_HW_RT120207_PV\":77.517204,\"_1001_HW_RT120208_PV\":75.671143,\"_1001_HW_RT120301_PV\":77.504120,\"_1001_HW_RT120302_PV\":77.242203,\"_1001_HW_RT120303_PV\":84.898422,\"_1001_HW_RT120304_PV\":32.393776,\"_1001_HW_RT120305_PV\":33.556263,\"_1001_HW_RT120306_PV\":53.083485,\"_1001_HW_RT120307_PV\":41.677189,\"_1001_HW_RT120308_PV\":617.500000,\"_1001_HW_RT120401_PV\":49.256416,\"_1001_HW_RT120402_PV\":50.436611,\"_1001_HW_RT120403_PV\":53.693520,\"_1001_HW_RT120404_PV\":61.878757,\"_1001_HW_RT120405_PV\":60.487522,\"_1001_HW_RT120406_PV\":61.072567,\"_1001_HW_RT120407_PV\":73.642967,\"_1001_HW_RT120408_PV\":617.500000,\"_1001_HW_TC070401_PV\":387.869080,\"_1001_HW_TC070402_PV\":254.179199,\"_1001_HW_TC070403_PV\":353.575409,\"_1001_HW_TC070404_PV\":307.199402,\"_1001_HW_TC070405_PV\":415.649994,\"_1001_HW_TC070406_PV\":415.649994,\"_1001_HW_TC070407_PV\":253.995499,\"_1001_HW_TC070408_PV\":254.975159,\"_1001_HW_TC080301_PV\":254.699692,\"_1001_HW_TC080302_PV\":252.525330,\"_1001_HW_TC080303_PV\":253.444260,\"_1001_HW_TC080304_PV\":252.433395,\"_1001_HW_TC080305_PV\":164.714584,\"_1001_HW_TC080306_PV\":166.182419,\"_1001_HW_TC080307_PV\":156.298096,\"_1001_HW_TC080308_PV\":128.250519,\"_1001_HW_TC090201_PV\":1231.250000,\"_1001_HW_TC090202_PV\":1231.250000,\"_1001_HW_TC090203_PV\":414.350769,\"_1001_HW_TC090204_PV\":1000.082825,\"_1001_HW_TC090205_PV\":1000.082825,\"_1001_HW_TC090206_PV\":1000.082825,\"_1001_HW_TC090207_PV\":1000.082825,\"_1001_HW_TC090208_PV\":1000.082825,\"_1001_HW_TC100601_PV\":24.031992,\"_1001_HW_TC100602_PV\":194.823456,\"_1001_HW_TC100603_PV\":179.698685,\"_1001_HW_TC100604_PV\":192.667862,\"_1001_HW_TC100605_PV\":203.118210,\"_1001_HW_TC100606_PV\":152.442245,\"_1001_HW_TC100607_PV\":535.039429,\"_1001_HW_TC100608_PV\":1231.250000,\"_1001_HW_TC100701_PV\":914.878967,\"_1001_HW_TC100702_PV\":912.467896,\"_1001_HW_TC100703_PV\":118.023125,\"_1001_HW_TC100704_PV\":97.233864,\"_1001_HW_TC100708_PV\":1231.250000,\"_1001_HW_TC110307_PV\":991.780945,\"_1001_HW_TC110308_PV\":1231.250000,\"_1001_HW_TC110401_PV\":983.006409,\"_1001_HW_TC110404_PV\":750.689270,\"_1001_HW_TC110405_PV\":738.990845,\"_1001_HW_TC110408_PV\":1231.250000,\"_1001_HW_TC110501_PV\":539.493347,\"_1001_HW_TC110502_PV\":519.510193,\"_1001_HW_TC110503_PV\":451.180145,\"_1001_HW_TC110504_PV\":446.121338,\"_1001_HW_TC110505_PV\":410.747162,\"_1001_HW_TC110506_PV\":400.516815,\"_1001_HW_TC110507_PV\":360.911163,\"_1001_HW_TC110508_PV\":364.455841,\"_1001_HW_TC110601_PV\":323.064697,\"_1001_HW_TC110602_PV\":318.976379,\"_1001_HW_TC110603_PV\":269.087616,\"_1001_HW_TC110604_PV\":268.477966,\"_1001_HW_TC110605_PV\":225.358551,\"_1001_HW_TC110606_PV\":228.385483,\"_1001_HW_TC110607_PV\":1231.250000,\"_1001_HW_TC110608_PV\":1231.250000,\"_1001_SH0003_AALM1_PV\":0.000000,\"_1001_SH0010_PRO1_IN\":-0.500000,\"_1001_SH0010_PRO2_IN\":9.479815,\"_1001_SH0010_PRO3_IN\":-0.625000,\"_1001_SH0010_PRO4_IN\":6.070231,\"_1001_SH0010_PRO5_IN\":0.023611,\"_1001_SH0010_PRO6_IN\":-0.006019,\"_1001_SH0010_PRO7_IN\":199.134247,\"_1001_SH0010_PRO8_IN\":199.254639,\"_1001_SH0012_AALM2_PV\":8203.656250,\"_1001_SH0013_AALM2_PV\":7217.221680,\"_1001_SH0014_AALM1_PV\":3.019795,\"_1001_SH0015_AALM1_PV\":1.650945,\"_1001_SH0016_AALM1_PV\":8808.333984,\"_1001_SH0016_AALM2_PV\":88150.007813,\"_1001_SH0016_AALM3_PV\":4945.833496,\"_1001_SH0016_AALM4_PV\":6539.750000,\"_1001_SH0016_AALM5_PV\":38745.000000,\"_1001_SH0016_AALM6_PV\":14045.839844,\"_1001_SH0016_AALM7_PV\":11685.000000,\"\u0001\"_1001_HW_DI060504_PV\":0.000000,\":19065.000000,\"_1001_SH0017_AALM1_PV\":-11.572223,\"_1001_SH0017_AALM2_PV\":8.162498,\"_1001_SH0017_AALM3_PV\":1.954865,\"_1001_SH0018_AALM1_PV\":77113.335938,\"_1001_SH0022_AALM3_ACFG\":0.000000,\"_1001_SH0022_AALM3_ACK\":0.000000,\"_1001_SH0022_AALM3_ALM\":0.000000,\"_1001_SH0022_AALM3_EN\":0.000000,\"_1001_SH0022_AALM3_HAL\":10000.000000,\"_1001_SH0022_AALM3_HHAL\":20000.000000,\"_1001_SH0022_AALM3_LAL\":-10000.000000,\"ts\":1675925762171}"))
	values := V.Get("values").String()
	bytes, err := base64.StdEncoding.DecodeString(values)
	if nil != err {
		fmt.Println("attempt extract RawData", V.String())
		return
	}

	if !json.Valid(bytes) {
		fmt.Println("RawData Json Valid Fail", V.String(), "#####", string(bytes))
	} else {
		fmt.Println("extract RawData successful", V.String(), "#####", string(bytes))
	}
	// output:
	// RawData Json Valid Fail {"_1001_HW_DI060504_PV":0.000000,"_1001_HW_DI060505_PV":0.000000,"_1001_HW_DI060506_PV":0.000000,"_1001_HW_DI060507_PV":0.000000,"_1001_HW_DI060508_PV":1.000000,"_1001_HW_DI060509_PV":0.000000,"_1001_HW_DI060510_PV":0.000000,"_1001_HW_DI060511_PV":0.000000,"_1001_HW_DI060512_PV":0.000000,"_1001_HW_DI060513_PV":0.000000,"_1001_HW_DI060514_PV":0.000000,"_1001_HW_DI060515_PV":0.000000,"_1001_HW_DI060516_PV":0.000000,"_1001_HW_DI070501_PV":1.000000,"_1001_HW_DI070502_PV":1.000000,"_1001_HW_DI070503_PV":1.000000,"_1001_HW_DI070504_PV":1.000000,"_1001_HW_DI070505_PV":0.000000,"_1001_HW_DI070506_PV":1.000000,"_1001_HW_DI070507_PV":0.000000,"_1001_HW_DI070508_PV":1.000000,"_1001_HW_DI070509_PV":0.000000,"_1001_HW_DI070510_PV":1.000000,"_1001_HW_DI070511_PV":1.000000,"_1001_HW_DI070512_PV":0.000000,"_1001_HW_DI070513_PV":1.000000,"_1001_HW_DI070514_PV":1.000000,"_1001_HW_DI070515_PV":0.000000,"_1001_HW_DI070516_PV":0.000000,"_1001_HW_DI080401_PV":1.000000,"_1001_HW_DI080402_PV":0.000000,"_1001_HW_DI080403_PV":1.000000,"_1001_HW_DI080404_PV":0.000000,"_1001_HW_DI080405_PV":1.000000,"_1001_HW_DI080406_PV":1.000000,"_1001_HW_DI080407_PV":1.000000,"_1001_HW_DI080408_PV":0.000000,"_1001_HW_DI080409_PV":1.000000,"_1001_HW_DI080410_PV":0.000000,"_1001_HW_DI080411_PV":1.000000,"_1001_HW_DI080412_PV":0.000000,"_1001_HW_DI080413_PV":1.000000,"_1001_HW_DI080414_PV":0.000000,"_1001_HW_DI080415_PV":1.000000,"_1001_HW_DI080416_PV":0.000000,"_1001_HW_DI080501_PV":1.000000,"_1001_HW_DI080502_PV":0.000000,"_1001_HW_DI080503_PV":1.000000,"_1001_HW_DI080504_PV":0.000000,"_1001_HW_DI080505_PV":1.000000,"_1001_HW_DI080506_PV":0.000000,"_1001_HW_DI080507_PV":1.000000,"_1001_HW_DI080508_PV":0.000000,"_1001_HW_DI080509_PV":1.000000,"_1001_HW_DI080510_PV":0.000000,"_1001_HW_DI080511_PV":1.000000,"_1001_HW_DI080512_PV":0.000000,"_1001_HW_DI080513_PV":1.000000,"_1001_HW_DI080514_PV":0.000000,"_1001_HW_DI080515_PV":1.000000,"_1001_HW_DI080516_PV":0.000000,"_1001_HW_DI090301_PV":1.000000,"_1001_HW_DI090302_PV":0.000000,"_1001_HW_DI090303_PV":0.000000,"_1001_HW_DI090304_PV":0.000000,"_1001_HW_DI090305_PV":0.000000,"_1001_HW_DI090306_PV":0.000000,"_1001_HW_DI090307_PV":0.000000,"_1001_HW_DI090308_PV":0.000000,"_1001_HW_DI090309_PV":0.000000,"_1001_HW_DI090310_PV":0.000000,"_1001_HW_DI090311_PV":1.000000,"_1001_HW_DI090312_PV":1.000000,"_1001_HW_DI090313_PV":0.000000,"_1001_HW_DI090314_PV":0.000000,"_1001_HW_DI090315_PV":0.000000,"_1001_HW_DI090316_PV":0.000000,"_1001_HW_DI090401_PV":1.000000,"_1001_HW_DI090402_PV":0.000000,"_1001_HW_DI090403_PV":0.000000,"_1001_HW_DI090404_PV":0.000000,"_1001_HW_DI090405_PV":0.000000,"_1001_HW_DI090406_PV":0.000000,"_1001_HW_DI090407_PV":0.000000,"_1001_HW_DI090408_PV":0.000000,"_1001_HW_DI090409_PV":0.000000,"_1001_HW_DI090410_PV":1.000000,"_1001_HW_DI090411_PV":0.000000,"_1001_HW_DI090412_PV":1.000000,"_1001_HW_DI090413_PV":1.000000,"_1001_HW_DI090414_PV":0.000000,"_1001_HW_DI090415_PV":0.000000,"_1001_HW_DI090416_PV":0.000000,"_1001_HW_DI090501_PV":0.000000,"_1001_HW_DI090502_PV":0.000000,"_1001_HW_DI090503_PV":0.000000,"_1001_HW_DI090504_PV":0.000000,"_1001_HW_DI090505_PV":1.000000,"_1001_HW_DI090506_PV":0.000000,"_1001_HW_DI090507_PV":0.000000,"_1001_HW_DI090508_PV":0.000000,"_1001_HW_DI090509_PV":0.000000,"_1001_HW_DI090510_PV":0.000000,"_1001_HW_DI090511_PV":1.000000,"_1001_HW_DI090512_PV":0.000000,"_1001_HW_DI090513_PV":1.000000,"_1001_HW_DI090514_PV":0.000000,"_1001_HW_DI090515_PV":1.000000,"_1001_HW_DI090516_PV":0.000000,"_1001_HW_DI120501_PV":0.000000,"_1001_HW_DI120502_PV":0.000000,"_1001_HW_DI120503_PV":0.000000,"_1001_HW_DI120504_PV":0.000000,"_1001_HW_DI120505_PV":0.000000,"_1001_HW_DI120506_PV":0.000000,"_1001_HW_DI120507_PV":0.000000,"_1001_HW_DI120508_PV":0.000000,"_1001_HW_DI120509_PV":0.000000,"_1001_HW_DI120510_PV":0.000000,"_1001_HW_DI120511_PV":0.000000,"_1001_HW_DI120512_PV":0.000000,"_1001_HW_DI120513_PV":0.000000,"_1001_HW_DI120514_PV":0.000000,"_1001_HW_DI120515_PV":0.000000,"_1001_HW_DI120516_PV":0.000000,"_1001_HW_RT070301_PV":41.806629,"_1001_HW_RT070302_PV":30.728159,"_1001_HW_RT070303_PV":37.446587,"_1001_HW_RT070304_PV":617.500000,"_1001_HW_RT070305_PV":38.235462,"_1001_HW_RT070306_PV":617.500000,"_1001_HW_RT070307_PV":617.500000,"_1001_HW_RT070308_PV":617.500000,"_1001_HW_RT080201_PV":37.575916,"_1001_HW_RT080202_PV":37.705231,"_1001_HW_RT080203_PV":38.558811,"_1001_HW_RT080204_PV":37.692287,"_1001_HW_RT080205_PV":32.794170,"_1001_HW_RT080206_PV":617.500000,"_1001_HW_RT080207_PV":617.500000,"_1001_HW_RT080208_PV":617.500000,"_1001_HW_RT090101_PV":617.500000,"_1001_HW_RT090102_PV":617.500000,"_1001_HW_RT090103_PV":617.500000,"_1001_HW_RT090104_PV":617.500000,"_1001_HW_RT090105_PV":617.500000,"_1001_HW_RT090106_PV":617.500000,"_1001_HW_RT090107_PV":617.500000,"_1001_HW_RT090108_PV":21.393274,"_1001_HW_RT100201_PV":16.850239,"_1001_HW_RT100203_PV":28.263302,"_1001_HW_RT100207_PV":118.363228,"_1001_HW_RT100208_PV":30.986353,"_1001_HW_RT100301_PV":69.629410,"_1001_HW_RT100302_PV":198.922623,"_1001_HW_RT100303_PV":21.444784,"_1001_HW_RT100304_PV":22.449375,"_1001_HW_RT100305_PV":26.483381,"_1001_HW_RT100306_PV":22.372091,"_1001_HW_RT100307_PV":22.848717,"_1001_HW_RT100308_PV":22.784311,"_1001_HW_RT100401_PV":34.964619,"_1001_HW_RT100402_PV":36.670815,"_1001_HW_RT100403_PV":39.684258,"_1001_HW_RT100404_PV":42.570431,"_1001_HW_RT100405_PV":50.034515,"_1001_HW_RT100406_PV":50.683060,"_1001_HW_RT100407_PV":51.539288,"_1001_HW_RT100408_PV":30.728159,"_1001_HW_RT100501_PV":49.321236,"_1001_HW_RT100502_PV":46.793304,"_1001_HW_RT100503_PV":47.972828,"_1001_HW_RT100504_PV":23.222336,"_1001_HW_RT100505_PV":24.098541,"_1001_HW_RT100506_PV":36.205418,"_1001_HW_RT100507_PV":28.069777,"_1001_HW_RT100508_PV":31.050894,"_1001_HW_RT110201_PV":30.883068,"_1001_HW_RT110202_PV":30.753973,"_1001_HW_RT110203_PV":31.218731,"_1001_HW_RT110204_PV":26.328642,"_1001_HW_RT110205_PV":25.928921,"_1001_HW_RT110206_PV":30.883068,"_1001_HW_RT110207_PV":73.224396,"_1001_HW_RT110208_PV":20.504822,"_1001_HW_RT120201_PV":31.541525,"_1001_HW_RT120202_PV":31.141277,"_1001_HW_RT120203_PV":59.759583,"_1001_HW_RT120204_PV":49.632477,"_1001_HW_RT120205_PV":80.504448,"_1001_HW_RT120206_PV":73.119759,"_1001_HW_RT120207_PV":77.517204,"_1001_HW_RT120208_PV":75.671143,"_1001_HW_RT120301_PV":77.504120,"_1001_HW_RT120302_PV":77.242203,"_1001_HW_RT120303_PV":84.898422,"_1001_HW_RT120304_PV":32.393776,"_1001_HW_RT120305_PV":33.556263,"_1001_HW_RT120306_PV":53.083485,"_1001_HW_RT120307_PV":41.677189,"_1001_HW_RT120308_PV":617.500000,"_1001_HW_RT120401_PV":49.256416,"_1001_HW_RT120402_PV":50.436611,"_1001_HW_RT120403_PV":53.693520,"_1001_HW_RT120404_PV":61.878757,"_1001_HW_RT120405_PV":60.487522,"_1001_HW_RT120406_PV":61.072567,"_1001_HW_RT120407_PV":73.642967,"_1001_HW_RT120408_PV":617.500000,"_1001_HW_TC070401_PV":387.869080,"_1001_HW_TC070402_PV":254.179199,"_1001_HW_TC070403_PV":353.575409,"_1001_HW_TC070404_PV":307.199402,"_1001_HW_TC070405_PV":415.649994,"_1001_HW_TC070406_PV":415.649994,"_1001_HW_TC070407_PV":253.995499,"_1001_HW_TC070408_PV":254.975159,"_1001_HW_TC080301_PV":254.699692,"_1001_HW_TC080302_PV":252.525330,"_1001_HW_TC080303_PV":253.444260,"_1001_HW_TC080304_PV":252.433395,"_1001_HW_TC080305_PV":164.714584,"_1001_HW_TC080306_PV":166.182419,"_1001_HW_TC080307_PV":156.298096,"_1001_HW_TC080308_PV":128.250519,"_1001_HW_TC090201_PV":1231.250000,"_1001_HW_TC090202_PV":1231.250000,"_1001_HW_TC090203_PV":414.350769,"_1001_HW_TC090204_PV":1000.082825,"_1001_HW_TC090205_PV":1000.082825,"_1001_HW_TC090206_PV":1000.082825,"_1001_HW_TC090207_PV":1000.082825,"_1001_HW_TC090208_PV":1000.082825,"_1001_HW_TC100601_PV":24.031992,"_1001_HW_TC100602_PV":194.823456,"_1001_HW_TC100603_PV":179.698685,"_1001_HW_TC100604_PV":192.667862,"_1001_HW_TC100605_PV":203.118210,"_1001_HW_TC100606_PV":152.442245,"_1001_HW_TC100607_PV":535.039429,"_1001_HW_TC100608_PV":1231.250000,"_1001_HW_TC100701_PV":914.878967,"_1001_HW_TC100702_PV":912.467896,"_1001_HW_TC100703_PV":118.023125,"_1001_HW_TC100704_PV":97.233864,"_1001_HW_TC100708_PV":1231.250000,"_1001_HW_TC110307_PV":991.780945,"_1001_HW_TC110308_PV":1231.250000,"_1001_HW_TC110401_PV":983.006409,"_1001_HW_TC110404_PV":750.689270,"_1001_HW_TC110405_PV":738.990845,"_1001_HW_TC110408_PV":1231.250000,"_1001_HW_TC110501_PV":539.493347,"_1001_HW_TC110502_PV":519.510193,"_1001_HW_TC110503_PV":451.180145,"_1001_HW_TC110504_PV":446.121338,"_1001_HW_TC110505_PV":410.747162,"_1001_HW_TC110506_PV":400.516815,"_1001_HW_TC110507_PV":360.911163,"_1001_HW_TC110508_PV":364.455841,"_1001_HW_TC110601_PV":323.064697,"_1001_HW_TC110602_PV":318.976379,"_1001_HW_TC110603_PV":269.087616,"_1001_HW_TC110604_PV":268.477966,"_1001_HW_TC110605_PV":225.358551,"_1001_HW_TC110606_PV":228.385483,"_1001_HW_TC110607_PV":1231.250000,"_1001_HW_TC110608_PV":1231.250000,"_1001_SH0003_AALM1_PV":0.000000,"_1001_SH0010_PRO1_IN":-0.500000,"_1001_SH0010_PRO2_IN":9.479815,"_1001_SH0010_PRO3_IN":-0.625000,"_1001_SH0010_PRO4_IN":6.070231,"_1001_SH0010_PRO5_IN":0.023611,"_1001_SH0010_PRO6_IN":-0.006019,"_1001_SH0010_PRO7_IN":199.134247,"_1001_SH0010_PRO8_IN":199.254639,"_1001_SH0012_AALM2_PV":8203.656250,"_1001_SH0013_AALM2_PV":7217.221680,"_1001_SH0014_AALM1_PV":3.019795,"_1001_SH0015_AALM1_PV":1.650945,"_1001_SH0016_AALM1_PV":8808.333984,"_1001_SH0016_AALM2_PV":88150.007813,"_1001_SH0016_AALM3_PV":4945.833496,"_1001_SH0016_AALM4_PV":6539.750000,"_1001_SH0016_AALM5_PV":38745.000000,"_1001_SH0016_AALM6_PV":14045.839844,"_1001_SH0016_AALM7_PV":11685.000000,""_1001_HW_DI060504_PV":0.000000,":19065.000000,"_1001_SH0017_AALM1_PV":-11.572223,"_1001_SH0017_AALM2_PV":8.162498,"_1001_SH0017_AALM3_PV":1.954865,"_1001_SH0018_AALM1_PV":77113.335938,"_1001_SH0022_AALM3_ACFG":0.000000,"_1001_SH0022_AALM3_ACK":0.000000,"_1001_SH0022_AALM3_ALM":0.000000,"_1001_SH0022_AALM3_EN":0.000000,"_1001_SH0022_AALM3_HAL":10000.000000,"_1001_SH0022_AALM3_HHAL":20000.000000,"_1001_SH0022_AALM3_LAL":-10000.000000,"ts":1675925762171} #####
}

func Test_adjustTSData1(t *testing.T) {
	tests := []struct {
		name           string
		args           string
		wantDataAdjust string
	}{
		{
			"1", `{"ts":111,"field_bool":1}`, "",
		},
	}
	bytes := []byte(entityStr)
	entity, err := NewEntity("foo", bytes)
	assert.Nil(t, err)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.wantDataAdjust, string(adjustTSData([]byte(tt.args), entity)), "adjustTSData(%v, %v)", tt.args, entity)
		})
	}
}
