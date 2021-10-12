package mapper

type mapper struct {
	id string
}

func NewMapper() *mapper {
	return &mapper{}
}

// Id returns mapper id.
func (m *mapper) Id() string {
	return m.id
}

// String returns MQL text.
func (m *mapper) String() string {
	panic("implement me.")
}

// TargetEntity returns target entity.
func (m *mapper) TargetEntity() string {
	panic("implement me.")
}

// SourceEntities returns source entities.
func (m *mapper) SourceEntities() []string {
	panic("implement me.")
}

// Tentacles returns tentacles.
func (m *mapper) Tentacles() []Tentacler {
	panic("implement me.")
}

// Copy duplicate a mapper.
func (m *mapper) Copy() Mapper {
	panic("implement me.")
}

// Exec excute input returns output.
func (m *mapper) Exec(map[string]map[string]interface{}) (map[string]map[string]interface{}, error) {
	panic("implement me.")
}
