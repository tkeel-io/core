package mapper

import (
	"github.com/tkeel-io/core/pkg/mql"
)

type mapper struct {
	id      string
	mqlText string
}

func NewMapper(id, mqlText string) *mapper {
	return &mapper{id: id, mqlText: mqlText}
}

// Id returns mapper id.
func (m *mapper) Id() string {
	return m.id
}

// String returns MQL text.
func (m *mapper) String() string {
	return m.mqlText
}

// TargetEntity returns target entity.
func (m *mapper) TargetEntity() string {
	return mql.NewMQL(m.mqlText).Target()
}

// SourceEntities returns source entities(include target entity).
func (m *mapper) SourceEntities() []string {
	mqlInst := mql.NewMQL(m.mqlText)
	sourceEntities := mqlInst.Entities()
	return append(sourceEntities, mqlInst.Target())
}

// Tentacles returns tentacles.
func (m *mapper) Tentacles() []Tentacler {

	tentacles := make([]Tentacler, 0)
	mqlInst := mql.NewMQL(m.mqlText)
	tts := mqlInst.Tentacles()

	mItems := make([]string, 0)
	for entityId, items := range tts {

		eItems := make([]string, 0)
		for _, item := range items {
			tentacleKey := GenTentacleKey(entityId, item)
			mItems = append(mItems, tentacleKey)
			eItems = append(mItems, tentacleKey)
		}

		tentacles = append(tentacles, NewTentacle(TentacleTypeEntity, entityId, eItems))
	}

	tentacles = append(tentacles, NewTentacle(TentacleTypeMapper, m.id, mItems))

	return tentacles
}

// Copy duplicate a mapper.
func (m *mapper) Copy() Mapper {
	return NewMapper(m.id, m.mqlText)
}

// Exec excute input returns output.
func (m *mapper) Exec(values map[string]map[string]interface{}) (res map[string]map[string]interface{}, err error) {
	return mql.NewMQL(m.mqlText).Exec(values)
}
