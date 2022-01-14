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

package environment

import (
	"errors"
	"strings"

	"github.com/tkeel-io/core/pkg/mapper"
	"go.etcd.io/etcd/api/v3/mvccpb"
)

var (
	ErrInvalidTQLKey = errors.New("invalid TQL key")
)

type CleanHandler func() []string

type EtcdPair struct {
	Key   string
	Value []byte
}

// MaSummary: mapper summary.
type MaSummary struct {
	Type     string
	Name     string
	EntityID string
}

// parseTQLKey parse TQL-key.
func parseTQLKey(key string) (MaSummary, error) {
	arr := strings.Split(key, ".")
	if len(arr) != 5 {
		return MaSummary{}, ErrInvalidTQLKey
	}

	// core.mapper.{type}.{entityID}.{name} .
	return MaSummary{Type: arr[2], Name: arr[4], EntityID: arr[3]}, nil
}

type ActorEnv struct {
	Mappers   map[string]mapper.Mapper
	Tentacles []mapper.Tentacler
}

type IEnvironment interface {
	GetActorEnv(string) ActorEnv
	StoreMappers([]EtcdPair) []MaSummary
	OnMapperChanged(mvccpb.Event_EventType, EtcdPair) ([]string, error)
}
