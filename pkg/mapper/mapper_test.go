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

package mapper

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tkeel-io/tdtl"
)

func TestMapper1(t *testing.T) {
	input := map[string]tdtl.Node{
		"entity1.property1":      tdtl.IntNode(123),
		"entity2.property2.name": tdtl.StringNode("tom"),
		"entity2.property3":      tdtl.IntNode(123),
	}

	tqlTexts := []struct {
		id       string
		tqlText  string
		input    map[string]tdtl.Node
		computed bool
		target   string
		sources  []string
		output   map[string]tdtl.Node
	}{
		{"tql1", "insert into device1 select device3.*", map[string]tdtl.Node{"device2.abc": tdtl.StringNode("abcs")}, true, "device1", []string{"device3"}, map[string]tdtl.Node{}},
		{"tql2", "insert into test123 select test234.temp as temp", map[string]tdtl.Node{"test234.temp": tdtl.IntNode(123)}, true, "test123", []string{"test234"}, map[string]tdtl.Node{"temp": tdtl.IntNode(123)}},
		{"tql3", "insert into test123 select test234.temp", map[string]tdtl.Node{"test234.temp": tdtl.IntNode(123)}, true, "test123", []string{"test234"}, map[string]tdtl.Node{}},
		{"tql4", `insert into entity3 select entity1.property1 as property1, entity2.property2.name as property2, entity1.property1 + entity2.property3 as property3`, input, true, "entity3", []string{"entity1", "entity2"}, map[string]tdtl.Node{"property1": tdtl.IntNode(123), "property2": tdtl.StringNode("tom"), "property3": tdtl.IntNode(246)}},
		{"tql5", "insert into sub123 select test123.temp", nil, false, "sub123", []string{"test123"}, map[string]tdtl.Node{}},
	}

	for _, tqlInst := range tqlTexts {
		t.Run(tqlInst.id, func(t *testing.T) {
			m, err := NewMapper(Mapper{ID: tqlInst.id, TQL: tqlInst.tqlText}, 0)
			if nil != err {
				t.Log("error: ", err)
				return
			}

			t.Log("parse ID: ", m.ID())

			tentacles := m.Tentacles()
			t.Logf("parse tentacles, count %d.", len(tentacles))
			for _, tens := range tentacles {
				for index, tentacle := range tens {
					t.Logf("tentacle.%d, type: %s, target: %s, items: %s.",
						index, tentacle.Type(), tentacle.TargetID(), tentacle.Items())
				}
			}

			t.Log("parse target entity: ", m.TargetEntity())
			t.Log("parse source entities: ", m.SourceEntities())
			sources := make([]string, 0, len(m.SourceEntities()))
			for k := range m.SourceEntities() {
				sources = append(sources, k)
			}
			sort.Strings(sources)
			assert.Equal(t, tqlInst.target, m.TargetEntity())
			assert.Equal(t, tqlInst.sources, sources)
			m.Copy()

			if tqlInst.computed {
				out, err := m.Exec(tqlInst.input)
				t.Logf("exec input: %v\n output: %v\n error: %v", tqlInst.input, out, err)
				assert.Equal(t, tqlInst.output, out)
			}
		})
	}
}

func TestMapper2(t *testing.T) {
	tqlText := "insert into x4c1e33a1-6899-4643-a6b3-46cf37950b7f select x54cf69fc-78c3-4f79-9f6b-5d5e5bd8d3c0.sysField._spacePath  + '/x4c1e33a1-6899-4643-a6b3-46cf37950b7f' as sysField._spacePath"
	mapperIns, err := NewMapper(Mapper{ID: "mapper123", TQL: tqlText}, 0)
	assert.Nil(t, err)
	t.Log("id: ", mapperIns.ID())
	t.Log("target: ", mapperIns.TargetEntity())
	t.Log("sources: ", mapperIns.SourceEntities())
	for _, tentacle := range mapperIns.Tentacles() {
		t.Log("tentacle: ", tentacle)
	}

	res, err := mapperIns.Exec(map[string]tdtl.Node{
		"x54cf69fc-78c3-4f79-9f6b-5d5e5bd8d3c0.sysField._spacePath": tdtl.IntNode(123),
	})
	assert.Nil(t, err)
	t.Log("result: ", res)
	assert.Equal(t, map[string]tdtl.Node{"sysField._spacePath": tdtl.StringNode("123/x4c1e33a1-6899-4643-a6b3-46cf37950b7f")}, res)
	res, err = mapperIns.Exec(map[string]tdtl.Node{
		"x54cf69fc-78c3-4f79-9f6b-5d5e5bd8d3c0.sysField._spacePath": tdtl.StringNode("abc"),
	})
	assert.Nil(t, err)
	t.Log("result: ", res)
	assert.Equal(t, map[string]tdtl.Node{"sysField._spacePath": tdtl.StringNode("abc/x4c1e33a1-6899-4643-a6b3-46cf37950b7f")}, res)
}

func TestMapper3(t *testing.T) {
	tqlText := `insert into b3a22c80-6afe-44a0-91b7-f1e49f3c962e select x49ff9ece-bc90-4e2c-b02e-b96ddedb8e2d.sysField._spacePath  + '/b3a22c80-6afe-44a0-91b7-f1e49f3c962e' as sysField._spacePath, aaa.p1, b3a22c80-6afe-44a0-91b7-f1e49f3c962e.temp * 2 as temp2 `

	mapperIns, err := NewMapper(Mapper{ID: "mapper123", TQL: tqlText}, 0)
	assert.Nil(t, err)
	t.Log("id: ", mapperIns.ID())
	t.Log("target: ", mapperIns.TargetEntity())
	t.Log("sources: ", mapperIns.SourceEntities())
	for _, tentacle := range mapperIns.Tentacles() {
		t.Log("tentacle: ", tentacle)
	}

	tentacles := mapperIns.Tentacles()
	t.Log("tentacles: ", tentacles)

	res, err := mapperIns.Exec(map[string]tdtl.Node{
		"x49ff9ece-bc90-4e2c-b02e-b96ddedb8e2d.sysField._spacePath": tdtl.New(`"tom"`),
		"aaa.p1": tdtl.StringNode("p1"),
		"b3a22c80-6afe-44a0-91b7-f1e49f3c962e.temp": tdtl.IntNode(5),
	})

	assert.Nil(t, err)
	t.Log("result: ", res)
	assert.Equal(t, map[string]tdtl.Node{"sysField._spacePath": tdtl.StringNode("tom/b3a22c80-6afe-44a0-91b7-f1e49f3c962e"), "temp2": tdtl.IntNode(10)}, res)
}
