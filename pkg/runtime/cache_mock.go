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

package runtime

import (
	"context"
)

type cacheMock struct {
	entities map[string]Entity
}

func NewCacheMock(entities map[string]Entity) EntityCache {
	return &cacheMock{entities: entities}
}

func (ec *cacheMock) Load(ctx context.Context, id string) (Entity, error) {
	if state, ok := ec.entities[id]; ok {
		return state, nil
	}

	panic("load cache entity")
}

func (ec *cacheMock) Snapshot() error {
	panic("implement me")
}
