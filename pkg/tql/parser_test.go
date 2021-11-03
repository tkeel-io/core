package tql

import (
	"fmt"
	"testing"
)


func TestParserAndComputing(t *testing.T) {
	tql := `insert into entity3 select
		entity1.property1 as property1,
		entity.property2.name as property2,
		entity1.property1 + entity.property3 as property3`

	fmt.Println("parse tql: ", tql)
	l := Parse(tql)
	cfg := l.GetParseConfigs()
	fmt.Println("========\n ", cfg)

	in := make(map[string]interface{})
	in["entity1.property1"] = 1
	in["entity.property2.name"] = 2
	in["entity.property3"] = 3
	fmt.Println("========\n in: ", in)
	out := l.GetComputeResults(in)
	fmt.Println(" out: ", out)

	// out2
	in["entity1.property1"] = 4
	in["entity.property2.name"] = 7
	in["entity.property3"] = 6
	fmt.Println("========\n in: ", in)
	out2 := l.GetComputeResults(in)
	fmt.Println(" out2: ", out2)
	
}

func TestParser(t *testing.T) {

	tql := `insert into target_entity select *`

	fmt.Println("parse tql: ", tql)
	l := Parse(tql)
	cfg := l.GetParseConfigs()
	fmt.Println("========\n ", cfg)
}