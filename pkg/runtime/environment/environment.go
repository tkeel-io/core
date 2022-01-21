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
	"github.com/pkg/errors"
	zfield "github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/core/pkg/mapper"
	"github.com/tkeel-io/core/pkg/repository/dao"
	"github.com/tkeel-io/kit/log"
	"go.uber.org/zap"
)

// cache for state machine.
type MapperCache struct {
	mappers       map[string]mapper.Mapper    // map[mapperID]Mapper.
	tentacles     map[string]mapper.Tentacler // tentacle set.
	cleanHandlers map[string]CleanHandler     // map[mapperID]Handler.
}

func newMapperCache() *MapperCache {
	return &MapperCache{
		mappers:       make(map[string]mapper.Mapper),
		tentacles:     make(map[string]mapper.Tentacler),
		cleanHandlers: make(map[string]CleanHandler),
	}
}

type Environment struct {
	mapperCaches map[string]*MapperCache // map[stateID]MapperCache
}

// NewEnvironment returns *Environment.
func NewEnvironment() *Environment {
	return &Environment{
		mapperCaches: make(map[string]*MapperCache),
	}
}

// GetActorEnv returns Actor Environments.
func (env *Environment) GetActorEnv(stateID string) ActorEnv {
	var actorEnv = newActorEnv()
	if mCache, has := env.mapperCaches[stateID]; has {
		for _, m := range mCache.mappers {
			actorEnv.Mappers[m.ID()] = m.Copy()
		}
		for _, tentacle := range mCache.tentacles {
			actorEnv.Tentacles = append(actorEnv.Tentacles, tentacle.Copy())
		}
	}
	return actorEnv
}

func (env *Environment) StoreMappers(mappers []dao.Mapper) []dao.Mapper {
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
func (env *Environment) OnMapperChanged(et dao.EnventType, m dao.Mapper) ([]string, error) {
	var err error
	var effects []string

	switch et {
	case dao.PUT:
		log.Debug("mapper changed", zfield.ID(m.ID), zfield.Name(m.Name), zfield.TQL(m.TQL),
			zfield.Eid(m.EntityID), zfield.Type(m.EntityType), zfield.Desc(m.Description))
		// parse mapper again.
		var mapperInstence mapper.Mapper
		if mapperInstence, err = mapper.NewMapper(m.ID, m.TQL); nil != err {
			log.Error("parse mapper", zap.Error(err), zfield.ID(m.ID), zfield.Eid(m.EntityID))
			return effects, errors.Wrap(err, "parse mapper again")
		}
		// add mapper into caches.
		effects = env.addMapper(mapperInstence)
	case dao.DELETE:
		log.Debug("mapper changed", zfield.ID(m.ID), zfield.Name(m.Name), zfield.TQL(m.TQL),
			zfield.Eid(m.EntityID), zfield.Type(m.EntityType), zfield.Desc(m.Description))
		// remove mapper from caches.
		effects = env.removeMapper(m.EntityID, m.ID)
	default:
		log.Error("invalid operator", zap.Any("operator", et.String()), zfield.ID(m.ID), zfield.Name(m.Name),
			zfield.TQL(m.TQL), zfield.Eid(m.EntityID), zfield.Type(m.EntityType), zfield.Desc(m.Description))
	}

	log.Debug("update environment", zfield.ID(m.ID),
		zfield.Name(m.Name), zfield.Eid(m.EntityID), zap.Strings("effects", effects))
	return effects, nil
}

// generateCleanHandler generate clean-handler for mapper.
func (env *Environment) generateCleanHandler(stateID, mapperID string, tentacleIDs []string) CleanHandler {
	return func() []string {
		var targets []string
		if mCache, ok := env.mapperCaches[stateID]; ok {
			for _, id := range tentacleIDs {
				if tentacle, ok := mCache.tentacles[id]; ok {
					delete(mCache.tentacles, id)
					targets = append(targets, tentacle.TargetID())
				}
			}
			delete(mCache.mappers, mapperID)
			return targets
		}
		return targets
	}
}

// addMapper add mapper into Environment.
func (env *Environment) addMapper(m mapper.Mapper) (effects []string) {
	// check mapper exists.
	targetID := m.TargetEntity()
	if _, has := env.mapperCaches[targetID]; !has {
		env.mapperCaches[targetID] = newMapperCache()
	}

	tentacleIDs := make([]string, 0)
	mCache := env.mapperCaches[targetID]
	if _, exists := mCache.mappers[m.ID()]; exists {
		effects = mCache.cleanHandlers[m.ID()]()
	}

	// generate tentacles.
	for _, tentacle := range m.Tentacles() {
		switch tentacle.Type() {
		case mapper.TentacleTypeEntity:
			remoteID := tentacle.TargetID()
			effects = append(effects, tentacle.TargetID())
			tentacle = mapper.NewTentacle(tentacle.Type(), targetID, tentacle.Items())
			env.addTentacle(remoteID, tentacle)
			log.Info("tentacle ", zap.String("target", tentacle.TargetID()), zap.Any("items", tentacle.Items()))
		case mapper.TentacleTypeMapper:
			// 如果是Mapper类型的Tentacle，那么将该Tentacle分配到mapper所在stateMachine.
			mCache.tentacles[tentacle.ID()] = tentacle
			log.Info("tentacle ", zap.String("target", tentacle.TargetID()), zap.Any("items", tentacle.Items()))
		default:
			log.Error("invalid tentacle type", zap.String("target", tentacle.TargetID()), zap.String("type", tentacle.Type()))
		}
		tentacleIDs = append(tentacleIDs, tentacle.ID())
	}

	mCache.mappers[m.ID()] = m
	mCache.cleanHandlers[m.ID()] = env.generateCleanHandler(targetID, m.ID(), tentacleIDs)

	return effects
}

// removeMapper remove mapper from Environment.
func (env *Environment) removeMapper(stateID, mapperID string) []string {
	if _, exists := env.mapperCaches[stateID]; !exists {
		log.Warn("state machine environment not found",
			zap.String("stateID", stateID), zap.String("mapperID", mapperID))
		return nil
	}

	// clean mapper.
	mCache, ok := env.mapperCaches[stateID]
	if !ok {
		return []string{}
	}

	effects := mCache.cleanHandlers[mapperID]()
	if len(mCache.mappers) == 0 {
		delete(env.mapperCaches, stateID)
	}

	return effects
}

// addTentacle add tentacle into Environment.
func (env *Environment) addTentacle(stateID string, tentacle mapper.Tentacler) {
	if _, has := env.mapperCaches[stateID]; !has {
		env.mapperCaches[stateID] = newMapperCache()
	}

	mCache := env.mapperCaches[stateID]
	mCache.tentacles[tentacle.ID()] = tentacle
}
