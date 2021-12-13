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

package entities

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"sync"

	dapr "github.com/dapr/go-sdk/client"
	"github.com/pkg/errors"
	pb "github.com/tkeel-io/core/api/core/v1"
	"github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/core/pkg/runtime"
	"github.com/tkeel-io/core/pkg/statem"
	"github.com/tkeel-io/kit/log"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
)

const EntityStateName = "core-state"

type EntityManager struct {
	daprClient   dapr.Client
	searchClient pb.SearchHTTPServer
	stateManager *runtime.Manager

	lock   sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc
}

func NewEntityManager(ctx context.Context, mgr *runtime.Manager, searchClient pb.SearchHTTPServer) (*EntityManager, error) {
	daprClient, err := dapr.NewClient()
	if nil != err {
		return nil, errors.Wrap(err, "create entity manager failed")
	}

	ctx, cancel := context.WithCancel(ctx)

	return &EntityManager{
		ctx:          ctx,
		cancel:       cancel,
		stateManager: mgr,
		daprClient:   daprClient,
		searchClient: searchClient,
		lock:         sync.RWMutex{},
	}, nil
}

func (m *EntityManager) Start() error {
	return errors.Wrap(m.stateManager.Start(), "start entity manager")
}

// ------------------------------------APIs-----------------------------.

// CreateEntity create a entity.
func (m *EntityManager) CreateEntity(ctx context.Context, base *statem.Base) (*statem.Base, error) {
	if base.ID == "" {
		base.ID = uuid()
	}

	bytes, _ := json.Marshal(base)
	m.daprClient.SaveState(ctx, EntityStateName, base.ID, bytes)
	return base, nil
}

// DeleteEntity delete an entity from manager.
func (m *EntityManager) DeleteEntity(ctx context.Context, en *statem.Base) (*statem.Base, error) {
	// 1. delete from runtime.
	base, err := m.stateManager.DeleteStateMarchin(ctx, en)
	if nil != err {
		return nil, errors.Wrap(err, "delete entity from runtime")
	}

	// 2. delete from elasticsearch.
	if _, err = m.searchClient.DeleteByID(ctx, &pb.DeleteByIDRequest{Id: en.ID}); nil != err {
		return nil, errors.Wrap(err, "delete entity from es state")
	}

	// 3. delete from state.
	if err = m.daprClient.DeleteState(ctx, EntityStateName, en.ID); nil != err {
		return nil, errors.Wrap(err, "delete entity from state")
	}

	// 4. log record.
	log.Info("delete entity", logger.EntityID(en.ID), zap.Any("entity", base))

	return base, nil
}

// GetProperties returns statem.Base.
func (m *EntityManager) GetProperties(ctx context.Context, en *statem.Base) (*statem.Base, error) {
	base, err := m.getEntityFromState(ctx, en)
	return base, errors.Wrap(err, "get entity properties")
}

func (m *EntityManager) getEntityFromState(ctx context.Context, en *statem.Base) (*statem.Base, error) {
	var base statem.Base
	data, err := m.daprClient.GetState(ctx, EntityStateName, en.ID)
	if nil != err {
		return nil, errors.Wrap(err, "get entity properties")
	} else if err = json.Unmarshal(data.Value, &base); nil != err {
		return nil, errors.Wrap(err, "get entity properties")
	}

	return &base, nil
}

// SetProperties set properties into entity.
func (m *EntityManager) SetProperties(ctx context.Context, en *statem.Base) (*statem.Base, error) {
	if err := m.stateManager.SetProperties(ctx, en); nil != err {
		return nil, errors.Wrap(err, "set entity properties")
	}

	base, err := m.getEntityFromState(ctx, en)
	return base, errors.Wrap(err, "set entity properties")
}

func (m *EntityManager) PatchEntity(ctx context.Context, en *statem.Base, patchData []*pb.PatchData) (*statem.Base, error) {
	if err := m.stateManager.PatchEntity(ctx, en, patchData); nil != err {
		return nil, errors.Wrap(err, "patch entity properties")
	}

	base, err := m.getEntityFromState(ctx, en)
	return base, errors.Wrap(err, "patch entity properties")
}

// SetProperties set properties into entity.
func (m *EntityManager) SetConfigs(ctx context.Context, en *statem.Base) (*statem.Base, error) {
	if err := m.stateManager.SetConfigs(ctx, en); nil != err {
		return nil, errors.Wrap(err, "set entity configs")
	}

	base, err := m.getEntityFromState(ctx, en)
	return base, errors.Wrap(err, "set entity configs")
}

// AppendMapper append a mapper into entity.
func (m *EntityManager) AppendMapper(ctx context.Context, en *statem.Base) (*statem.Base, error) {
	if err := m.stateManager.AppendMapper(ctx, en); nil != err {
		return nil, errors.Wrap(err, "append mapper")
	}

	base, err := m.getEntityFromState(ctx, en)
	return base, errors.Wrap(err, "append mapper")
}

// DeleteMapper delete mapper from entity.
func (m *EntityManager) RemoveMapper(ctx context.Context, en *statem.Base) (*statem.Base, error) {
	if err := m.stateManager.RemoveMapper(ctx, en); nil != err {
		return nil, errors.Wrap(err, "remove mapper")
	}

	base, err := m.getEntityFromState(ctx, en)
	return base, errors.Wrap(err, "remove mapper")
}

func (m *EntityManager) SearchFlush(ctx context.Context, values map[string]interface{}) error {
	var err error
	var val *structpb.Value
	if val, err = structpb.NewValue(values); nil != err {
		log.Error("search index failed.", zap.Error(err))
	} else if _, err = m.searchClient.Index(ctx, &pb.IndexObject{Obj: val}); nil != err {
		log.Error("search index failed.", zap.Error(err))
	}
	return errors.Wrap(err, "SearchFlushfailed")
}

// uuid generate an uuid.
func uuid() string {
	uuid := make([]byte, 16)
	if _, err := rand.Read(uuid); err != nil {
		return ""
	}
	// see section 4.1.1.
	uuid[8] = uuid[8]&^0xc0 | 0x80
	// see section 4.1.3.
	uuid[6] = uuid[6]&^0xf0 | 0x40
	return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:])
}
