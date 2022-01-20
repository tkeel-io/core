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
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
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
		entity1.property1 + entity.property3 as property3,
		pp1+pp2 as pp3`

	t.Log("parse tql: ", tql)
	l, _ := Parse(tql)
	cfg, _ := l.GetParseConfigs()
	t.Log("parse tql, result: ", cfg)

	in := make(map[string][]byte)
	in["entity1.property1"] = []byte("1")
	in["entity.property2.name"] = []byte("2")
	in["entity.property3"] = []byte("3")
	in["pp1"] = []byte("'device'")
	in["pp2"] = []byte("'123'")
	t.Log("in: ", in)
	out := l.GetComputeResults(in)
	t.Log(" out: ", out)
	t.Log(" out: ", string(out["pp3"]))

	// out2
	in["entity1.property1"] = []byte("4")
	in["entity.property2.name"] = []byte("7")
	in["entity.property3"] = []byte("6")
	t.Log("in: ", in)
	out2 := l.GetComputeResults(in)
	t.Log(" out2: ", out2)
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

func TestComputingString(t *testing.T) {
	tqlS := `insert into device123 select field1 + field2 as field3`
	l, err := Parse(tqlS)
	assert.Equal(t, nil, err)

	in := make(map[string][]byte)
	in["field1"] = []byte("4")
	in["field2"] = []byte("7")
	out1 := l.GetComputeResults(in)
	t.Log("out1: ", string(out1["field3"]))

	in["field1"] = []byte("'111'")
	in["field2"] = []byte("'xxx'")
	out2 := l.GetComputeResults(in)
	t.Log("out2: ", string(out2["field3"]))
}
