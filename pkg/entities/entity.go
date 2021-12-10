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

	"github.com/pkg/errors"
	"github.com/tkeel-io/core/pkg/constraint"
	"github.com/tkeel-io/core/pkg/statem"
)

type Entity struct {
	stateMarchine statem.StateMarchiner
}

func newEntity(ctx context.Context, mgr *EntityManager, in *statem.Base) (EntityOp, error) {
	stateM, err := statem.NewState(ctx, mgr, in, nil)
	if nil != err {
		return nil, errors.Wrap(err, "create subscription failed")
	}

	return &Entity{stateMarchine: stateM}, nil
}

// GetID return state marchine id.
func (e *Entity) GetID() string {
	return e.stateMarchine.GetID()
}

// GetBase returns state.Base.
func (e *Entity) GetBase() *statem.Base {
	return e.stateMarchine.GetBase()
}

// Setup state marchine setup.
func (e *Entity) Setup() error {
	return errors.Wrap(e.stateMarchine.Setup(), "entity setup failed")
}

func (e *Entity) SetConfig(configs map[string]constraint.Config) error {
	return errors.Wrap(e.stateMarchine.SetConfig(configs), "entity.SetConfig failed")
}

// OnMessage recv message from pubsub.
func (e *Entity) OnMessage(msg statem.Message) bool {
	return e.stateMarchine.OnMessage(msg)
}

// InvokeMsg dispose entity message.
func (e *Entity) HandleLoop() {
	e.stateMarchine.HandleLoop()
}

// StateManager returns state manager.
func (e *Entity) GetManager() statem.StateManager {
	return e.stateMarchine.GetManager()
}
