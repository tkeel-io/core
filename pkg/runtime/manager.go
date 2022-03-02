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
	"runtime"
	"sync"

	"github.com/pkg/errors"
	"github.com/tkeel-io/core/pkg/config"
	"github.com/tkeel-io/core/pkg/dispatch"
	xerrors "github.com/tkeel-io/core/pkg/errors"
	"github.com/tkeel-io/core/pkg/inbox"
	zfield "github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/core/pkg/placement"
	"github.com/tkeel-io/core/pkg/runtime/environment"
	"github.com/tkeel-io/core/pkg/runtime/message"
	"github.com/tkeel-io/core/pkg/runtime/state"
	"github.com/tkeel-io/core/pkg/types"
	"github.com/tkeel-io/kit/log"
	"go.uber.org/zap"
)

type Manager struct {
	containers      map[string]*Container
	actorEnv        environment.IEnvironment
	resourceManager types.ResourceManager
	dispatcher      dispatch.Dispatcher
	inboxes         map[string]inbox.Inboxer
	processing      sync.Map

	shutdown chan struct{}
	lock     sync.RWMutex
	ctx      context.Context
	cancel   context.CancelFunc
}

func NewManager(ctx context.Context, resourceManager types.ResourceManager, dispatcher dispatch.Dispatcher) (types.Manager, error) {
	ctx, cancel := context.WithCancel(ctx)
	stateManager := &Manager{
		ctx:             ctx,
		cancel:          cancel,
		inboxes:         make(map[string]inbox.Inboxer),
		actorEnv:        environment.NewEnvironment(),
		containers:      make(map[string]*Container),
		dispatcher:      dispatcher,
		resourceManager: resourceManager,
		processing:      sync.Map{},
		lock:            sync.RWMutex{},
	}

	return stateManager, nil
}

func (m *Manager) Start() error {
	log.Info("start runtime manager")
	m.initializeMetadata()
	m.initializeSources()
	return nil
}

func (m *Manager) Shutdown() error {
	m.cancel()
	m.shutdown <- struct{}{}
	return nil
}

func (m *Manager) selectContainer(id string) *Container {
	if _, ok := m.containers[id]; !ok {
		m.containers[id] = NewContainer(m.ctx, id, m)
	}

	return m.containers[id]
}

func (m *Manager) HandleMessage(ctx context.Context, msgCtx message.Context) error {
	reqID := msgCtx.Get(message.ExtAPIRequestID)
	entityID := msgCtx.Get(message.ExtEntityID)
	msgSender := msgCtx.Get(message.ExtMessageSender)
	channelID, _ := ctx.Value(inbox.IDKey{}).(string)
	log.Debug("dispose message", zfield.ReqID(reqID),
		zfield.Header(msgCtx.Attributes()), zfield.ID(entityID),
		zfield.Message(string(msgCtx.Message())), zfield.Channel(channelID))

	_, loaded := m.processing.Load(entityID)
	for ; loaded; _, loaded = m.processing.LoadOrStore(entityID, struct{}{}) {
		log.Debug("state processing, wait a moment", zfield.Sender(msgSender),
			zfield.ReqID(reqID), zfield.ID(entityID), zfield.Channel(channelID))
		runtime.Gosched()
	}

	// handle message.
	err := m.handleMessage(ctx, msgCtx)

	// invoke message completed.
	m.processing.Delete(entityID)
	return errors.Wrap(err, "handle message")
}

func (m *Manager) handleMessage(ctx context.Context, msgCtx message.Context) error {
	reqID := msgCtx.Get(message.ExtAPIRequestID)
	entityID := msgCtx.Get(message.ExtEntityID)
	channelID, _ := ctx.Value(inbox.IDKey{}).(string)
	container := m.selectContainer(channelID)
	machine, err := container.Load(ctx, entityID)
	if nil != err {
		if !errors.Is(err, xerrors.ErrEntityNotFound) {
			log.Error("undefine error, load state machine", zfield.ReqID(reqID),
				zap.Error(err), zfield.ID(entityID), zfield.Channel(channelID))
			return xerrors.ErrInternal
		}

		// state machine not exists, then create.
		enDao := message.ParseEntityFrom(msgCtx)
		if machine, err = container.MakeMachine(enDao); nil != err {
			log.Error("create state machine", zfield.Channel(channelID),
				zfield.ReqID(reqID), zfield.ID(entityID), zap.Error(err))
			return xerrors.ErrInternal
		}
	}

	log.Debug("handle message", zfield.Channel(channelID), zfield.Header(msgCtx.Attributes()),
		zfield.Eid(entityID), zfield.ReqID(reqID), zfield.Message(string(msgCtx.Message())))

	result := machine.Invoke(ctx, msgCtx)
	if result.Err != nil {
		log.Error("handle message", zap.Error(err),
			zfield.ID(entityID), zfield.ReqID(reqID),
			zfield.Message(string(msgCtx.Message())),
			zfield.Channel(channelID), zfield.Header(msgCtx.Attributes()))
		return errors.Wrap(result.Err, "handle message")
	}

	log.Debug("invoke message completed",
		zfield.ID(entityID), zfield.ReqID(reqID),
		zap.String("result.status", string(result.Status)))

	// handle result.
	switch result.Status {
	case state.MCreated:
		container.Add(machine)
	case state.MDeleted:
		container.Remove(machine.GetID())
	case state.MCompleted:
	default:
		// never.
	}

	return nil
}

// Resource return resource manager.
func (m *Manager) Resource() types.ResourceManager {
	return m.resourceManager
}

func (m *Manager) loadMachine(stateID string) {
	var stateType string
	switch stateType {
	case SMTypeSubscription:
		// TODO: load subscription.
	default:
	}
}

func (m *Manager) reloadMachineEnv(stateIDs []string) {
	for _, stateID := range stateIDs {
		// load state machine.
		queue := placement.Global().Select(stateID)
		container := m.selectContainer(queue.ID)

		log.Debug("reload state machine", zfield.Eid(stateID), zfield.Queue(queue))
		if config.Get().Server.Name != queue.NodeName {
			continue
		}

		var has bool
		var err error
		var machine state.Machiner
		// load state machine from runtime.
		if machine, has = container.Get(stateID); has {
			// update state machine context.
			stateEnv := m.actorEnv.GetStateEnv(stateID)
			machine.Context().LoadEnvironments(stateEnv)
		} else if _, err = container.Load(context.Background(), stateID); nil != err {
			log.Warn("load state machine from state store",
				zfield.Reason(err.Error()), zfield.Queue(queue), zfield.Eid(stateID))
			continue
		}

		ctx := context.Background()
		msgCtx := message.New(ctx)
		msgCtx.Set(message.ExtEntityID, stateID)
		msgCtx.Set(message.ExtMessageType, string(message.MessageTypeMapperInit))

		// set channel id.
		ctx = context.WithValue(ctx, inbox.IDKey{}, queue.ID)
		m.HandleMessage(ctx, msgCtx)
	}
}

// Tools.

func (m *Manager) EscapedEntities(expression string) []string {
	return []string{expression}
}
