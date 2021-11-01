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
