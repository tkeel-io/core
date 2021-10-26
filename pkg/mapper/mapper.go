package mapper

import (
	"github.com/pkg/errors"
	"github.com/tkeel-io/core/pkg/mql"
)

type mapper struct {
	id      string
	mqlText string
}

func newMapper(id, mqlText string) *mapper {
	return &mapper{id: id, mqlText: mqlText}
}

// ID returns mapper id.
func (m *mapper) ID() string {
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
	mqlInst := mql.NewMQL(m.mqlText)
	tts := mqlInst.Tentacles()
	tentacles := make([]Tentacler, 0, len(tts))
	mItems := make([]WatchKey, 0)

	for entityID, items := range tts {
		eItems := make([]WatchKey, len(items))
		for index, item := range items {
			watchKey := WatchKey{
				EntityId:    entityID,
				PropertyKey: item,
			}
			eItems[index] = watchKey
			mItems = append(mItems, watchKey)
		}

		tentacles = append(tentacles, NewTentacle(TentacleTypeEntity, entityID, eItems))
	}

	tentacles = append(tentacles, NewTentacle(TentacleTypeMapper, m.id, mItems))

	return tentacles
}

// Copy duplicate a mapper.
func (m *mapper) Copy() Mapper {
	return newMapper(m.id, m.mqlText)
}

// Exec input returns output.
func (m *mapper) Exec(values map[string]interface{}) (res map[string]interface{}, err error) {
	res, err = mql.NewMQL(m.mqlText).Exec(values)
	if err != nil {
		return nil, errors.Unwrap(err)
	}
	return res, nil
}
