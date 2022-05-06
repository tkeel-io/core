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
	"github.com/stretchr/testify/assert"
	"github.com/tkeel-io/core/pkg/placement"
	"github.com/tkeel-io/core/pkg/repository"
	"github.com/tkeel-io/core/pkg/util/json"
	"github.com/tkeel-io/core/pkg/util/path"
	"testing"

	"github.com/tkeel-io/tdtl"
)

// import (
// 	"context"
// 	"reflect"
// 	"testing"
// 	"time"

// 	"github.com/stretchr/testify/assert"
// 	proto "github.com/tkeel-io/core/api/core/v1"
// 	"github.com/tkeel-io/core/pkg/util"
// )

// func TestRuntime_HandleEvent(t *testing.T) {
// 	ctx := context.Background()
// 	ev := &proto.ProtoEvent{
// 		Id:        util.UUID("ev"),
// 		RawData:   []byte(`{}`),
// 		Timestamp: time.Now().UnixNano(),
// 		Metadata:  map[string]string{},
// 	}

// 	cc := NewRuntime(ctx, "core-1")
// 	ret, err := cc.HandleEvent(ctx, ev)
// 	assert.NoError(t, err, "err is %v", err)
// 	byt, err := cc.entities["entity1"].Raw()
// 	t.Log(cc, string(byt))
// 	t.Log(string(ret.State))
// }

// func TestRuntime_OpEntity_HandleEvent(t *testing.T) {
// 	type args struct {
// 	}
// 	tests := []struct {
// 		name       string
// 		typ        EntityEventType
// 		path       string
// 		eventValue []byte
// 		wantErr    bool
// 		want       string
// 	}{
// 		{"1", OpEntityPropsUpdata, "a.b.c", []byte(`"abc"`), false, `{"ID":"","Type":"","Owner":"","Source":"","Version":0,"LastTime":0,"TemplateID":"","Property":{"a":{"b":{"c":"abc"}}},"Scheme":{}}`},
// 		//{"2", OpEntityPropsPatch, "a.b.c", []byte(`"abc"`), false, `{"Property":{"a":{"b":{"c":"abc"}}},"Scheme":{}}`},
// 		{"3", OpEntityPropsGet, "a.b.c", []byte(`"abc"`), false, `{"ID":"","Type":"","Owner":"","Source":"","Version":0,"LastTime":0,"TemplateID":"","Property":{},"Scheme":{}}`},
// 		{"4", OpEntityConfigsUpdata, "a.b.c", []byte(`"abc"`), false, `{"ID":"","Type":"","Owner":"","Source":"","Version":0,"LastTime":0,"TemplateID":"","Property":{},"Scheme":{"a":{"b":{"c":"abc"}}}}`},
// 		//{"5", OpEntityConfigsPatch, "a.b.c", []byte(`"abc"`), false, `{"Property":{},"Scheme":{"a":{"b":{"c":"abc"}}}}`},
// 		{"6", OpEntityConfigsGet, "a.b.c", []byte(`{
// 			Id:     "device123",
// 			Type:   "DEVICE",
// 			Owner:  "tomas",
// 			Source: "CORE-SDK",
// 			Properties: map[string]interface{}{
// 			"temp": 20,
// 		},`), false, `{"ID":"","Type":"","Owner":"","Source":"","Version":0,"LastTime":0,"TemplateID":"","Property":{},"Scheme":{}}`},
// 	}
// 	ctx := context.Background()
// 	for _, tt := range tests {
// 		cc := newRuntime()
// 		t.Run(tt.name, func(t *testing.T) {
// 			ev := &proto.ProtoEvent{}
// 			got, err := cc.HandleEvent(ctx, ev)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("HandleEvent() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if !reflect.DeepEqual(string(got.State), tt.want) {
// 				t.Errorf("HandleEvent() \ngot = %v, \nwant  %v", string(got.State), tt.want)
// 			}
// 		})
// 	}
// }

// func newRuntime() *Runtime {
// 	cc := NewRuntime(context.Background(), "core-1")
// 	return cc
// }

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
}

var expr1 = `{"ID":"/core/v1/expressions/usr-57bea3a2d74e21ebbedde8268610/iotd-b10bcaa1-ba98-4e03-bece-6f852feb6edf/properties.telemetry.yc1",
"Path":"properties.telemetry.yc1",
"Name":"","Type":"eval","Owner":"usr-57bea3a2d74e21ebbedde8268610",
"EntityID":"iotd-b10bcaa1-ba98-4e03-bece-6f852feb6edf",
"Expression":"iotd-06a96c8d-c166-447c-afd1-63010636b362.properties.telemetry.yc1",
"Description":"iotd-06a96c8d-c166-447c-afd1-63010636b362=映射3,yc1=遥测1"}`
var expr2 = `{"ID":"/core/v1/expressions/usr-57bea3a2d74e21ebbedde8268610/iotd-b10bcaa1-ba98-4e03-bece-6f852feb6edf/properties.telemetry.yc2",
"Path":"properties.telemetry.yc2",
"Name":"","Type":"eval","Owner":"usr-57bea3a2d74e21ebbedde8268610",
"EntityID":"iotd-b10bcaa1-ba98-4e03-bece-6f852feb6edf",
"Expression":"iotd-06a96c8d-c166-447c-afd1-63010636b362.properties.telemetry.yc1",
"Description":"iotd-06a96c8d-c166-447c-afd1-63010636b362=映射3,yc1=遥测1"}`

//iotd-b10bcaa1-ba98-4e03-bece-6f852feb6edf
//- yc2 = iotd-06a96c8d-c166-447c-afd1-63010636b362.properties.telemetry.yc1
//- yc1 = iotd-06a96c8d-c166-447c-afd1-63010636b362.properties.telemetry.yc1
func TestRuntime_handleComputed(t *testing.T) {
	placement.Initialize()
	placement.Global().Append(placement.Info{
		ID:   "core/1234",
		Flag: true,
	})
	rt := Runtime{
		expressions: map[string]ExpressionInfo{},
		subTree:     path.NewRefTree(),
		evalTree:    path.New(),
	}
	var expr = repository.Expression{}
	expr.Decode([]byte(""), []byte(expr1))
	exprInfo1, err := parseExpression(expr, 1)
	assert.Nil(t, err)
	for runtimeID, exprIns := range exprInfo1 {
		t.Log(runtimeID)
		rt.AppendExpression(*exprIns)
	}

	expr = repository.Expression{}
	expr.Decode([]byte(""), []byte(expr2))
	exprInfo2, err := parseExpression(expr, 1)
	assert.Nil(t, err)
	for runtimeID, exprIns := range exprInfo2 {
		t.Log(runtimeID)
		rt.AppendExpression(*exprIns)
	}
	feed := Feed{
		TTL:      0,
		Err:      nil,
		Event:    nil,
		State:    nil,
		EntityID: "iotd-06a96c8d-c166-447c-afd1-63010636b362",
		Patches:  nil,
		Changes: []Patch{
			{
				Op:    json.OpReplace,
				Path:  "properties.telemetry.yc1",
				Value: tdtl.New(123),
			},
		},
	}
	got := rt.handleComputed(context.Background(), &feed)
	t.Log(got)
}
