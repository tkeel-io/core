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

	ants "github.com/panjf2000/ants/v2"
	"github.com/pkg/errors"
	xerrors "github.com/tkeel-io/core/pkg/errors"
	zfield "github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/core/pkg/repository/dao"
	"github.com/tkeel-io/core/pkg/runtime/environment"
	"github.com/tkeel-io/core/pkg/runtime/message"
	"github.com/tkeel-io/core/pkg/runtime/statem"
	"github.com/tkeel-io/core/pkg/runtime/subscription"
	"github.com/tkeel-io/kit/log"
	"go.uber.org/zap"
)

type Manager struct {
	coroutinePool   *ants.Pool
	containers      map[string]*Container
	msgCh           chan message.MessageContext
	disposeCh       chan message.MessageContext
	actorEnv        environment.IEnvironment
	resourceManager statem.ResourceManager

	shutdown chan struct{}
	lock     sync.RWMutex
	ctx      context.Context
	cancel   context.CancelFunc
}

func NewManager(ctx context.Context, resourceManager statem.ResourceManager) (statem.StateManager, error) {
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
		msgCh:           make(chan message.MessageContext, 10),
		disposeCh:       make(chan message.MessageContext, 10),
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
				eid := msgCtx.Headers.GetReceiver()
				channelID := msgCtx.Headers.Get(message.MsgCtxHeaderChannelID)
				log.Debug("dispose message", zfield.ID(eid), zfield.Message(msgCtx))
				channelID, stateMachine := m.getStateMachine(channelID, eid)
				if nil == stateMachine {
					var err error
					en := &dao.Entity{
						ID:     eid,
						Type:   msgCtx.Headers.GetType(),
						Owner:  msgCtx.Headers.GetOwner(),
						Source: msgCtx.Headers.GetSource(),
					}
					stateMachine, err = m.loadOrCreate(m.ctx, channelID, true, en)
					if nil != err {
						log.Error("disposing message", zap.Error(err),
							zfield.ID(eid), zap.String("channel", channelID), zfield.Message(msgCtx))
						continue
					}
				}

				if stateMachine.OnMessage(msgCtx.Message) {
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

func (m *Manager) RouteMessage(ctx context.Context, msgCtx message.MessageContext) error {
	// assume single node.
	log.Debug("route message",
		zfield.ReqID(msgCtx.Headers.GetRequestID()),
		zfield.MsgID(msgCtx.Headers.GetMessageID()),
		zfield.Sender(msgCtx.Headers.GetSender()),
		zfield.Receiver(msgCtx.Headers.GetReceiver()))

	return m.HandleMessage(ctx, msgCtx)
}

func (m *Manager) HandleMessage(ctx context.Context, msgCtx message.MessageContext) error {
	log.Debug("handle message",
		zfield.ReqID(msgCtx.Headers.GetRequestID()),
		zfield.MsgID(msgCtx.Headers.GetMessageID()),
		zfield.Sender(msgCtx.Headers.GetSender()),
		zfield.Receiver(msgCtx.Headers.GetReceiver()))

	m.msgCh <- msgCtx
	return nil
}

// Resource return resource manager.
func (m *Manager) Resource() statem.ResourceManager {
	return m.resourceManager
}

func (m *Manager) getStateMachine(cid, eid string) (string, statem.StateMachiner) {
	if cid == "" {
		cid = "default"
	}

	if container, ok := m.containers[cid]; ok {
		if sm := container.Get(eid); nil != sm {
			if sm.GetStatus() == statem.SMStatusDeleted {
				container.Remove(eid)
				return cid, nil
			}
			return cid, sm
		}
	}

	for channelID, container := range m.containers {
		if sm := container.Get(eid); sm != nil {
			if sm.GetStatus() == statem.SMStatusDeleted {
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
			var stateMachine statem.StateMachiner
			base := &dao.Entity{ID: stateID, Type: SMTypeBasic}
			if _, stateMachine = m.getStateMachine("", stateID); nil != stateMachine {
				log.Warn("load state machine", zfield.ID(stateID))
			} else if stateMachine, err = m.loadOrCreate(m.ctx, "", false, base); nil == err {
				continue
			}
			actorEnv := m.actorEnv.GetActorEnv(stateID)
			stateMachine.WithContext(statem.NewContext(stateMachine, actorEnv.Mappers, actorEnv.Tentacles))
		}
	}
	return nil
}

func (m *Manager) loadActor(ctx context.Context, typ string, id string) error {
	_, err := m.loadOrCreate(ctx, "", false, &dao.Entity{ID: id, Type: typ})
	return errors.Wrap(err, "load entity")
}

func (m *Manager) loadOrCreate(ctx context.Context, channelID string, flagCreate bool, base *dao.Entity) (sm statem.StateMachiner, err error) {
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
		if sm, err = statem.NewState(ctx, m, base, nil); nil != err {
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
	sm.WithContext(statem.NewContext(sm, thisActorEnv.Mappers, thisActorEnv.Tentacles))

	m.containers[channelID].Add(sm)
	return sm, nil
}

// Tools.

func (m *Manager) EscapedEntities(expression string) []string {
	return []string{expression}
}
