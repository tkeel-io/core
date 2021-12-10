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

package main

import (
	"fmt"
	 TQL "github.com/tkeel-io/core/pkg/tql"
)

func main() {
	tql := `insert into entity3 select
		entity1.property1 as property1,
		entity2.property2.name as property2,
		entity1.property1 + entity2.property3 as property3`

	//tql := `insert into target_entity select *`

	//tql := `insert into target_entity select src_entity`

	fmt.Println("parse tql: ", tql)
	l := TQL.Parse(tql)
	cfg := l.GetParseConfigs()
	fmt.Println("========\n ", cfg)

	in := make(map[string]interface{})
	in["entity1.property1"] = 1
	in["entity2.property2.name"] = 2
	in["entity2.property3"] = 3
	fmt.Println("========\n in: ", in)
	out := l.GetComputeResults(in)
	fmt.Println(" out: ", out)

	// out2
	in["entity1.property1"] = 4
	in["entity2.property2.name"] = 7
	in["entity2.property3"] = 6
	fmt.Println("========\n in: ", in)
	out2 := l.GetComputeResults(in)
	fmt.Println(" out2: ", out2)
}
