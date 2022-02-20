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

	"github.com/pkg/errors"
	zfield "github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/core/pkg/repository"
	"github.com/tkeel-io/core/pkg/repository/dao"
	"github.com/tkeel-io/core/pkg/runtime/state"
	"github.com/tkeel-io/core/pkg/runtime/subscription"
	"github.com/tkeel-io/kit/log"
	"go.uber.org/zap"
)

type Container struct {
	id      string
	states  map[string]state.Machiner
	manager *Manager

	ctx    context.Context
	cancel context.CancelFunc
}

func NewContainer(ctx context.Context, id string, mgr *Manager) *Container {
	ctx, cancel := context.WithCancel(ctx)
	return &Container{
		ctx:     ctx,
		cancel:  cancel,
		manager: mgr,
		states:  make(map[string]state.Machiner),
	}
}

func (c *Container) Remove(stateID string) {
	delete(c.states, stateID)
}

func (c *Container) Add(en *dao.Entity) (machine state.Machiner, err error) {
	// make machine.
	if machine, err = makeMachine(c.ctx, c.manager, en); nil != err {
		log.Error("make state machine", zap.Error(err), zfield.ID(en.ID), zfield.Channel(c.id))
		return machine, errors.Wrap(err, "make state machine")
	}

	// load state context.
	stateEnv := c.manager.actorEnv.GetActorEnv(en.ID)
	machine.Context().LoadEnvironments(stateEnv)

	// add machine into container.
	c.states[en.ID] = machine
	return machine, errors.Wrap(err, "add machine")
}

func (c *Container) Load(ctx context.Context, stateID string) (machine state.Machiner, err error) {
	var (
		ok   bool
		en   *dao.Entity
		repo repository.IRepository
	)
	// 1. load from container.
	if machine, ok = c.states[stateID]; ok {
		return machine, nil
	}

	// 2. load from state store.
	log.Info("load machine from store", zfield.ID(stateID), zfield.Channel(c.id))

	repo = c.manager.Resource().Repo()
	if en, err = repo.GetEntity(ctx, &dao.Entity{ID: stateID}); nil != err {
		log.Warn("load machine from store",
			zap.Error(err), zfield.ID(stateID), zfield.Channel(c.id))
		return nil, errors.Wrap(err, "load entity from store")
	}

	// make machine.
	if machine, err = makeMachine(c.ctx, c.manager, en); nil != err {
		log.Error("make state machine", zap.Error(err), zfield.ID(en.ID), zfield.Channel(c.id))
		return machine, errors.Wrap(err, "make state machine")
	}

	// load state context.
	stateEnv := c.manager.actorEnv.GetActorEnv(en.ID)
	machine.Context().LoadEnvironments(stateEnv)

	// add machine into container.
	c.states[en.ID] = machine
	return machine, nil
}

func (c *Container) Close() {}

func makeMachine(ctx context.Context, mgr *Manager, en *dao.Entity) (machine state.Machiner, err error) {
	// make state machine.
	switch en.Type {
	case SMTypeSubscription:
		if machine, err = subscription.NewSubscription(ctx, en); nil != err {
			log.Error("load subscription", zap.Error(err), zfield.Eid(en.ID), zfield.Type(en.Type))
		}
	default:
		machine, err = state.NewState(ctx, en, mgr.dispatcher, mgr.resourceManager, nil)
		if nil != err {
			log.Error("load machine", zap.Error(err), zfield.Eid(en.ID), zfield.Type(en.Type))
		}
	}

	return machine, errors.Wrap(err, "make machine")
}
