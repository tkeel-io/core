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
	"github.com/tkeel-io/kit/log"
	"go.etcd.io/etcd/api/v3/mvccpb"
	"go.uber.org/zap"
)

type EtcdPair struct {
	Key   string
	Value []byte
}

type Environment struct {
}

func NewEnv() *Environment {
	return &Environment{}
}

func (env *Environment) LoadMapper(descs []EtcdPair) error {
	return nil
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
