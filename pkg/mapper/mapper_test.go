package mapper

import "testing"

func TestMapper(t *testing.T) {
	tqlTexts := []struct {
		id      string
		tqlText string
	}{
		{"tql1", "insert into device1 select *"},
		{"tql2", `insert into entity3 select entity1.property1 as property1, entity2.property2.name as property2, entity1.property1 + entity2.property3 as property3`},
	}

	for _, tqlInst := range tqlTexts {
		t.Run(tqlInst.id, func(t *testing.T) {
			m := NewMapper(tqlInst.id, tqlInst.tqlText)

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
		})
	}
}
