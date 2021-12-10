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
	"testing"
)

func TestParserAndComputing(t *testing.T) {
	tql := `insert into entity3 select
		entity1.property1 as property1,
		entity.property2.name as property2,
		entity1.property1 + entity.property3 as property3`

	log.Info("parse tql: ", tql)
	l := Parse(tql)
	cfg := l.GetParseConfigs()
	log.Info("========\n ", cfg)

	in := make(map[string][]byte)
	in["entity1.property1"] = []byte("1")
	in["entity.property2.name"] = []byte("2")
	in["entity.property3"] = []byte("3")
	log.Info("========\n in: ", in)
	out := l.GetComputeResults(in)
	log.Info(" out: ", out)

	// out2
	in["entity1.property1"] = []byte("4")
	in["entity.property2.name"] = []byte("7")
	in["entity.property3"] = []byte("6")
	log.Info("========\n in: ", in)
	out2 := l.GetComputeResults(in)
	log.Info(" out2: ", out2)
}

func TestParser(t *testing.T) {
	tql := `insert into target_entity select *`

	log.Info("parse tql: ", tql)
	l := Parse(tql)
	cfg := l.GetParseConfigs()
	log.Info("========\n ", cfg)
}
