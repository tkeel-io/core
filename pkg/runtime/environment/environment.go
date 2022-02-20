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
	"sort"
	"sync"

	"github.com/pkg/errors"
	zfield "github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/core/pkg/mapper"
	"github.com/tkeel-io/core/pkg/repository/dao"
	"github.com/tkeel-io/core/pkg/util"
	"github.com/tkeel-io/kit/log"
	"go.uber.org/zap"
)

type CleanFunc func() []string

// cache for state machine.
type MapperCache struct {
	mid       string
	mapper    mapper.Mapper
	tentacles []mapper.Tentacler
}

type StateCache struct {
	stateID string
	mappers map[string]MapperCache
}

func newMapperCache(id string, m mapper.Mapper) MapperCache {
	return MapperCache{
		mid:    id,
		mapper: m,
	}
}

func newStateCache(id string) *StateCache {
	return &StateCache{
		stateID: id,
		mappers: make(map[string]MapperCache),
	}
}

type Environment struct {
	lock         sync.RWMutex
	stateCaches  map[string]*StateCache
	mapperStates map[string]sort.StringSlice
}

// NewEnvironment returns *Environment.
func NewEnvironment() IEnvironment {
	return &Environment{
		lock:         sync.RWMutex{},
		stateCaches:  make(map[string]*StateCache),
		mapperStates: make(map[string]sort.StringSlice),
	}
}

// GetActorEnv returns Actor Environments.
func (env *Environment) GetStateEnv(stateID string) ActorEnv {
	env.lock.RLock()
	defer env.lock.RUnlock()

	var actorEnv = newActorEnv()
	if sCache, has := env.stateCaches[stateID]; has {
		for _, mc := range sCache.mappers {
			for _, tentacle := range mc.tentacles {
				actorEnv.Tentacles = append(actorEnv.Tentacles, tentacle.Copy())
			}
			if nil != mc.mapper {
				actorEnv.Mappers[mc.mapper.ID()] = mc.mapper.Copy()
			}
		}
	}
	return actorEnv
}

func (env *Environment) StoreMappers(mappers []dao.Mapper) []dao.Mapper {
	env.lock.Lock()
	defer env.lock.Unlock()

	var err error
	for _, m := range mappers {
		log.Debug("store mapper", zfield.ID(m.ID), zfield.Name(m.Name), zfield.TQL(m.TQL),
			zfield.Eid(m.EntityID), zfield.Type(m.EntityType), zfield.Desc(m.Description))

		// parse mapper.
		var mapperInstence mapper.Mapper
		if mapperInstence, err = mapper.NewMapper(m.ID, m.TQL); nil != err {
			log.Error("parse mapper", zap.Error(err), zfield.ID(m.ID), zfield.Eid(m.EntityID))
			continue
		}
		env.addMapper(mapperInstence)
	}
	return mappers
}

// OnMapperChanged respond for mapper.
func (env *Environment) OnMapperChanged(et dao.EnventType, m dao.Mapper) (Effect, error) {
	var err error
	var effect = Effect{
		MapperID: m.ID,
		StateID:  m.EntityID,
	}

	env.lock.Lock()
	defer env.lock.Unlock()

	switch et {
	case dao.PUT:
		log.Debug("mapper changed", zfield.ID(m.ID), zfield.Name(m.Name), zfield.TQL(m.TQL),
			zfield.Eid(m.EntityID), zfield.Type(m.EntityType), zfield.Desc(m.Description))
		// parse mapper again.
		var mapperInstence mapper.Mapper
		if mapperInstence, err = mapper.NewMapper(m.ID, m.TQL); nil != err {
			log.Error("parse mapper", zap.Error(err), zfield.ID(m.ID), zfield.Eid(m.EntityID))
			return effect, errors.Wrap(err, "parse mapper again")
		}
		// add mapper into caches.
		effect.EffectStateIDs = env.addMapper(mapperInstence)
	case dao.DELETE:
		log.Debug("mapper changed", zfield.ID(m.ID), zfield.Name(m.Name), zfield.TQL(m.TQL),
			zfield.Eid(m.EntityID), zfield.Type(m.EntityType), zfield.Desc(m.Description))
		// remove mapper from caches.
		effect.EffectStateIDs = env.removeMapper(m.EntityID, m.ID)
	default:
		log.Error("invalid operator", zap.Any("operator", et.String()), zfield.ID(m.ID), zfield.Name(m.Name),
			zfield.TQL(m.TQL), zfield.Eid(m.EntityID), zfield.Type(m.EntityType), zfield.Desc(m.Description))
	}

	log.Debug("update environment", zfield.ID(m.ID),
		zfield.Name(m.Name), zfield.Eid(m.EntityID), zap.Any("effect", effect))
	return effect, nil
}

// addMapper add mapper into Environment.
func (env *Environment) addMapper(m mapper.Mapper) (effects []string) {
	// create cache if not exist.
	targetID := m.TargetEntity()
	if _, has := env.stateCaches[targetID]; !has {
		env.stateCaches[targetID] = newStateCache(targetID)
	}

	sCache := env.stateCaches[targetID]
	if _, exists := sCache.mappers[m.ID()]; exists {
		effects = env.cleanMapper(m.ID())
	}

	// generate tentacles.
	mCache := newMapperCache(m.ID(), m)
	for _, tentacle := range m.Tentacles() {
		switch tentacle.Type() {
		case mapper.TentacleTypeEntity:
			remoteID := tentacle.TargetID()
			tentacle = mapper.NewTentacle(tentacle.Type(), targetID, tentacle.Items())
			env.addTentacle(remoteID, m.ID(), tentacle)
			log.Info("tentacle ", zap.String("target", tentacle.TargetID()), zap.Any("items", tentacle.Items()))
		case mapper.TentacleTypeMapper:
			// 如果是Mapper类型的Tentacle，那么将该Tentacle分配到mapper所在stateMachine.
			mCache.tentacles = append(mCache.tentacles, tentacle)
			log.Info("tentacle ", zap.String("target", tentacle.TargetID()), zap.Any("items", tentacle.Items()))
		default:
			log.Error("invalid tentacle type", zap.String("target", tentacle.TargetID()), zap.String("type", tentacle.Type()))
		}
	}

	// reset mapper cache.
	sCache.mappers[m.ID()] = mCache
	env.index(targetID, m.ID())

	return util.Unique(append(effects, env.mapperStates[m.ID()]...))
}

// removeMapper remove mapper from Environment.
func (env *Environment) removeMapper(stateID, mapperID string) (effects []string) {
	if _, exists := env.stateCaches[stateID]; !exists {
		log.Warn("state machine environment not found",
			zap.String("stateID", stateID), zap.String("mapperID", mapperID))
		return nil
	}

	// clean mapper.
	if _, ok := env.stateCaches[stateID]; !ok {
		return effects
	}

	effects = append(effects, env.cleanMapper(mapperID)...)
	delete(env.stateCaches[stateID].mappers, mapperID)

	return effects
}

// addTentacle add tentacle into Environment.
func (env *Environment) addTentacle(stateID, mapperID string, tentacle mapper.Tentacler) {
	var (
		has    bool
		sCache *StateCache
		mCache MapperCache
	)

	if _, has = env.stateCaches[stateID]; !has {
		env.stateCaches[stateID] = newStateCache(stateID)
	}

	sCache = env.stateCaches[stateID]
	if mCache, has = sCache.mappers[mapperID]; !has {
		mCache = newMapperCache(mapperID, nil)
	}

	env.index(stateID, mapperID)
	mCache.tentacles = append(mCache.tentacles, tentacle.Copy())
	sCache.mappers[mapperID] = mCache
}

func (env *Environment) index(stateID, mapperID string) {
	if _, ok := env.mapperStates[mapperID]; !ok {
		env.mapperStates[mapperID] = sort.StringSlice{}
	}

	slice := env.mapperStates[mapperID]
	if util.RangeOutIndex != util.Search(slice, stateID) {
		return
	}

	slice = append(slice, stateID)
	slice.Sort()

	// reset index.
	env.mapperStates[mapperID] = slice
}

func (env *Environment) cleanMapper(id string) []string {
	for _, stateID := range env.mapperStates[id] {
		if sCache, has := env.stateCaches[stateID]; has {
			delete(sCache.mappers, id)
		}
	}

	stateIDs := env.mapperStates[id]
	delete(env.mapperStates, id)
	return stateIDs
}
