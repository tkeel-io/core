package mapper

import (
	"github.com/pkg/errors"
	"github.com/tkeel-io/core/pkg/tql"
)

type mapper struct {
	id      string
	tqlText string
	tqlInst tql.TQL
}

func NewMapper(id, tqlText string) Mapper {
	return &mapper{
		id:      id,
		tqlText: tqlText,
		tqlInst: tql.NewTQL(tqlText),
	}
}

// ID returns mapper id.
func (m *mapper) ID() string {
	return m.id
}

// String returns MQL text.
func (m *mapper) String() string {
	return m.tqlText
}

// TargetEntity returns target entity.
func (m *mapper) TargetEntity() string {
	return m.tqlInst.Target()
}

// SourceEntities returns source entities(include target entity).
func (m *mapper) SourceEntities() []string {
	return m.tqlInst.Entities()
}

// Tentacles returns tentacles.
func (m *mapper) Tentacles() []Tentacler {
	tentacleConfigs := m.tqlInst.Tentacles()
	tentacles := make([]Tentacler, 0, len(tentacleConfigs))
	mItems := make([]WatchKey, 0)

	for _, tentacleConf := range tentacleConfigs {
		eItems := make([]WatchKey, len(tentacleConf.PropertyKeys))
		for index, item := range tentacleConf.PropertyKeys {
			watchKey := WatchKey{
				EntityId:    tentacleConf.SourceEntity,
				PropertyKey: item,
			}
			eItems[index] = watchKey
			mItems = append(mItems, watchKey)
		}

		tentacles = append(tentacles, NewTentacle(TentacleTypeEntity, tentacleConf.SourceEntity, eItems))
	}

	tentacles = append(tentacles, NewTentacle(TentacleTypeMapper, m.id, mItems))

	return tentacles
}

// Copy duplicate a mapper.
func (m *mapper) Copy() Mapper {
	return NewMapper(m.id, m.tqlText)
}

// Exec input returns output.
func (m *mapper) Exec(values map[string]interface{}) (res map[string]interface{}, err error) {
	res, err = m.tqlInst.Exec(values)
	return res, errors.Wrap(err, "execute tql failed")
}
