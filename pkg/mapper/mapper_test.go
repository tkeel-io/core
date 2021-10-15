package mapper

import "testing"

func TestMapper(t *testing.T) {
	m := newMapper("id", "sql")

	m.ID()
	m.Copy()
	m.SourceEntities()
	m.TargetEntity()
	m.Tentacles()
}
