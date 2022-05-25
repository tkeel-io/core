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
	"github.com/tkeel-io/collectjs"
	pb "github.com/tkeel-io/core/api/core/v1"
	v1 "github.com/tkeel-io/core/api/core/v1"
	"github.com/tkeel-io/core/pkg/config"
	"github.com/tkeel-io/core/pkg/placement"
	"github.com/tkeel-io/core/pkg/repository"
	"github.com/tkeel-io/core/pkg/repository/dao"
	_ "github.com/tkeel-io/core/pkg/resource/store/memory"
	"github.com/tkeel-io/core/pkg/util/json"
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
	out := adjustTSData(in)
	t.Log(string(out))
	in = []byte(`{"ts":1646954803319,"values":{"humidity5":83.0,"temperature5":43.6}}`)
	out = adjustTSData(in)
	t.Log(string(out))

	in = []byte(`{"ModBus-TCP":{"ts":1649215733364,"values":{"wet":42,"temperature":"abc"}},"OPC-UA":{"ts":1649215733364,"values":{"counter":15}}}`)
	out = adjustTSData(in)
	t.Log(string(out))
	tt := collectjs.ByteNew(out)
	item := tt.Get("OPC-UA@counter")
	t.Log(string(item.GetRaw()))
	item = tt.Get("OPC-UA")
	t.Log(string(item.GetRaw()))
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
				Op:    json.OpReplace,
				Path:  "properties.telemetry.src1",
				Value: tdtl.New(123),
			},
			{
				Op:    json.OpReplace,
				Path:  "properties.telemetry.src2",
				Value: tdtl.New(123),
			},
		},
	}
	got := rt.handleTentacle(context.Background(), &feed)
	t.Log(string(feed.State))
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

func NewRepo() repository.IRepository {
	daoIns, _ := dao.NewMock(context.Background(), config.Metadata{}, config.EtcdConfig{})
	return repository.New(daoIns)
}

func TestRuntime_HandleEvent(t *testing.T) {
	placement.Initialize()
	placement.Global().Append(placement.Info{
		ID:   "core/1234",
		Flag: true,
	})
	entityID := "ev-12345"

	er := EntityResource{
		FlushHandler:  nil,
		RemoveHandler: nil,
	}
	repo := NewRepo()

	ctx := context.Background()
	repo.PutEntity(ctx, entityID, []byte(`{"id":"ev-12345","type":"template","owner":"usr-6e3f3707346822583797131e283f","source":"device","version":2,"last_time":1652164204493,"mappers":null,"template_id":"","properties":{"basicInfo":{"description":"这是一个测试摸版","name":"test-a06239"}}}`))
	enB, err := repo.GetEntity(ctx, entityID)
	t.Log(err)
	t.Log(string(enB))

	r := NewRuntime(ctx, er, "core", nil, repo)
	ev := &pb.ProtoEvent{
		Id:        entityID,
		Timestamp: time.Now().UnixNano(),
		Metadata:  map[string]string{},
		Data: &pb.ProtoEvent_Patches{
			Patches: &pb.PatchDatas{
				Patches: []*pb.PatchData{
					{
						Path:     "scheme.attributes.define.fields.serial-N",
						Operator: "replace",
						Value:    []byte(`{"id":"serial-N","type":"string","name":"序列号N","weight":0,"enabled":false,"enabled_search":false,"enabled_time_series":false,"description":"设备批次","define":{"default_value":"xxxxxxxn","rw":"w"},"last_time":1652164204383}`),
					},
					{
						Path:     "scheme.attributes.define.fields.serial-3",
						Operator: "replace",
						Value:    []byte(`{"id":"serial-3","type":"string","name":"序列号3","weight":0,"enabled":false,"enabled_search":false,"enabled_time_series":false,"description":"设备批次","define":{"default_value":"xxxxxxx3","rw":"r"},"last_time":1652164204383}`),
					},
				},
			},
		},
	}
	t.Log(ev.Entity())
	ev.SetEntity(entityID)
	ev.SetType(pb.ETEntity)
	ev.SetAttr(pb.MetaPathConstructor, string(pb.PCScheme))
	err = r.HandleEvent(ctx, ev)
	t.Log(err)
}
