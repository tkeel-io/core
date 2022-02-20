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

	"github.com/tkeel-io/core/pkg/mapper"
	"github.com/tkeel-io/core/pkg/repository/dao"
)

var (
	ErrInvalidTQLKey = errors.New("invalid TQL key")
)

type CleanHandler func() []string

type ActorEnv struct {
	Mappers   map[string]mapper.Mapper
	Tentacles []mapper.Tentacler
}

type Effect struct {
	StateID        string
	MapperID       string
	EffectStateIDs []string
}

func newActorEnv() ActorEnv {
	return ActorEnv{Mappers: make(map[string]mapper.Mapper)}
}

type IEnvironment interface {
	GetActorEnv(string) ActorEnv
	StoreMappers([]dao.Mapper) []dao.Mapper
	OnMapperChanged(dao.EnventType, dao.Mapper) (Effect, error)
}
