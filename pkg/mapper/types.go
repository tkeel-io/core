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

import "github.com/tkeel-io/core/pkg/constraint"

const (
	TentacleTypeUndefined = "undefined"
	TentacleTypeEntity    = "entity"
	TentacleTypeMapper    = "mapper"

	WatchKeyDelimiter = "."
)

type Mapper interface {
	// ID returns mapper id.
	ID() string
	// String returns MQL text.
	String() string
	// TargetEntity returns target entity.
	TargetEntity() string
	// SourceEntities returns source entities.
	SourceEntities() []string
	// Tentacles returns tentacles.
	Tentacles() []Tentacler
	// Copy duplicate a mapper.
	Copy() Mapper
	// Exec excute input returns output.
	Exec(map[string]constraint.Node) (map[string]constraint.Node, error)
}

type TentacleType = string

type Tentacler interface {
	// ID return id.
	ID() string
	// Type returns tentacle type.
	Type() TentacleType
	// TargetID returns target id.
	TargetID() string
	// Items returns watch keys(watchKey=entityId#propertyKey).
	Items() []WatchKey
	// Copy duplicate a mapper.
	Copy() Tentacler
	// IsRemote return remote flag.
	IsRemote() bool
}

type WatchKey struct {
	EntityId    string //nolint
	PropertyKey string
}

func (wk *WatchKey) String() string {
	return wk.EntityId + WatchKeyDelimiter + wk.PropertyKey
}
