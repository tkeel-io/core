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
	"crypto/rand"
	"fmt"
	"strings"

	"github.com/tkeel-io/tdtl"
)

const (
	TentacleTypeUndefined = "undefined"
	TentacleTypeEntity    = "entity"
	TentacleTypeMapper    = "mapper"

	VersionInited = 0

	WatchKeyDelimiter = "."
)

type IMapper interface {
	// ID returns mapper id.
	ID() string
	Name() string
	// String returns MQL text.
	String() string
	// Version returns mapper version.
	Version() int64
	// TargetEntity returns target entity.
	TargetEntity() string
	// SourceEntities returns source entities.
	SourceEntities() map[string][]string
	// Tentacles returns tentacles.
	Tentacles() map[string][]Tentacler
	// Copy duplicate a mapper.
	Copy() IMapper
	// Exec excute input returns output.
	Exec(map[string]tdtl.Node) (map[string]tdtl.Node, error)
}

type TentacleType = string

type Tentacler interface {
	// ID return id.
	ID() string
	// Type returns tentacle type.
	Type() TentacleType
	String() string
	// Mapper return mapper.
	Mapper() IMapper
	// TargetID returns target id.
	TargetID() string
	// Items returns watch keys(watchKey=entityId#propertyKey).
	Items() []WatchKey
	// Copy duplicate a mapper.
	Copy() Tentacler
	// Version return tentacle version.
	Version() int64
}

type WatchKey struct {
	EntityID    string
	PropertyKey string
}

func NewWatchKey(path string) WatchKey {
	if segs := strings.SplitN(path, WatchKeyDelimiter, 2); len(segs) == 2 {
		return WatchKey{EntityID: segs[0], PropertyKey: segs[1]}
	}
	return WatchKey{}
}

func (wk *WatchKey) String() string {
	return wk.EntityID + WatchKeyDelimiter + wk.PropertyKey
}

// uuid generate an uuid.
func uuid() string {
	uuid := make([]byte, 16)
	if _, err := rand.Read(uuid); err != nil {
		return ""
	}
	// see section 4.1.1.
	uuid[8] = uuid[8]&^0xc0 | 0x80
	// see section 4.1.3.
	uuid[6] = uuid[6]&^0xf0 | 0x40
	return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:])
}
