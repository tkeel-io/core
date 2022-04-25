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

import "log"

type tentacle struct {
	id       string
	tp       TentacleType
	version  int64
	targetID string
	mapper   IMapper
	items    []WatchKey // key=entityId#propertyKey
}

func NewTentacle(mp IMapper, tp TentacleType, targetID string, items []WatchKey, version int64) Tentacler {
	return &tentacle{
		id:       uuid(),
		tp:       tp,
		items:    items,
		mapper:   mp,
		version:  version,
		targetID: targetID,
	}
}

func (t *tentacle) ID() string {
	return t.id
}

// Type returns tentacle type.
func (t *tentacle) Type() TentacleType {
	return t.tp
}

// TargetID returns target id.
func (t *tentacle) TargetID() string {
	return t.targetID
}

func (t *tentacle) String() string {
	return t.targetID + t.id
}

// Items returns watch keys(watchKey=entityId#propertyKey).
func (t *tentacle) Items() []WatchKey {
	return t.items
}

func (t *tentacle) Version() int64 {
	return t.version
}

func (t *tentacle) Mapper() IMapper {
	return t.mapper
}

func (t *tentacle) Copy() Tentacler {
	items := make([]WatchKey, len(t.items))
	for index, item := range t.items {
		items[index] = item
	}

	ten := &tentacle{
		id:       t.id,
		tp:       t.tp,
		items:    items,
		version:  t.version,
		targetID: t.targetID,
	}

	t.version++
	return ten
}

func MergeTentacles(tentacles ...Tentacler) Tentacler {
	if len(tentacles) == 0 {
		return nil
	}

	tentacle0, ok := tentacles[0].(*tentacle)
	if !ok {
		log.Fatalln("not want struct")
	}

	var version int64
	itemMap := make(map[string]WatchKey)
	for _, tentacle := range tentacles {
		for _, item := range tentacle.Items() {
			itemMap[item.String()] = item
		}
		if tentacle.Version() > version {
			version = tentacle.Version()
		}
	}

	index := -1
	items := make([]WatchKey, len(itemMap))
	for _, item := range itemMap {
		index++
		items[index] = item
	}

	return NewTentacle(tentacle0.mapper, tentacle0.tp, tentacle0.targetID, items, version+1)
}
