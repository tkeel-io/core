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

// mapper info.
type Mapper struct {
	ID          string
	TQL         string
	Name        string
	Owner       string
	EntityID    string
	Description string
}

type mapper struct {
	version int64
	mapper  Mapper
	tqlInst tdtl.TDTL
}

func NewMapper(mp Mapper, version int64) (IMapper, error) {
	tqlInst, err := tdtl.NewTDTL(mp.TQL, nil)
	if nil != err {
		return nil, errors.Wrap(err, "construct mapper")
	}
	return &mapper{
		mapper:  mp,
		version: version,
		tqlInst: tqlInst,
	}, nil
}

func fmtMapperID(entityID, mapperID string) string {
	return entityID + "-" + mapperID
}

// ID returns mapper id.
func (m *mapper) ID() string {
	return fmtMapperID(m.mapper.EntityID, m.mapper.ID)
}

func (m *mapper) Name() string {
	return m.mapper.Name
}

// String returns MQL text.
func (m *mapper) String() string {
	return m.mapper.TQL
}

func (m *mapper) Version() int64 {
	return m.version
}

// TargetEntity returns target entity.
func (m *mapper) TargetEntity() string {
	return m.tqlInst.Target()
}

// SourceEntities returns source entities(include target entity).
func (m *mapper) SourceEntities() map[string][]string {
	return m.tqlInst.Entities()
}

// Tentacles returns tentacles.
func (m *mapper) Tentacles() map[string][]Tentacler {
	tentacleConfigs := tentacles(m.tqlInst)
	tentacles := make(map[string][]Tentacler)
	mItems := make([]WatchKey, 0)

	for _, tentacleConf := range tentacleConfigs {
		entityID := tentacleConf.SourceEntity
		eItems := make([]WatchKey, len(tentacleConf.PropertyKeys))
		for index, item := range tentacleConf.PropertyKeys {
			watchKey := WatchKey{
				EntityID:    entityID,
				PropertyKey: item}
			eItems[index] = watchKey
			mItems = append(mItems, watchKey)
		}
		tentacles[entityID] = append(tentacles[entityID],
			NewTentacle(nil, TentacleTypeEntity, m.TargetEntity(), eItems, m.version))
	}

	targetEid := m.TargetEntity()
	tentacles[targetEid] = append(tentacles[targetEid],
		NewTentacle(m, TentacleTypeMapper, fmtMapperID(m.TargetEntity(), m.mapper.ID), mItems, m.version))

	return tentacles
}

// Copy duplicate a mapper.
func (m *mapper) Copy() IMapper {
	mCopy, _ := NewMapper(m.mapper, m.version)
	return mCopy
}

// Exec input returns output.
func (m *mapper) Exec(values map[string]tdtl.Node) (map[string]tdtl.Node, error) {
	res, err := m.tqlInst.Exec(values)
	return res, errors.Wrap(err, "execute tql")
}
