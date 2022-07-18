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
	"github.com/tkeel-io/core/pkg/repository"
	"sync"
)

type SchemaStore struct {
	sync.RWMutex
	data map[string]*repository.Schema
}

func NewSchemaStore() *SchemaStore {
	return &SchemaStore{
		RWMutex: sync.RWMutex{},
		data:    map[string]*repository.Schema{},
	}
}

func (s SchemaStore) Get(schemaID string) *repository.Schema {
	s.RLock()
	defer s.RUnlock()
	sm, ok := s.data[schemaID]
	if !ok{
		return nil
	}
	return sm
}

func (s SchemaStore) Set(schemaID string, sm *repository.Schema) bool {
	s.Lock()
	defer s.Unlock()
	_, ok := s.data[schemaID]
	s.data[schemaID] = sm
	return ok
}

func (s SchemaStore) Del(schemaID string) bool {
	s.Lock()
	defer s.Unlock()
	_, ok := s.data[schemaID]
	delete(s.data,schemaID)
	return ok
}
