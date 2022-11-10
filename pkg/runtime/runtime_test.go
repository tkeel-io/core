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
	"time"

	"github.com/stretchr/testify/assert"
	v1 "github.com/tkeel-io/core/api/core/v1"
	"github.com/tkeel-io/core/pkg/placement"
	"github.com/tkeel-io/core/pkg/repository"
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
"telemetry": {"src1": 123, "src2": 123}
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
        "value": 9.79
      },
      "GGGG": {
        "ts": 1666777942241,
        "value": "8.64"
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

	rt := &Runtime{
		dispatcher: &dispatcherMock{},
		entities: map[string]Entity{
			entity.ID(): entity,
		},
		enCache: NewCacheMock(map[string]Entity{
			"iotd-06a96c8d-c166-447c-afd1-63010636b362": en,
		}),
		expressions:     map[string]ExpressionInfo{},
		subTree:         path.NewRefTree(),
		evalTree:        path.New(),
		entityResourcer: EntityResource{},
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
						Value:    []byte(`{"id":"iotd-aeb86fb2-e694-431b-b98a-8447c6745817","ts":1666790831700,"values":"ewoJIkZGRkYiOiAiMS4zOSIsCgkiR0dHRyI6ICI5LjY0IiwKCSJmaXJlX2NvbmZpZGVuY2UiOiAiOS40NiIsCgkic21va2VfY29uZmlkZW5jZSI6ICI0LjY0IiwKCSJzd2l0Y2giOiAiMS44MiIsCgkidGVtcGVyYXR1cmUiOiAiNi40NCIsCgkidmlkZW9fdGltZV9vZmZzZXQiOiAiOS44NyIKfQ==","path":"iotd-aeb86fb2-e694-431b-b98a-8447c6745817/v1/devices/me/telemetry","type":"telemetry","mark":"upstream"}`),
					},
				},
			},
		},
	}

	execer, feed := rt.handleEntityEvent(ev)
	newFeed := execer.Exec(context.Background(), feed)

	t.Log(newFeed)
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
