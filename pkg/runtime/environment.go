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
	"strings"

	"github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/kit/log"
	"go.etcd.io/etcd/api/v3/mvccpb"
	"go.uber.org/zap"
)

type EtcdPair struct {
	Key   string
	Value []byte
}

type Environment struct {
	stateManager *Manager
}

func NewEnv(mgr *Manager) *Environment {
	return &Environment{stateManager: mgr}
}

func (env *Environment) LoadMapper(pairs []EtcdPair) error {
	var err error
	var info KeyInfo
	for _, pair := range pairs {
		log.Info("load mapper & actor", zap.String("key", pair.Key), zap.String("value", string(pair.Value)))
		if info, err = parseTQLKey(pair.Key); nil != err {
			log.Error("load mapper", zap.Error(err), zap.String("key", pair.Key), zap.String("value", string(pair.Value)))
			continue
		}
		if StateMarchineTypeSubscription == info.Type {
			log.Info("load subscription", logger.EntityID(info.EntityID), zap.String("mapper-name", info.Name))
			if err = env.stateManager.loadActor(context.Background(), info.Type, info.EntityID); nil != err {
				log.Error("load Actor", zap.Error(err), zap.String("key", pair.Key), zap.String("value", string(pair.Value)))
			}
		}
	}

	return nil
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
	case mvccpb.DELETE:
		log.Info("tql deleted", zap.String("tql.Key", pair.Key), zap.String("tql.Val", string(pair.Value)))
	default:
		log.Error("invalid etcd operator type", zap.Any("operator", op),
			zap.String("tql.Key", pair.Key), zap.String("tql.Val", string(pair.Value)))
	}
	return nil
}
