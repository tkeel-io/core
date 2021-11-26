package mapper

import (
	"testing"

	"github.com/tkeel-io/core/pkg/constraint"
)

func TestMapper(t *testing.T) {
	input := map[string]constraint.Node{
		"entity1.property1":      constraint.NewNode("123"),
		"entity2.property2.name": constraint.NewNode("123"),
		"entity2.property3":      constraint.NewNode("123"),
	}

	tqlTexts := []struct {
		id       string
		tqlText  string
		input    map[string]constraint.Node
		computed bool
	}{
		{"tql1", "insert into device1 select *", map[string]constraint.Node{}, false},
		{"tql2", "insert into test123 select test234.temp as temp", map[string]constraint.Node{"test234.temp": constraint.NewNode(`123`)}, true},
		{"tql3", `insert into entity3 select entity1.property1 as property1, entity2.property2.name as property2, entity1.property1 + entity2.property3 as property3`, input, true},
		{"tql4", "insert into sub123 select test123.temp", nil, false},
	}

	for _, tqlInst := range tqlTexts {
		t.Run(tqlInst.id, func(t *testing.T) {
			m, _ := NewMapper(tqlInst.id, tqlInst.tqlText)

			t.Log("parse ID: ", m.ID())

			tentacles := m.Tentacles()
			t.Logf("parse tentacles, count %d.", len(tentacles))
			for index, tentacle := range tentacles {
				t.Logf("tentacle.%d, type: %s, target: %s, items: %s.",
					index, tentacle.Type(), tentacle.TargetID(), tentacle.Items())
			}

			t.Log("parse target entity: ", m.TargetEntity())
			t.Log("parse source entities: ", m.SourceEntities())

			m.Copy()

			if tqlInst.computed {
				out, err := m.Exec(tqlInst.input)
				t.Logf("exec input: %v\n output: %v\n error: %v", tqlInst.input, out, err)
			}
		})
	}
}
