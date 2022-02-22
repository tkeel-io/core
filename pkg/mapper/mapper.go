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

package mapper

import (
	"github.com/pkg/errors"
	"github.com/tkeel-io/tdtl"
)

type mapper struct {
	id      string
	tqlText string
	tqlInst tdtl.TDTL
}

func NewMapper(id, tqlText string) (Mapper, error) {
	tqlInst, err := tdtl.NewTDTL(tqlText, nil)
	if nil != err {
		return nil, errors.Wrap(err, "construct mapper")
	}
	return &mapper{
		id:      id,
		tqlText: tqlText,
		tqlInst: tqlInst,
	}, nil
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
				EntityID:    tentacleConf.SourceEntity,
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
	mCopy, _ := NewMapper(m.id, m.tqlText)
	return mCopy
}

// Exec input returns output.
func (m *mapper) Exec(values map[string]tdtl.Node) (map[string]tdtl.Node, error) {
	res, err := m.tqlInst.Exec(values)
	return res, errors.Wrap(err, "execute tql")
}
