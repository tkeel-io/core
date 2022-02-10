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

package tql

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/dop251/goja"
	"github.com/stretchr/testify/assert"
	"github.com/tkeel-io/collectjs"
	"github.com/tkeel-io/core/pkg/constraint"
	"github.com/tkeel-io/kit/log"
)

func TestMain(m *testing.M) {
	log.InitLogger("core", "debug", true)
	os.Exit(m.Run())
}

func TestParserAndComputing(t *testing.T) {
	tql := `insert into entity3 select
		entity1.property1 as property1,
		entity.property2.name as property2,
		entity1.property1 + entity.property3 as property3`

	t.Log("parse tql: ", tql)
	l, _ := Parse(tql)
	cfg, _ := l.GetParseConfigs()
	t.Log("parse tql, result: ", cfg)

	in := make(map[string][]byte)
	in["entity1.property1"] = []byte("1")
	in["entity.property2.name"] = []byte("2")
	in["entity.property3"] = []byte("3")
	t.Log("in: ", in)
	out := l.GetComputeResults(in)
	t.Log(" out: ", out)

	// out2
	in["entity1.property1"] = []byte(`'233'`)
	in["entity.property2.name"] = []byte(`'2222'`)
	in["entity.property3"] = []byte(`'test'`)
	t.Log("in: ", in)
	out2 := l.GetComputeResults(in)
	for key, val := range out2 {
		t.Log(key, ":", val)
	}
}

func TestParser(t *testing.T) {
	tqls := map[string]string{
		"tql1": "insert into device123 select device234.*",
		"tql2": "insert into device123 select device234.temp as temp",
		"tql3": "insert into sub123 select *",
		"tql4": "insert into sub123 select *.temp",
	}

	for name, tqlString := range tqls {
		l, err := Parse(tqlString)
		assert.Equal(t, nil, err)
		_, err = l.GetParseConfigs()
		assert.Equal(t, nil, err)
		t.Logf("TQL name: %s, TQL: %s", name, tqlString)
	}
}

func TestGetParseConfigs(t *testing.T) {
	tqlString := "insert into device123 select device234.*"

	l, err := Parse(tqlString)
	assert.Equal(t, nil, err)
	cfg, err := l.GetParseConfigs()
	assert.Equal(t, nil, err)

	expectCfg := TQLConfig{
		TargetEntity:   "device123",
		SourceEntities: []string{"device234"},
		Tentacles:      []TentacleConfig{{SourceEntity: "device234", PropertyKeys: []string{"*"}}},
	}
	assert.Equal(t, expectCfg, cfg)
}

func TestJson(t *testing.T) {
	bytes, _ := collectjs.Set([]byte(`{}`), "test", []byte(`'123'`))
	t.Log(string(bytes))

	v := make(map[string]interface{})
	t.Log(json.Unmarshal(bytes, &v))
	t.Log(v)
}

func TestExec(t *testing.T) {
	tqlString := `insert into entity3 select
	entity1.property1 as property1,
	entity.property2.name as property2,
	entity1.property1 + '/' + '123' as property3`

	tqlInstance, err := NewTQL(tqlString)

	t.Log(err)

	result, err := tqlInstance.Exec(map[string]constraint.Node{
		"entity1.property1":     constraint.NewNode("test"),
		"entity.property2.name": constraint.NewNode("123"),
		"entity.property3":      constraint.NewNode("g123"),
	})

	t.Log(err)
	t.Log(result)

	tqlString = `insert into 7ffed0dc-3ed5-4137-9c16-a2c9c74e0bf6 select f8f0327b-51e4-400a-a2e1-c95e371ec99d.path  + '/' + '7ffed0dc-3ed5-4137-9c16-a2c9c74e0bf6' as path`

	tqlInstance, err = NewTQL(tqlString)

	t.Log(err)

	t.Log("target: ", tqlInstance.Target())
	t.Log("sources: ", tqlInstance.Entities())
	t.Log("tentacles: ", tqlInstance.Tentacles())

	result, err = tqlInstance.Exec(map[string]constraint.Node{
		"f8f0327b-51e4-400a-a2e1-c95e371ec99d.path": constraint.NewNode("test"),
		"entity.property2.name":                     constraint.NewNode("123"),
		"entity.property3":                          constraint.NewNode("g123"),
	})

	t.Log(err)
	t.Log(result)
}

func TestGoja(t *testing.T) {
	vm := goja.New()
	v, err := vm.RunString(`2+222.3`)
	if err != nil {
		panic(err)
	}
	t.Log(v)
	t.Log(reflect.ValueOf(v).Kind().String())
}

func BenchmarkGoja(b *testing.B) {
	exprs := []struct {
		Name       string
		Expression string
	}{
		{
			Name:       "computting-integer-add",
			Expression: "1223+324",
		},
		{
			Name:       "computting-integer-sub",
			Expression: "1223 - 324",
		},
		{
			Name:       "computting-integer-mul",
			Expression: "1223 * 324",
		},
		{
			Name:       "computting-integer-div",
			Expression: "1223/324",
		},
		{
			Name:       "computting-float-add",
			Expression: "1223.344+324.7",
		},
		{
			Name:       "computting-float-sub",
			Expression: "1223.344 - 324.7",
		},
		{
			Name:       "computting-float-mul",
			Expression: "1223.344*324.7",
		},
		{
			Name:       "computting-float-div",
			Expression: "1223.344/324.7",
		},
		{
			Name:       "computting-string",
			Expression: "'too yong too simple' + 'He knows most who speaks least'",
		},
	}

	vm := goja.New()
	b.ResetTimer()
	for _, expr := range exprs {
		b.Log("bench for ", expr.Name)
		b.Run(expr.Name, func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				vm.RunString(expr.Expression)
			}
		})
	}
}
