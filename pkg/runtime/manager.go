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
	"sync"

	cloudevents "github.com/cloudevents/sdk-go"
	ants "github.com/panjf2000/ants/v2"
	"github.com/pkg/errors"
	"github.com/tkeel-io/core/pkg/constraint"
	xerrors "github.com/tkeel-io/core/pkg/errors"
	zfield "github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/core/pkg/repository/dao"
	"github.com/tkeel-io/core/pkg/runtime/environment"
	"github.com/tkeel-io/core/pkg/runtime/message"
	"github.com/tkeel-io/core/pkg/runtime/state"
	"github.com/tkeel-io/core/pkg/runtime/subscription"
	"github.com/tkeel-io/kit/log"
	"go.uber.org/zap"
)

type Manager struct {
	coroutinePool   *ants.Pool
	containers      map[string]*Container
	msgCh           chan message.Context
	disposeCh       chan message.Context
	actorEnv        environment.IEnvironment
	resourceManager state.ResourceManager
	republisher     state.Republisher

	shutdown chan struct{}
	lock     sync.RWMutex
	ctx      context.Context
	cancel   context.CancelFunc
}

func NewManager(ctx context.Context, resourceManager state.ResourceManager) (state.Manager, error) {
	coroutinePool, err := ants.NewPool(5000)
	if err != nil {
		return nil, errors.Wrap(err, "new coroutine pool")
	}

	ctx, cancel := context.WithCancel(ctx)
	stateManager := &Manager{
		ctx:             ctx,
		cancel:          cancel,
		actorEnv:        environment.NewEnvironment(),
		containers:      make(map[string]*Container),
		msgCh:           make(chan message.Context, 10),
		disposeCh:       make(chan message.Context, 10),
		resourceManager: resourceManager,
		coroutinePool:   coroutinePool,
		lock:            sync.RWMutex{},
	}

	// set default container.
	stateManager.containers["default"] = NewContainer()
	return stateManager, nil
}

func (m *Manager) Start() error {
	m.initialize()
	go m.watchResource()

	go func() {
		for {
			select {
			case <-m.ctx.Done():
				log.Info("entity manager exited.")
				return
			case msgCtx := <-m.msgCh:
				// dispatch message. 将消息分发到不同的节点。
				m.disposeCh <- msgCtx

			case msgCtx := <-m.disposeCh:
				var err error
				entityID := msgCtx.Get(message.ExtEntityID)
				channelID := msgCtx.Get(message.ExtChannelID)
				log.Debug("dispose message", zfield.ID(entityID), zfield.Message(msgCtx))
				channelID, stateMachine := m.getMachiner(channelID, entityID)
				if nil == stateMachine {
					en := &dao.Entity{
						ID:         entityID,
						Type:       msgCtx.Get(message.ExtEntityType),
						Owner:      msgCtx.Get(message.ExtEntityOwner),
						Source:     msgCtx.Get(message.ExtEntitySource),
						TemplateID: msgCtx.Get(message.ExtTemplateID),
					}

					// load entity, create if not exists.
					if stateMachine, err = m.loadOrCreate(m.ctx, channelID, true, en); nil != err {
						log.Error("disposing message", zap.Error(err),
							zfield.ID(entityID), zap.String("channel", channelID), zfield.Message(msgCtx))
						continue
					}
				}

				if stateMachine.OnMessage(msgCtx) {
					// attatch goroutine to entity.
					m.coroutinePool.Submit(stateMachine.HandleLoop)
				}
			case <-m.shutdown:
				log.Info("state machine manager exit.")
				return
			}
		}
	}()

	return nil
}

func (m *Manager) Shutdown() error {
	m.cancel()
	m.shutdown <- struct{}{}
	return nil
}

func (m *Manager) SetRepublisher(republisher state.Republisher) {
	m.republisher = republisher
}

func (m *Manager) RouteMessage(ctx context.Context, ev cloudevents.Event) error {
	// assume single node.
	log.Debug("route event", zfield.ID(ev.ID()), zfield.Type(ev.Type()), zfield.Event(ev))
	return errors.Wrap(m.republisher.RouteMessage(ctx, ev), "route message")
}

func (m *Manager) HandleMessage(ctx context.Context, ev cloudevents.Event) error {
	log.Debug("handle event", zfield.ID(ev.ID()), zfield.Type(ev.Type()), zfield.Event(ev))

	var err error
	var msgCtx message.Context
	if msgCtx, err = message.From(ctx, ev); nil != err {
		log.Error("parse event", zfield.ID(ev.ID()), zfield.Event(ev))
		return errors.Wrap(err, "parse event")
	}

	// squash properties.
	switch msg := msgCtx.Message().(type) {
	case message.StateMessage:
		// ignore this message.
	case message.PropertyMessage:
		requireds := make(map[string]string)
		for name, reserved := range state.RequiredFields {
			if reserved {
				if prop, has := msg.Properties[name]; has {
					requireds[name] = constraint.Unwrap(prop)
				}
			}
			delete(msg.Properties, name)
		}

		// squash fields.
		for key, val := range state.SquashFields(requireds) {
			msgCtx.Set(key, val)
		}
	default:
		log.Error("invalid message type",
			zfield.Header(msgCtx.Attributes()))
	}

	m.msgCh <- msgCtx
	msgCtx.Wait()

	return nil
}

// Resource return resource manager.
func (m *Manager) Resource() state.ResourceManager {
	return m.resourceManager
}

func (m *Manager) getMachiner(cid, eid string) (string, state.Machiner) {
	if cid == "" {
		cid = "default"
	}

	if container, ok := m.containers[cid]; ok {
		if sm := container.Get(eid); nil != sm {
			if sm.GetStatus() == state.SMStatusDeleted {
				container.Remove(eid)
				return cid, nil
			}
			return cid, sm
		}
	}

	for channelID, container := range m.containers {
		if sm := container.Get(eid); sm != nil {
			if sm.GetStatus() == state.SMStatusDeleted {
				container.Remove(eid)
				return cid, nil
			}

			if channelID == "default" && cid != channelID {
				container.Remove(sm.GetID())
				if _, ok := m.containers[cid]; !ok {
					m.containers[cid] = NewContainer()
				}
				m.containers[cid].Add(sm)
			}
			return cid, sm
		}
	}

	return cid, nil
}

func (m *Manager) isThisNode() bool {
	return true
}

func (m *Manager) reloadActor(stateIDs []string) error {
	// 判断 actor 是否在当前节点.
	if m.isThisNode() {
		var err error
		for _, stateID := range stateIDs {
			var stateMachine state.Machiner
			base := &dao.Entity{ID: stateID, Type: SMTypeBasic}
			if _, stateMachine = m.getMachiner("", stateID); nil != stateMachine {
				log.Warn("load state machine", zfield.ID(stateID))
			} else if stateMachine, err = m.loadOrCreate(m.ctx, "", false, base); nil == err {
				continue
			}
			actorEnv := m.actorEnv.GetActorEnv(stateID)
			stateMachine.WithContext(state.NewContext(stateMachine, actorEnv.Mappers, actorEnv.Tentacles))
		}
	}
	return nil
}

func (m *Manager) loadActor(ctx context.Context, typ string, id string) error {
	_, err := m.loadOrCreate(ctx, "", false, &dao.Entity{ID: id, Type: typ})
	return errors.Wrap(err, "load entity")
}

func (m *Manager) loadOrCreate(ctx context.Context, channelID string, flagCreate bool, base *dao.Entity) (sm state.Machiner, err error) {
	log.Debug("load or create actor", zfield.ID(base.ID),
		zap.String("type", base.Type), zap.String("owner", base.Owner), zap.String("source", base.Source))

	var en *dao.Entity
	if en, err = m.repo().GetEntity(ctx, base); nil == err {
		base = en
	} else {
		log.Warn("load or create actor", zap.Error(err),
			zfield.Eid(base.ID), zfield.Type(base.Type), zfield.Template(base.TemplateID))

		// notfound.
		if !flagCreate || !errors.Is(err, xerrors.ErrEntityNotFound) {
			return nil, errors.Wrap(err, "load or create actor")
		}
	}

	switch base.Type {
	case SMTypeSubscription:
		if sm, err = subscription.NewSubscription(ctx, m, base); nil != err {
			return nil, errors.Wrap(err, "load subscription")
		}
	default:
		// default base entity type.
		if sm, err = state.NewState(ctx, m, base, nil); nil != err {
			return nil, errors.Wrap(err, "load state machine")
		}
	}

	if channelID == "" {
		channelID = "defult"
	}

	if _, has := m.containers[channelID]; !has {
		m.containers[channelID] = NewContainer()
	}

	thisActorEnv := m.actorEnv.GetActorEnv(sm.GetID())
	sm.WithContext(state.NewContext(sm, thisActorEnv.Mappers, thisActorEnv.Tentacles))

	m.containers[channelID].Add(sm)
	return sm, nil
}

// Tools.

func (m *Manager) EscapedEntities(expression string) []string {
	return []string{expression}
}
