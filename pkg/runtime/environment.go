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
	"strings"

	"github.com/pkg/errors"
	"github.com/tkeel-io/core/pkg/mapper"
	"github.com/tkeel-io/core/pkg/statem"
	"github.com/tkeel-io/kit/log"
	"go.etcd.io/etcd/api/v3/mvccpb"
	"go.uber.org/zap"
)

// return clean targetID.
type CleanHandler func() []string

type EtcdPair struct {
	Key   string
	Value []byte
}

// cache for state marchine.
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

func NewEnv() *Environment {
	return &Environment{
		mapperCaches: make(map[string]*MapperCache),
	}
}

func (env *Environment) generateCleanHandler(stateID, mapperID string, tentacleIDs []string) CleanHandler {
	return func() []string {
		var targets []string
		mCache, ok := env.mapperCaches[stateID]
		if !ok {
			return []string{}
		}

		delete(mCache.mappers, mapperID)
		for _, id := range tentacleIDs {
			if tentacle, ok := mCache.tentacles[id]; ok {
				targets = append(targets, tentacle.TargetID())
				delete(mCache.tentacles, id)
			}
		}
		return targets
	}
}

func (env *Environment) GetEnvBy(stateID string) statem.EnvDescription {
	var envDesc statem.EnvDescription
	if mCache, has := env.mapperCaches[stateID]; has {
		for _, m := range mCache.mappers {
			envDesc.Mappers = append(envDesc.Mappers, m.Copy())
		}
		for _, tentacle := range mCache.tentacles {
			envDesc.Tentacles = append(envDesc.Tentacles, tentacle.Copy())
		}
	}
	return envDesc
}

func (env *Environment) addMapper(m mapper.Mapper) (effects []string) {
	// check mapper exists.
	targetID := m.TargetEntity()
	if _, has := env.mapperCaches[targetID]; !has {
		env.mapperCaches[targetID] = newMapperCache()
	}

	tentacleIDs := make([]string, 0)
	mCache := env.mapperCaches[targetID]
	// check mapper exists in cache.
	if _, exists := mCache.mappers[m.ID()]; exists {
		effects = mCache.cleanHandlers[m.ID()]()
	}

	// generate tentacles.
	for _, tentacle := range m.Tentacles() {
		switch tentacle.Type() {
		case mapper.TentacleTypeEntity:
			effects = append(effects, tentacle.TargetID())
			tentacle = mapper.NewTentacle(tentacle.Type(), targetID, tentacle.Items())
			env.addTentacle(tentacle.TargetID(), tentacle)
		case mapper.TentacleTypeMapper:
			// 如果是Mapper类型的Tentacle，那么将该Tentacle分配到mapper所在stateMarchine.
			mCache.tentacles[tentacle.ID()] = tentacle
		default:
			log.Error("invalid tentacle type", zap.String("target", tentacle.TargetID()), zap.String("type", tentacle.Type()))
		}
		tentacleIDs = append(tentacleIDs, tentacle.ID())
	}

	mCache.mappers[m.ID()] = m
	mCache.cleanHandlers[m.ID()] = env.generateCleanHandler(targetID, m.ID(), tentacleIDs)

	return effects
}

func (env *Environment) removeMapper(stateID, mapperID string) []string {
	if _, exists := env.mapperCaches[stateID]; !exists {
		log.Warn("state marchine environment not found",
			zap.String("stateID", stateID), zap.String("mapperID", mapperID))
		return nil
	}

	// clean mapper.
	mCache, ok := env.mapperCaches[stateID]
	if !ok {
		return []string{}
	}

	return mCache.cleanHandlers[mapperID]()
}

func (env *Environment) addTentacle(stateID string, tentacle mapper.Tentacler) {
	if _, has := env.mapperCaches[stateID]; !has {
		env.mapperCaches[stateID] = newMapperCache()
	}

	mCache := env.mapperCaches[stateID]
	mCache.tentacles[tentacle.ID()] = tentacle
}

func (env *Environment) LoadMapper(pairs []EtcdPair) []KeyInfo {
	var err error
	var info KeyInfo
	var loadEntities []KeyInfo

	for _, pair := range pairs {
		log.Debug("load mapper", zap.String("key", pair.Key), zap.String("value", string(pair.Value)))
		if info, err = parseTQLKey(pair.Key); nil != err {
			log.Error("load mapper", zap.Error(err), zap.String("key", pair.Key), zap.String("value", string(pair.Value)))
			continue
		}

		var mapperInstence mapper.Mapper
		if mapperInstence, err = mapper.NewMapper(pair.Key, string(pair.Value)); nil != err {
			log.Error("parse TQL", zap.String("key", pair.Key), zap.String("value", string(pair.Value)))
			continue
		}

		env.addMapper(mapperInstence)
		if StateMarchineTypeSubscription == info.Type {
			loadEntities = append(loadEntities, info)
		}
	}

	return loadEntities
}

type KeyInfo struct {
	Type     string
	Name     string
	EntityID string
}

func parseTQLKey(key string) (KeyInfo, error) {
	arr := strings.Split(key, ".")
	if len(arr) != 5 {
		return KeyInfo{}, ErrInvalidTQLKey
	}

	// core.mapper.{type}.{entityID}.{name}
	return KeyInfo{Type: arr[2], Name: arr[4], EntityID: arr[3]}, nil
}

func (env *Environment) OnMapperChanged(op mvccpb.Event_EventType, pair EtcdPair) ([]string, error) {
	var (
		err     error
		info    KeyInfo
		effects []string
	)
	switch op {
	case mvccpb.PUT:
		var mapperInstence mapper.Mapper
		log.Debug("tql changed", zap.String("tql.Key", pair.Key), zap.String("tql.Val", string(pair.Value)))
		mapperInstence, err = mapper.NewMapper(pair.Key, string(pair.Value))
		if nil != err {
			log.Error("parse TQL", zap.String("key", pair.Key), zap.String("value", string(pair.Value)))
			return effects, errors.Wrap(err, "mapper changed")
		}

		effects = env.addMapper(mapperInstence)
	case mvccpb.DELETE:
		log.Debug("tql deleted", zap.String("tql.Key", pair.Key), zap.String("tql.Val", string(pair.Value)))
		if info, err = parseTQLKey(pair.Key); nil != err {
			log.Error("load mapper", zap.Error(err), zap.String("key", pair.Key), zap.String("value", string(pair.Value)))
			return effects, errors.Wrap(err, "mapper changed")
		}
		effects = env.removeMapper(info.EntityID, pair.Key)
	default:
		log.Error("invalid etcd operator type", zap.Any("operator", op),
			zap.String("tql.Key", pair.Key), zap.String("tql.Val", string(pair.Value)))
	}

	log.Debug("onMapperChanged", zap.String("key", pair.Key), zap.Any("effects", effects))

	return effects, nil
}
