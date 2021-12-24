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
	"github.com/tkeel-io/kit/log"
	"go.etcd.io/etcd/api/v3/mvccpb"
	"go.uber.org/zap"
)

type EtcdPair struct {
	Key   string
	Value []byte
}

type Environment struct {
	mappers   map[string][]mapper.Tentacler
	tentacles map[string][]mapper.Tentacler
}

func NewEnv() *Environment {
	return &Environment{
		mappers:   make(map[string][]mapper.Tentacler),
		tentacles: make(map[string][]mapper.Tentacler),
	}
}

func (env *Environment) LoadMapper(pairs []EtcdPair) []KeyInfo {
	var err error
	var info KeyInfo
	var loadEntities []KeyInfo
	for _, pair := range pairs {
		log.Info("load mapper & actor", zap.String("key", pair.Key), zap.String("value", string(pair.Value)))
		if info, err = parseTQLKey(pair.Key); nil != err {
			log.Error("load mapper", zap.Error(err), zap.String("key", pair.Key), zap.String("value", string(pair.Value)))
			continue
		}

		// TODO: 这里处理所有的pair.Value.
		mapperInstence, err := mapper.NewMapper("", string(pair.Value))
		if nil != err {
			log.Error("parse TQL", zap.String("key", pair.Key), zap.String("value", string(pair.Value)))
			continue
		}

		tentacles := mapperInstence.Tentacles()
		env.mappers[pair.Key] = append(env.mappers[pair.Key], tentacles...)

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

	return KeyInfo{Type: arr[2], Name: arr[4], EntityID: arr[3]}, nil
}

func (env *Environment) OnMapperChanged(op mvccpb.Event_EventType, pair EtcdPair) error {
	switch op {
	case mvccpb.PUT:
		log.Info("tql changed", zap.String("tql.Key", pair.Key), zap.String("tql.Val", string(pair.Value)))
		mapperInstence, err := mapper.NewMapper("", string(pair.Value))
		if nil != err {
			log.Error("parse TQL", zap.String("key", pair.Key), zap.String("value", string(pair.Value)))
			return errors.Wrap(err, "mapper changed")
		}

		tentacles := mapperInstence.Tentacles()
		env.mappers[pair.Key] = append(env.mappers[pair.Key], tentacles...)
	case mvccpb.DELETE:
		log.Info("tql deleted", zap.String("tql.Key", pair.Key), zap.String("tql.Val", string(pair.Value)))
	default:
		log.Error("invalid etcd operator type", zap.Any("operator", op),
			zap.String("tql.Key", pair.Key), zap.String("tql.Val", string(pair.Value)))
	}
	return nil
}
